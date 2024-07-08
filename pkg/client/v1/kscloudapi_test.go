package v1

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testOptions = []KSCloudOption{
		WithTrace(os.Getenv("DEBUG_TEST") != ""),
	}
)

func TestFallBackGUID(t *testing.T) {
	t.Run("should yield a GUID even though the account ID is not set", func(t *testing.T) {
		ks := NewEmptyKSCloudAPI()
		require.NotEmpty(t, ks.getCustomerGUIDFallBack())
	})
}

func TestKSCloudAPI(t *testing.T) {
	// NOTE:
	// (i)  mock handlers do not use "require" in order to let goroutines end normally upon failure.
	// (ii) run with DEBUG_TEST=1 go test -v -run KSCloudAPI to get a trace of all HTTP traffic.

	srv := MockAPIServer(t) // assert that a token is passed as header
	t.Cleanup(srv.Close)

	ks, err := NewKSCloudAPI(
		srv.Root(),
		srv.Root(),
		"account",
		"",
		testOptions...,
	)
	require.NoError(t, err)

	t.Run("with authenticated", func(t *testing.T) {

		t.Run("should retrieve AttackTracks", func(t *testing.T) {
			t.Parallel()

			tracks, err := ks.GetAttackTracks()
			require.NoError(t, err)
			require.NotNil(t, tracks)

			expected := mockAttackTracks()

			// make sure controls don't leak
			for i := range expected {
				expected[i].Spec.Data.Controls = nil // doesn't pass the JSON marshal
				for j := range expected[i].Spec.Data.SubSteps {
					expected[i].Spec.Data.SubSteps[j].Controls = nil
				}
			}
			require.EqualValues(t, expected, tracks)
		})

		t.Run("with frameworks", func(t *testing.T) {
			t.Run("should retrieve Framework #1", func(t *testing.T) {
				t.Parallel()

				framework, err := ks.GetFramework("mock-1")
				require.NoError(t, err)
				require.NotNil(t, framework)

				mocked := mockFrameworks()
				expected := &mocked[0]
				require.EqualValues(t, expected, framework)
			})

			t.Run("should retrieve Framework #2", func(t *testing.T) {
				t.Parallel()

				framework, err := ks.GetFramework("mock-2")
				require.NoError(t, err)
				require.NotNil(t, framework)

				mocked := mockFrameworks()
				expected := &mocked[1]
				require.EqualValues(t, expected, framework)
			})

			t.Run("should retrieve native Framework", func(t *testing.T) {
				t.Parallel()

				const testFramework = "MITRE"
				expected, err := os.ReadFile(TestFrameworkFile(testFramework))
				require.NoError(t, err)

				framework, err := ks.GetFramework("miTrE")
				require.NoError(t, err)
				require.NotNil(t, framework)
				jazon, err := json.Marshal(framework)
				require.NoError(t, err)
				require.JSONEq(t, string(expected), string(jazon))
			})

			t.Run("should retrieve all Frameworks", func(t *testing.T) {
				t.Parallel()

				// NOTE: MITRE fixture is not part of the base mock

				expected := mockFrameworks()
				frameworks, err := ks.GetFrameworks()
				require.NoError(t, err)
				require.Len(t, frameworks, 3)
				require.EqualValues(t, expected, frameworks)
			})

			t.Run("should list all Frameworks", func(t *testing.T) {
				t.Parallel()

				mocks := mockFrameworks()
				expected := make([]string, 0, 3)
				for _, fw := range mocks {
					expected = append(expected, fw.Name)
				}

				frameworkNames, err := ks.ListFrameworks()
				require.NoError(t, err)
				require.Len(t, frameworkNames, 3)
				require.ElementsMatch(t, expected, frameworkNames)
			})

			t.Run("should list custom Frameworks", func(t *testing.T) {
				t.Parallel()

				mocks := mockFrameworks()
				expected := make([]string, 0, 2)
				for _, fw := range mocks[:len(mocks)-1] {
					expected = append(expected, fw.Name)
				}

				frameworkNames, err := ks.ListCustomFrameworks()
				require.NoError(t, err)
				require.Len(t, frameworkNames, 2)
				require.ElementsMatch(t, expected, frameworkNames)
			})
		})

		t.Run("with controls", func(t *testing.T) {
			t.Run("should NOT retrieve Control (not a public API)", func(t *testing.T) {
				t.Parallel()

				const id = "control-1"

				control, err := ks.GetControl(id)
				require.Error(t, err)
				require.Nil(t, control)
				require.Contains(t, err.Error(), "is not public")
			})

			t.Run("should NOT list Controls (not a public API)", func(t *testing.T) {
				t.Parallel()

				control, err := ks.ListControls()
				require.Error(t, err)
				require.Nil(t, control)
				require.Contains(t, err.Error(), "is not public")
			})
		})

		t.Run("with exceptions", func(t *testing.T) {
			t.Run("should retrieve Exceptions", func(t *testing.T) {
				t.Parallel()

				expected := mockExceptions()
				exceptions, err := ks.GetExceptions("")
				require.NoError(t, err)
				require.Len(t, exceptions, 2)
				require.EqualValues(t, expected, exceptions)
			})
		})

		t.Run("with CustomerConfig", func(t *testing.T) {
			t.Run("empty CustomerConfig", func(t *testing.T) {
				t.Parallel()

				kno, err := NewKSCloudAPI(
					srv.Root(),
					"",
					"",
					"",
				)
				require.NoError(t, err)

				account, err := kno.GetAccountConfig("")
				require.NoError(t, err)
				require.NotNil(t, account)
				require.Empty(t, *account)
			})

			t.Run("should retrieve CustomerConfig", func(t *testing.T) {
				t.Parallel()

				expected := mockCustomerConfig("", "")()
				account, err := ks.GetAccountConfig("")
				require.NoError(t, err)
				require.NotNil(t, account)
				require.EqualValues(t, expected, account)
			})

			t.Run("should retrieve CustomerConfig for cluster", func(t *testing.T) {
				t.Parallel()

				const cluster = "special-cluster"

				expected := mockCustomerConfig(cluster, "")()
				account, err := ks.GetAccountConfig(cluster)
				require.NoError(t, err)
				require.NotNil(t, account)
				require.EqualValues(t, expected, account)
			})

			t.Run("should retrieve ControlInputs", func(t *testing.T) {
				t.Parallel()

				config := mockCustomerConfig("", "")()
				expected := config.Settings.PostureControlInputs

				inputs, err := ks.GetControlsInputs("")
				require.NoError(t, err)
				require.NotNil(t, inputs)
				require.EqualValues(t, expected, inputs)
			})
		})

		t.Run("should submit report", func(t *testing.T) {
			t.Parallel()

			const (
				cluster  = "special-cluster"
				reportID = "5d817063-096f-4d91-b39b-8665240080af"
			)

			submitted := mockPostureReport(t, reportID, cluster)
			_, err := ks.SubmitReport(submitted)
			require.NoError(t, err)
		})
	})

	t.Run("with getters & setters", func(t *testing.T) {

		kno, err := NewKSCloudAPI(
			srv.Root(),
			"",
			"",
			"",
		)
		require.NoError(t, err)

		pickString := func() string {
			return strconv.Itoa(rand.Intn(10000)) //nolint:gosec
		}

		t.Run("should get&set account", func(t *testing.T) {
			str := pickString()
			kno.accountID = str
			require.Equal(t, str, kno.GetAccountID())
		})

		t.Run("shouldn't set invalid report URL", func(t *testing.T) {
			malformedUrl := "http://%41:8080/"
			err := kno.SetCloudReportURL(malformedUrl)
			require.Error(t, err)
			require.Equal(t, "", kno.GetCloudReportURL())
		})

		t.Run("shouldn't set invalid API URL", func(t *testing.T) {
			malformedUrl := "http://%41:8080/"
			err := kno.SetCloudAPIURL(malformedUrl)
			require.Error(t, err)
			require.Equal(t, "", kno.GetCloudAPIURL())
		})

		t.Run("should get&set report URL", func(t *testing.T) {
			str := "https://report.example.com"
			err := kno.SetCloudReportURL(str)
			require.NoError(t, err)
			require.Equal(t, str, kno.GetCloudReportURL())
		})

		t.Run("should get&set API URL", func(t *testing.T) {
			str := "https://api.example.com"
			err := kno.SetCloudAPIURL(str)
			require.NoError(t, err)
			require.Equal(t, str, kno.GetCloudAPIURL())
		})
	})

	t.Run("with API errors", func(t *testing.T) {
		// exercise the client when the API returns errors
		t.Parallel()

		errAPI := errors.New("test error")
		errSrv := MockAPIServer(t, withAPIError(errAPI))
		t.Cleanup(errSrv.Close)

		ke, err := NewKSCloudAPI(
			errSrv.Root(),
			"",
			"account",
			"",
		)
		require.NoError(t, err)

		t.Run("API calls should error", func(t *testing.T) {
			_, err = ke.GetExceptions("")
			require.Error(t, err)
			require.Contains(t, err.Error(), errAPI.Error())

			_, err = ke.GetControlsInputs("")
			require.Error(t, err)
			require.Contains(t, err.Error(), errAPI.Error())

			_, err = ke.GetAccountConfig("")
			require.Error(t, err)
			require.Contains(t, err.Error(), errAPI.Error())

			_, err = ke.GetAttackTracks()
			require.Error(t, err)
			require.Contains(t, err.Error(), errAPI.Error())

			_, err = ke.GetFramework("mock-1")
			require.Error(t, err)
			require.Contains(t, err.Error(), errAPI.Error())

			_, err = ke.GetFrameworks()
			require.Error(t, err)
			require.Contains(t, err.Error(), errAPI.Error())

			_, err = ke.ListFrameworks()
			require.Error(t, err)
			require.Contains(t, err.Error(), errAPI.Error())

			_, err = ke.ListCustomFrameworks()
			require.Error(t, err)
			require.Contains(t, err.Error(), errAPI.Error())
		})
	})

	t.Run("with API returning invalid response", func(t *testing.T) {
		// exercise the client when the API returns an invalid response
		t.Parallel()

		errSrv := MockAPIServer(t, withAPIGarbled(true))
		t.Cleanup(errSrv.Close)

		ke, err := NewKSCloudAPI(
			errSrv.Root(),
			"",
			"account",
			"",
		)
		require.NoError(t, err)

		t.Run("API calls should return unmarshalling error", func(t *testing.T) {
			// only API calls that return a typed response are checked

			_, err := ke.GetExceptions("")
			require.Error(t, err)

			_, err = ke.GetAccountConfig("")
			require.Error(t, err)

			_, err = ke.GetControlsInputs("")
			require.Error(t, err)

			_, err = ke.GetAttackTracks()
			require.Error(t, err)

			_, err = ke.GetFramework("mock-1")
			require.Error(t, err)

			_, err = ke.GetFrameworks()
			require.Error(t, err)

			_, err = ke.ListFrameworks()
			require.Error(t, err)

			_, err = ke.ListCustomFrameworks()
			require.Error(t, err)
		})
	})
}

