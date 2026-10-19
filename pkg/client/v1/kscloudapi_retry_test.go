package v1

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/kubescape/backend/pkg/utils"
	"github.com/stretchr/testify/require"
)

// sequenceServer answers each POST with the next status of its script (the last one repeats) and records the bodies.
type sequenceServer struct {
	mu       sync.Mutex
	statuses []int
	bodies   []string
}

func (s *sequenceServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	s.mu.Lock()
	idx := min(len(s.bodies), len(s.statuses)-1)
	status := s.statuses[idx]
	s.bodies = append(s.bodies, string(body))
	s.mu.Unlock()

	w.WriteHeader(status)
	if status < 400 {
		_, _ = w.Write([]byte("ok"))
		return
	}
	_, _ = w.Write([]byte("error code: " + http.StatusText(status)))
}

func (s *sequenceServer) calls() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.bodies...)
}

func newRetryTestAPI(t *testing.T, reportURL string, opts ...KSCloudOption) *KSCloudAPI {
	t.Helper()

	ks, err := NewKSCloudAPI("https://api.armo.cloud", reportURL, "account", "", append(testOptions, opts...)...)
	require.NoError(t, err)

	return ks
}

func TestSubmitReportRetry(t *testing.T) {
	const (
		cluster  = "special-cluster"
		reportID = "5d817063-096f-4d91-b39b-8665240080af"
	)

	tests := []struct {
		name          string
		statuses      []int
		attempts      int
		wantCalls     int
		wantErr       bool
		wantErrStatus int
	}{
		{name: "success on the first attempt", statuses: []int{200}, attempts: 3, wantCalls: 1},
		{name: "cloudflare 520 then success", statuses: []int{520, 200}, attempts: 3, wantCalls: 2},
		{name: "502 then 503 then success", statuses: []int{502, 503, 200}, attempts: 3, wantCalls: 3},
		{name: "504 then success", statuses: []int{504, 200}, attempts: 3, wantCalls: 2},
		{name: "524 then success", statuses: []int{524, 200}, attempts: 3, wantCalls: 2},
		{name: "persistent 520 gives up after the attempts", statuses: []int{520}, attempts: 3, wantCalls: 3, wantErr: true, wantErrStatus: 520},
		{name: "400 is not retried", statuses: []int{400, 200}, attempts: 3, wantCalls: 1, wantErr: true, wantErrStatus: 400},
		{name: "402 is not retried", statuses: []int{402, 200}, attempts: 3, wantCalls: 1, wantErr: true, wantErrStatus: 402},
		{name: "429 is not retried", statuses: []int{429, 200}, attempts: 3, wantCalls: 1, wantErr: true, wantErrStatus: 429},
		{name: "500 is not retried", statuses: []int{500, 200}, attempts: 3, wantCalls: 1, wantErr: true, wantErrStatus: 500},
		{name: "retry disabled", statuses: []int{520, 200}, attempts: 1, wantCalls: 1, wantErr: true, wantErrStatus: 520},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := &sequenceServer{statuses: tc.statuses}
			ts := httptest.NewServer(srv)
			defer ts.Close()

			ks := newRetryTestAPI(t, ts.URL, WithReportUploadRetry(tc.attempts, time.Millisecond))

			body, err := ks.SubmitReport(mockPostureReport(t, reportID, cluster))

			calls := srv.calls()
			require.Len(t, calls, tc.wantCalls)
			for _, c := range calls[1:] {
				require.Equal(t, calls[0], c, "every attempt must send the full report body")
			}
			require.NotEmpty(t, calls[0])

			if !tc.wantErr {
				require.NoError(t, err)
				require.Equal(t, "ok", body)
				return
			}

			require.Error(t, err)
			var apiErr *utils.APIError
			require.True(t, errors.As(err, &apiErr), "the error must expose the HTTP status")
			require.Equal(t, tc.wantErrStatus, apiErr.StatusCode)
		})
	}
}

func TestSubmitReportRetryTransportError(t *testing.T) {
	// a server that is closed refuses the connection: a transport error, retried until the attempts run out
	ts := httptest.NewServer(http.NotFoundHandler())
	reportURL := ts.URL
	ts.Close()

	rt := &countingRoundTripper{next: http.DefaultTransport}
	ks := newRetryTestAPI(t, reportURL,
		WithReportUploadRetry(3, time.Millisecond),
		WithHTTPClient(&http.Client{Transport: rt}),
	)

	_, err := ks.SubmitReport(mockPostureReport(t, "5d817063-096f-4d91-b39b-8665240080af", "special-cluster"))
	require.Error(t, err)
	require.Equal(t, 3, rt.count())
}

