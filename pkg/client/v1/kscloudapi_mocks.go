package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/armoapi-go/identifiers"
	jsoniter "github.com/json-iterator/go"

	backendServer "github.com/kubescape/backend/pkg/server/v1"
	"github.com/kubescape/backend/pkg/utils"
	"github.com/kubescape/opa-utils/reporthandling"
	"github.com/kubescape/opa-utils/reporthandling/attacktrack/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// extra mock API routes

	pathTestPost   = "/test-post"
	pathTestDelete = "/test-delete"
	pathTestGet    = "/test-get"
)

type (
	testServer struct {
		*httptest.Server
		*mockAPIOptions
	}

	mockAPIOption  func(*mockAPIOptions)
	mockAPIOptions struct {
		withError   error // responds error systematically
		withGarbled bool  // responds garbled JSON (if a JSON response is expected)
		withAuth    bool  // asserts a token in headers
	}
)

func MockAPIServer(t testing.TB, opts ...mockAPIOption) *testServer {
	h := http.NewServeMux()

	// test options: regular mock (default), error or garbled JSON output
	server := &testServer{
		Server:         httptest.NewServer(h),
		mockAPIOptions: apiOptions(opts),
	}

	h.HandleFunc(pathTestPost, func(w http.ResponseWriter, r *http.Request) {
		if !isPost(t, r) {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		if !server.AssertAuth(t, r) {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if server.WantsError(w) {
			return
		}

		if server.WantsGarbled(w) {
			return
		}

		echoRequest(w, r)
	})

	h.HandleFunc(pathTestDelete, func(w http.ResponseWriter, r *http.Request) {
		if !isDelete(t, r) {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		if !server.AssertAuth(t, r) {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if server.WantsError(w) {
			return
		}

		if server.WantsGarbled(w) {
			return
		}

		echoHeaders(w, r)
		fmt.Fprintf(w, "body-delete")
	})

	h.HandleFunc(pathTestGet, func(w http.ResponseWriter, r *http.Request) {
		if !isGet(t, r) {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		if !server.AssertAuth(t, r) {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if server.WantsError(w) {
			return
		}

		if server.WantsGarbled(w) {
			return
		}

		echoHeaders(w, r)
		fmt.Fprintf(w, "body-get")
	})

	h.HandleFunc(backendServer.ApiServerAttackTracksPath, mockHandlerAttackTracks(t, opts...))
	h.HandleFunc(backendServer.ApiServerFrameworksPath, mockHandlerFrameworks(t, opts...))
	h.HandleFunc(backendServer.ApiServerExceptionsPath, mockHandlerExceptions(t, opts...))
	h.HandleFunc(backendServer.ApiServerCustomerConfigPath, mockHandlerCustomerConfiguration(t, opts...))
	h.HandleFunc(backendServer.ReporterReportPath, mockHandlerReport(t, opts...))

	return server
}

func mockAttackTracks() []v1alpha1.AttackTrack {
	return []v1alpha1.AttackTrack{
		{
			ApiVersion: "v1",
			Kind:       "track",
			Metadata:   map[string]interface{}{"label": "name"},
			Spec: v1alpha1.AttackTrackSpecification{
				Version:     "v2",
				Description: "a mock",
				Data: v1alpha1.AttackTrackStep{
					Name:        "track1",
					Description: "mock-step",
					SubSteps: []v1alpha1.AttackTrackStep{
						{
							Name:        "track1",
							Description: "mock-step",
							Controls: []v1alpha1.IAttackTrackControl{
								mockControlPtr("control-1"),
							},
						},
					},
					Controls: []v1alpha1.IAttackTrackControl{
						mockControlPtr("control-2"),
						mockControlPtr("control-3"),
					},
				},
			},
		},
		{
			ApiVersion: "v1",
			Kind:       "track",
			Metadata:   map[string]interface{}{"label": "stuff"},
			Spec: v1alpha1.AttackTrackSpecification{
				Version:     "v1",
				Description: "another mock",
				Data: v1alpha1.AttackTrackStep{
					Name:        "track2",
					Description: "mock-step2",
					SubSteps: []v1alpha1.AttackTrackStep{
						{
							Name:        "track3",
							Description: "mock-step",
							Controls: []v1alpha1.IAttackTrackControl{
								mockControlPtr("control-4"),
							},
						},
					},
					Controls: []v1alpha1.IAttackTrackControl{
						mockControlPtr("control-5"),
						mockControlPtr("control-6"),
					},
				},
			},
		},
	}
}

func mockFrameworks() []reporthandling.Framework {
	id1s := []string{"control-1", "control-2"}
	id2s := []string{"control-3", "control-4"}
	id3s := []string{"control-5", "control-6"}

	return []reporthandling.Framework{
		{
			PortalBase: armotypes.PortalBase{
				Name: "mock-1",
			},
			CreationTime: "now",
			Description:  "mock-1",
			Controls: []reporthandling.Control{
				mockControl("control-1"),
				mockControl("control-2"),
			},
			ControlsIDs: &id1s,
			SubSections: map[string]*reporthandling.FrameworkSubSection{
				"section1": {
					ID:         "section-id",
					ControlIDs: id1s,
				},
			},
		},
		{
			PortalBase: armotypes.PortalBase{
				Name: "mock-2",
			},
			CreationTime: "then",
			Description:  "mock-2",
			Controls: []reporthandling.Control{
				mockControl("control-3"),
				mockControl("control-4"),
			},
			ControlsIDs: &id2s,
			SubSections: map[string]*reporthandling.FrameworkSubSection{
				"section2": {
					ID:         "section-id",
					ControlIDs: id2s,
				},
			},
		},
		{
			PortalBase: armotypes.PortalBase{
				Name: "nsa",
			},
			CreationTime: "tomorrow",
			Description:  "nsa mock",
			Controls: []reporthandling.Control{
				mockControl("control-5"),
				mockControl("control-6"),
			},
			ControlsIDs: &id3s,
			SubSections: map[string]*reporthandling.FrameworkSubSection{
				"section2": {
					ID:         "section-id",
					ControlIDs: id3s,
				},
			},
		},
	}
}

func mockControl(controlID string) reporthandling.Control {
	return reporthandling.Control{
		ControlID: controlID,
	}
}
func mockControlPtr(controlID string) *reporthandling.Control {
	val := mockControl(controlID)

	return &val
}

func mockExceptions() []armotypes.PostureExceptionPolicy {
	return []armotypes.PostureExceptionPolicy{
		{
			PolicyType:   "postureExceptionPolicy",
			CreationTime: "now",
			Actions: []armotypes.PostureExceptionPolicyActions{
				"alertOnly",
			},
			Resources: []identifiers.PortalDesignator{
				{
					DesignatorType: "Attributes",
					Attributes: map[string]string{
						"kind":      "Pod",
						"name":      "coredns-[A-Za-z0-9]+-[A-Za-z0-9]+",
						"namespace": "kube-system",
					},
				},
				{
					DesignatorType: "Attributes",
					Attributes: map[string]string{
						"kind":      "Pod",
						"name":      "etcd-.*",
						"namespace": "kube-system",
					},
				},
			},
			PosturePolicies: []armotypes.PosturePolicy{
				{
					FrameworkName: "MITRE",
					ControlID:     "C-.*",
				},
				{
					FrameworkName: "another-framework",
					ControlID:     "a regexp",
				},
			},
		},
		{
			PolicyType:   "postureExceptionPolicy",
			CreationTime: "then",
			Actions: []armotypes.PostureExceptionPolicyActions{
				"alertOnly",
			},
			Resources: []identifiers.PortalDesignator{
				{
					DesignatorType: "Attributes",
					Attributes: map[string]string{
						"kind": "Deployment",
						"name": "my-regexp",
					},
				},
				{
					DesignatorType: "Attributes",
					Attributes: map[string]string{
						"kind": "Secret",
						"name": "another-regexp",
					},
				},
			},
			PosturePolicies: []armotypes.PosturePolicy{
				{
					FrameworkName: "yet-another-framework",
					ControlID:     "a regexp",
				},
			},
		},
	}
}

func mockCustomerConfig(cluster, scope string) func() *armotypes.CustomerConfig {
	if cluster == "" {
		cluster = "my-cluster"
	}

	if scope == "" {
		scope = "default"
	}

	return func() *armotypes.CustomerConfig {
		return &armotypes.CustomerConfig{
			Name: "user",
			Attributes: map[string]interface{}{
				"label": "value",
			},
			Scope: identifiers.PortalDesignator{
				DesignatorType: "Attributes",
				Attributes: map[string]string{
					"kind":  "Cluster",
					"name":  cluster,
					"scope": scope,
				},
			},
			Settings: armotypes.Settings{
				PostureControlInputs: map[string][]string{
					"inputs-1": {"x1", "y2"},
					"inputs-2": {"x2", "y2"},
				},
				PostureScanConfig: armotypes.PostureScanConfig{
					ScanFrequency: armotypes.ScanFrequency("weekly"),
				},
				VulnerabilityScanConfig: armotypes.VulnerabilityScanConfig{
					ScanFrequency:             armotypes.ScanFrequency("daily"),
					CriticalPriorityThreshold: 1,
					HighPriorityThreshold:     2,
					MediumPriorityThreshold:   3,
					ScanNewDeployment:         true,
					AllowlistRegistries:       []string{"a", "b"},
					BlocklistRegistries:       []string{"c", "d"},
				},
				SlackConfigurations: armotypes.SlackSettings{
					Token: "slack-token",
				},
			},
		}
	}
}

func mockPostureReport(t testing.TB, reportID, cluster string) *PostureReport {
	fixture := filepath.Join(utils.CurrentDir(), "testdata", "mock_posture_report.json")

	buf, err := os.ReadFile(fixture)
	require.NoError(t, err)

	var report PostureReport
	require.NoError(t,
		jsoniter.Unmarshal(buf, &report),
	)

	return &report
}

func apiOptions(opts []mockAPIOption) *mockAPIOptions {
	o := &mockAPIOptions{}
	for _, apply := range opts {
		apply(o)
	}

	return o
}

func mockHandlerAttackTracks(t testing.TB, opts ...mockAPIOption) func(http.ResponseWriter, *http.Request) {
	return mockHandlerGetWithGUID(t, mockAttackTracks, opts...)
}

func mockHandlerExceptions(t testing.TB, opts ...mockAPIOption) func(http.ResponseWriter, *http.Request) {
	return mockHandlerGetWithGUID(t, mockExceptions, opts...)
}

func mockHandlerCustomerConfiguration(t testing.TB, opts ...mockAPIOption) func(http.ResponseWriter, *http.Request) {
	o := apiOptions(opts)

	return func(w http.ResponseWriter, r *http.Request) {
		if !assert.NoErrorf(t, r.ParseForm(), "expected params to parse") {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if !o.AssertAuth(t, r) {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if o.WantsError(w) {
			return
		}

		if o.WantsGarbled(w) {
			return
		}

		cluster := r.Form.Get("clusterName")
		scope := r.Form.Get("scope")

		mockHandlerGetWithGUID(t, mockCustomerConfig(cluster, scope), opts...)(w, r)
	}
}

func mockHandlerReport(t testing.TB, opts ...mockAPIOption) func(http.ResponseWriter, *http.Request) {
	o := apiOptions(opts)

	return func(w http.ResponseWriter, r *http.Request) {
		if !isPost(t, r) {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		if !assert.NoErrorf(t, r.ParseForm(), "expected params to parse") {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if !o.AssertAuth(t, r) {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if !assert.NotEmpty(t, r.Form) {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if o.WantsError(w) {
			return
		}

		if o.WantsGarbled(w) {
			return
		}

		if !isJSON(t, r) {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if name := r.Form.Get("contextName"); name == "" {
			w.WriteHeader(http.StatusBadRequest)
		}

		if name := r.Form.Get("clusterName"); name == "" {
			w.WriteHeader(http.StatusBadRequest)
		}

		if name := r.Form.Get("reportGUID"); name == "" {
			w.WriteHeader(http.StatusBadRequest)
		}

		buf, err := io.ReadAll(r.Body)
		defer func() {
			_ = r.Body.Close()
		}()

		if !assert.NoError(t, err) {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		var payload PostureReport
		if !assert.NoErrorf(t, json.Unmarshal(buf, &payload), "expected payload to unmarshal into PostureReport, but got: %q", string(buf)) {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func mockHandlerGetWithGUID[T any](t testing.TB, generator func() T, opts ...mockAPIOption) func(http.ResponseWriter, *http.Request) {
	o := apiOptions(opts)

	return func(w http.ResponseWriter, r *http.Request) {
		if !isGet(t, r) {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		if !o.AssertAuth(t, r) {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if !hasGUID(t, r) {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if o.WantsError(w) {
			return
		}

		if o.WantsGarbled(w) {
			return
		}

		enc := json.NewEncoder(w)
		var doc T
		assert.NoErrorf(t, enc.Encode(generator()), "expected %T fixture to marshal to JSON", doc)
	}
}

func mockHandlerFrameworks(t testing.TB, opts ...mockAPIOption) func(http.ResponseWriter, *http.Request) {
	o := apiOptions(opts)

	return func(w http.ResponseWriter, r *http.Request) {
		if !isGet(t, r) {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		if !o.AssertAuth(t, r) {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if !hasGUID(t, r) {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if o.WantsError(w) {
			return
		}

		if o.WantsGarbled(w) {
			return
		}

		frameworks := mockFrameworks()
		name := r.Form.Get("frameworkName")
		if name == "" {
			enc := json.NewEncoder(w)
			assert.NoErrorf(t, enc.Encode(frameworks), "expected Framework fixture to marshal to JSON")

			return
		}

		assert.Contains(t, []string{"mock-1", "mock-2", "MITRE"}, name)

		var framework Framework
		switch name {
		case "mock-1":
			framework = frameworks[0]
		case "mock-2":
			framework = frameworks[1]
		case "MITRE":
			// load MITRE from JSON fixture
			const testFramework = "MITRE"
			buf, err := os.ReadFile(TestFrameworkFile(testFramework))
			if !assert.NoError(t, err) {
				w.WriteHeader(http.StatusInternalServerError)

				return
			}
			_, _ = w.Write(buf)
		}

		enc := json.NewEncoder(w)
		assert.NoErrorf(t, enc.Encode(framework), "expected Framework fixture to marshal to JSON")
	}
}

func isPost(t testing.TB, r *http.Request) bool {
	return assert.Truef(t, strings.EqualFold(http.MethodPost, r.Method), "expected a POST method called, but got %q", r.Method)
}

func isDelete(t testing.TB, r *http.Request) bool {
	return assert.Truef(t, strings.EqualFold(http.MethodDelete, r.Method), "expected a DELETE method called, but got %q", r.Method)
}

func isGet(t testing.TB, r *http.Request) bool {
	return assert.Truef(t, strings.EqualFold(http.MethodGet, r.Method), "expected a GET method called, but got %q", r.Method)
}

func isJSON(t testing.TB, r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")

	return assert.Equalf(t, "application/json", contentType, "expected application/json content type")
}

func (s *testServer) Root() string {
	return s.Server.URL
}

func (s *testServer) URL(pth string) string {
	pth = strings.TrimLeft(pth, "/")

	return fmt.Sprintf("%s/%s", s.Server.URL, pth)
}

// WantsError responds with the configured error.
func (o *mockAPIOptions) WantsError(w http.ResponseWriter) bool {
	if o.withError == nil {
		return false
	}

	http.Error(w, o.withError.Error(), http.StatusInternalServerError)

	return true
}

// WantsGarbled responds with invalid JSON
func (o *mockAPIOptions) WantsGarbled(w http.ResponseWriter) bool {
	if !o.withGarbled {
		return false
	}

	invalidJSON(w)

	return true
}

// AssertAuth asserts the presence of an Authorization Bearer token.
func (o *mockAPIOptions) AssertAuth(t testing.TB, r *http.Request) bool {
	if !o.withAuth {
		return true
	}

	header := r.Header.Get("Authorization")
	if !assert.NotEmpty(t, header) {
		return false
	}

	var token string
	_, err := fmt.Sscanf(header, "Bearer %s", &token)
	if !assert.NoError(t, err) {
		return false
	}

	return assert.NotEmpty(t, token)
}

func invalidJSON(w http.ResponseWriter) {
	fmt.Fprintf(w, `{"garbled":`)
}

func hasGUID(t testing.TB, r *http.Request) bool {
	if !assert.NoErrorf(t, r.ParseForm(), "expected params to parse") {
		return false
	}

	if !assert.NotEmpty(t, r.Form) {
		return false
	}

	if !assert.NotEmpty(t, r.Form.Get("customerGUID")) {
		return false
	}

	return true
}

func echoHeaders(w http.ResponseWriter, r *http.Request) {
	for key, vals := range r.Header {
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}
}

func echoRequest(w http.ResponseWriter, r *http.Request) {
	echoHeaders(w, r)
	echoBody(w, r)
}

func echoBody(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	_, _ = io.Copy(w, r.Body)
}

func TestFrameworkFile(framework string) string {
	return filepath.Join(utils.CurrentDir(), "testdata", fmt.Sprintf("%s.json", framework))
}
