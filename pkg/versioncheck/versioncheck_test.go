package versioncheck

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"
)

func TestCheckLatestVersion_Semver_Compare(t *testing.T) {
	assert.Equal(t, -1, semver.Compare("v2.0.150", "v2.0.151"))
	assert.Equal(t, 0, semver.Compare("v2.0.150", "v2.0.150"))
	assert.Equal(t, 1, semver.Compare("v2.0.150", "v2.0.149"))
	assert.Equal(t, -1, semver.Compare("v2.0.150", "v3.0.150"))

}

func TestNormalizeVersion(t *testing.T) {
	assert.Equal(t, "v4.0.6", normalizeVersion("4.0.6"))
	assert.Equal(t, "v3.0.15", normalizeVersion("v3.0.15"))
	assert.Equal(t, "", normalizeVersion(""))
}

func TestSemverCompare_WithNormalization(t *testing.T) {
	result := semver.Compare(
		normalizeVersion("4.0.6"),
		normalizeVersion("v3.0.15"),
	)

	assert.Equal(t, 1, result)
}

func TestSemverCompare_WithoutNormalization(t *testing.T) {
	result := semver.Compare("4.0.6", "v3.0.15")

	assert.Equal(t, -1, result)
}

func TestCheckLatestVersion(t *testing.T) {
	type args struct {
		ctx         context.Context
		versionData *VersionCheckRequest
		versionURL  string
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "Get latest version",
			args: args{
				ctx:         context.Background(),
				versionData: &VersionCheckRequest{},
				versionURL:  "https://version-check.ks-services.co",
			},
			err: nil,
		},
		{
			name: "Failed to get latest version",
			args: args{
				ctx:         context.Background(),
				versionData: &VersionCheckRequest{},
				versionURL:  "https://example.com",
			},
			err: fmt.Errorf("failed to get latest version"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VersionCheckHandler{
				versionURL: tt.args.versionURL,
			}
			err := v.CheckLatestVersion(tt.args.ctx, tt.args.versionData)

			assert.Equal(t, tt.err, err)
		})
	}
}

func TestVersionCheckHandler_getLatestVersion(t *testing.T) {
	t.Run("Get latest version", func(t *testing.T) {
		// Live canary against the deployed version-check service. A request
		// with an empty clientVersion MUST be told the latest version -- this
		// guards the ksgf1v1 regression where empty/unparseable client
		// versions silently received no clientUpdate. We assert a *valid*
		// version rather than a hardcoded one so routine kubescape releases
		// (which bump the reference) don't break this test.
		v := &VersionCheckHandler{
			versionURL: "https://version-check.ks-services.co",
		}
		got, err := v.getLatestVersion(context.Background(), &VersionCheckRequest{Client: "kubescape"})
		if err != nil {
			t.Fatalf("getLatestVersion() unexpected error = %v", err)
		}
		if got == nil {
			t.Fatal("getLatestVersion() returned a nil response")
		}
		assert.Equal(t, "kubescape", got.Client)
		assert.NotEmpty(t, got.ClientUpdate, "empty clientVersion must be told the latest version")
		assert.True(t, semver.IsValid(normalizeVersion(got.ClientUpdate)),
			"ClientUpdate %q must be a valid semantic version", got.ClientUpdate)
	})

	t.Run("Failed to get latest version", func(t *testing.T) {
		v := &VersionCheckHandler{
			versionURL: "https://example.com",
		}
		got, err := v.getLatestVersion(context.Background(), &VersionCheckRequest{})
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestGetTriggerSource(t *testing.T) {
	// Running in github actions pipeline
	os.Setenv("GITHUB_ACTIONS", "true")
	source := getTriggerSource()
	assert.Equal(t, "pipeline", source)

	os.Args[0] = "ksserver"
	source = getTriggerSource()
	assert.Equal(t, "microservice", source)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// closeRecorder is a response body that remembers whether it was closed.
type closeRecorder struct {
	io.Reader
	closed atomic.Bool
}

func (b *closeRecorder) Close() error {
	b.closed.Store(true)
	return nil
}

// withTransport sends version-check requests to rt, so these tests never reach the real service.
func withTransport(t *testing.T, rt http.RoundTripper) {
	t.Helper()
	orig := versionCheckClient
	versionCheckClient = &http.Client{Transport: rt}
	t.Cleanup(func() { versionCheckClient = orig })
}

func TestVersionCheckHandler_fakeTransport(t *testing.T) {
	v := &VersionCheckHandler{versionURL: "https://version-check.invalid"}

	t.Run("should close the response body", func(t *testing.T) {
		body := &closeRecorder{Reader: strings.NewReader(`{"client":"kubescape","clientUpdate":"v3.0.0"}`)}
		withTransport(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header), Request: req}, nil
		}))

		got, err := v.getLatestVersion(context.Background(), &VersionCheckRequest{})
		require.NoError(t, err)
		assert.Equal(t, "v3.0.0", got.ClientUpdate)
		assert.True(t, body.closed.Load(), "response body was not closed")
	})

	t.Run("should stop at the caller deadline when the server never answers", func(t *testing.T) {
		ctxErr := make(chan error, 1)
		withTransport(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			select {
			case <-req.Context().Done():
				ctxErr <- req.Context().Err()
				return nil, req.Context().Err()
			case <-time.After(5 * time.Second):
				ctxErr <- nil
				return nil, errors.New("request context never ended")
			}
		}))

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		require.Error(t, v.CheckLatestVersion(ctx, &VersionCheckRequest{}))
		require.ErrorIs(t, <-ctxErr, context.DeadlineExceeded)
	})

	t.Run("should return the transport error", func(t *testing.T) {
		errDial := errors.New("dial failed")
		withTransport(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errDial
		}))

		got, err := v.getLatestVersion(context.Background(), &VersionCheckRequest{})
		require.ErrorIs(t, err, errDial)
		assert.Nil(t, got)
	})
}