func TestSubmitReportRetryDroppedConnection(t *testing.T) {
	// the server reads the request and closes the connection without answering, the way an origin does when the
	// edge reports a 520: retried, and the third attempt gets an answer
	var (
		mu    sync.Mutex
		calls int
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n < 3 {
			conn, _, err := w.(http.Hijacker).Hijack()
			require.NoError(t, err)
			_ = conn.Close()
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	ks := newRetryTestAPI(t, ts.URL, WithReportUploadRetry(3, time.Millisecond))

	body, err := ks.SubmitReport(mockPostureReport(t, "5d817063-096f-4d91-b39b-8665240080af", "special-cluster"))
	require.NoError(t, err)
	require.Equal(t, "ok", body)
	mu.Lock()
	require.Equal(t, 3, calls)
	mu.Unlock()
}

func TestSubmitReportNoRetryWithoutReportURL(t *testing.T) {
	// no report URL configured: the request can never be sent, so it must fail at once instead of backing off
	rt := &countingRoundTripper{next: http.DefaultTransport}
	ks := NewEmptyKSCloudAPI(append(testOptions,
		WithReportUploadRetry(3, time.Hour),
		WithHTTPClient(&http.Client{Transport: rt}),
	)...)

	start := time.Now()
	_, err := ks.SubmitReport(mockPostureReport(t, "5d817063-096f-4d91-b39b-8665240080af", "special-cluster"))
	require.Error(t, err)
	require.Less(t, time.Since(start), 5*time.Second)
	require.LessOrEqual(t, rt.count(), 1)
}

func TestSubmitReportRetryHonorsContext(t *testing.T) {
	srv := &sequenceServer{statuses: []int{520}}
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// a long backoff: cancelling the context must end the wait instead of sending again
	ks := newRetryTestAPI(t, ts.URL, WithReportUploadRetry(3, time.Hour))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := ks.SubmitReportWithContext(ctx, mockPostureReport(t, "5d817063-096f-4d91-b39b-8665240080af", "special-cluster"))
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	var apiErr *utils.APIError
	require.True(t, errors.As(err, &apiErr), "the upload error must be kept next to the context error")
	require.Equal(t, 520, apiErr.StatusCode)

	require.Len(t, srv.calls(), 1)
	require.Less(t, time.Since(start), 5*time.Second)
}

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		base     time.Duration
		attempt  int
		min, max time.Duration
	}{
		{base: time.Second, attempt: 1, min: time.Second, max: 1500 * time.Millisecond},
		{base: time.Second, attempt: 2, min: 2 * time.Second, max: 3 * time.Second},
		{base: time.Second, attempt: 3, min: 4 * time.Second, max: 6 * time.Second},
		// 8s plus jitter would reach 12s: capped
		{base: time.Second, attempt: 4, min: 8 * time.Second, max: maxReportUploadRetryDelay},
		{base: time.Second, attempt: 5, min: maxReportUploadRetryDelay, max: maxReportUploadRetryDelay},
		// a huge attempt number must not overflow the shift into a small or negative delay
		{base: time.Second, attempt: 1000, min: maxReportUploadRetryDelay, max: maxReportUploadRetryDelay},
		// a base above the cap is capped too
		{base: time.Minute, attempt: 1, min: maxReportUploadRetryDelay, max: maxReportUploadRetryDelay},
	}
	for _, tc := range tests {
		for range 50 { // the jitter is random: sample it
			got := retryDelay(tc.base, tc.attempt)
			require.GreaterOrEqual(t, got, tc.min, "base %s attempt %d", tc.base, tc.attempt)
			require.LessOrEqual(t, got, tc.max, "base %s attempt %d", tc.base, tc.attempt)
		}
	}
	require.Zero(t, retryDelay(0, 1))
}

func TestAPIErrorMessageIsUnchanged(t *testing.T) {
	err := &utils.APIError{StatusCode: 520, Status: "520 ", Reason: "error code: 520"}
	require.Equal(t, "http-error: '520 ', reason: 'error code: 520'", err.Error())
}

type countingRoundTripper struct {
	mu   sync.Mutex
	n    int
	next http.RoundTripper
}

func (c *countingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
	return c.next.RoundTrip(req)
}

func (c *countingRoundTripper) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}