func withAPIError(err error) mockAPIOption {
	return func(o *mockAPIOptions) {
		o.withError = err
	}
}

func withAPIGarbled(enabled bool) mockAPIOption {
	return func(o *mockAPIOptions) {
		o.withGarbled = enabled
	}
}

func TestGetExceptionsURL(t *testing.T) {
	ks, err := NewKSCloudAPI("https://api.kubescape.com", "https://api.google.com/report", "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		clusterName string
		expectedURL string
	}{
		{
			name:        "should return correct URL with given cluster name",
			clusterName: "testCluster",
			expectedURL: "https://api.kubescape.com/api/v1/controlExceptions?customerGUID=00000000-0000-0000-0000-000000000000&gitRegoStoreVersion=v2",
		},
		{
			name:        "should return correct URL with different cluster name",
			clusterName: "anotherTestCluster",
			expectedURL: "https://api.kubescape.com/api/v1/controlExceptions?customerGUID=00000000-0000-0000-0000-000000000000&gitRegoStoreVersion=v2",
		},
		{
			name:        "should return correct URL when cluster name is empty",
			clusterName: "",
			expectedURL: "https://api.kubescape.com/api/v1/controlExceptions?customerGUID=00000000-0000-0000-0000-000000000000&gitRegoStoreVersion=v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultURL := ks.getExceptionsURL(tt.clusterName)
			require.Equal(t, tt.expectedURL, resultURL)
		})
	}
}
