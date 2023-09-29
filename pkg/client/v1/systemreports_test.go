package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "embed"

	"github.com/kubescape/backend/pkg/server/v1/systemreports"
	"github.com/stretchr/testify/assert"
)

func TestSendDeadlock(t *testing.T) {
	MAX_RETRIES = 2
	RETRY_DELAY = 0
	done := make(chan interface{})

	go func() {
		snapshotNum := 0
		baseReport := systemreports.NewBaseReport("a-user-guid", "my-reporter")
		baseReport.SetDetails("testing reporter")
		baseReport.SetActionName("testing action")

		err1 := fmt.Errorf("dummy error")
		reporter := &BaseReportSender{
			eventReceiverUrl: "https://dummyeventreceiver.com",
			report:           baseReport,
			headers:          map[string]string{},
			httpSender:       &HttpReportSenderMock{},
			// httpSender: &HttpReportSender{httpClient: &http.Client{}},
		}
		reporter.SendError(err1, true, false)
		reporter.SendError(err1, false, false)
		reporter.SendError(err1, false, false)
		reporter.SendError(err1, true, false)
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 1

		reporter.SendError(err1, false, true)
		done <- 0
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 2

		err2 := fmt.Errorf("dummy error1")
		reporter.SendError(err2, false, false)
		done <- 1
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 3

		reporter.report.SetJobID("job-id")
		reporter.report.SetParentAction("parent-action")
		reporter.report.SetActionName("testing action2")
		reporter.report.SetActionIDN(20)
		reporter.report.SetStatus("testing status")
		reporter.report.SetTarget("testing target")
		reporter.report.SetReporter("testing reporter v2")
		reporter.report.SetCustomerGUID("new-customer-guid")
		reporter.report.SetTimestamp(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 4

		reporter.SendAsRoutine(false)
		reporter.SendAsRoutine(true)
		done <- 2

		reporter.SendError(nil, false, true)
		done <- 3
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 5

		reporter.SendStatus("status", true)
		reporter.SendStatus("status", false)
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 6

		reporter.SendStatus("status", true)
		done <- 4

		reporter.SendStatus("status", false)
		done <- 5

		reporter.SendAction("action", true)
		reporter.SendAction("action", false)

		reporter.SendAction("action", true)
		done <- 6

		reporter.SendAction("action", true)
		done <- 7
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 7

		reporter.SendDetails("details", true)
		reporter.SendDetails("details", false)

		reporter.SendDetails("details", true)
		done <- 8

		reporter.SendDetails("details", false)
		done <- 9
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 8

		reporter.SendWarning("warning", true, false)
		reporter.SendWarning("warning", false, false)
		reporter.SendWarning("warning", false, false)
		reporter.SendWarning("warning", true, false)
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 9

		reporter.SendWarning("warning", true, false)
		done <- 10

		reporter.SendWarning("warning", false, false)
		done <- 11
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 10

		reporter.SendWarning("warning", false, true)
		done <- 12
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report) // 11

		//finally test a caller that forgets to read the error channel
		errChan14 := make(chan error)
		timelocked := time.Now()
		reporter.SendWarning("warning", false, true)
		//call a mutex blocking function
		reporter.report.SetStatus("status")
		dur := time.Now().Sub(timelocked)
		assert.True(t, dur.Milliseconds() < 500)
		//if we are here the mutex was unlocked after few retries to write to the error channel
		done <- 13
		select {
		case <-errChan14:
			t.Error("should not have received an error")
		default:
			done <- 14
		}

	}()

	expectedMsgs := 14
	for i := 0; i <= expectedMsgs+1; i++ {
		select {
		case <-time.After(10 * time.Second):
			if i <= expectedMsgs {
				t.Fatalf("Deadlock detected message %d did not arrive", i)
			}
		case id := <-done:
			if i > expectedMsgs {
				t.Errorf("unexpected message %d id:%d", i, id)
			}
		}
	}
}

//go:embed testdata/systemreports/report1_snapshot.json
var report1_snapshot []byte

//go:embed testdata/systemreports/report2_snapshot.json
var report2_snapshot []byte

//go:embed testdata/systemreports/report3_snapshot.json
var report3_snapshot []byte

//go:embed testdata/systemreports/report4_snapshot.json
var report4_snapshot []byte

//go:embed testdata/systemreports/report5_snapshot.json
var report5_snapshot []byte

//go:embed testdata/systemreports/report6_snapshot.json
var report6_snapshot []byte

//go:embed testdata/systemreports/report7_snapshot.json
var report7_snapshot []byte

//go:embed testdata/systemreports/report8_snapshot.json
var report8_snapshot []byte

//go:embed testdata/systemreports/report9_snapshot.json
var report9_snapshot []byte

//go:embed testdata/systemreports/report10_snapshot.json
var report10_snapshot []byte

//go:embed testdata/systemreports/report11_snapshot.json
var report11_snapshot []byte

func compareSnapshot(id int, t *testing.T, r *systemreports.BaseReport) {
	r.Mutex.Lock()
	rStr, _ := json.MarshalIndent(r, "", "\t")
	r.Mutex.Unlock()

	/*uncomment to update expected
	os.WriteFile(fmt.Sprintf("./fixtures/report%d_snapshot.json", id), rStr, 0666)
	return
	*/

	actual := &systemreports.BaseReport{}
	if err := json.Unmarshal(rStr, actual); err != nil {
		t.Error(fmt.Sprintf("Could not decode actual to compare with report%d_snapshot.json ", id), err)
	}

	var expectedBytes []byte
	switch id {
	case 1:
		expectedBytes = report1_snapshot
	case 2:
		expectedBytes = report2_snapshot
	case 3:
		expectedBytes = report3_snapshot
	case 4:
		expectedBytes = report4_snapshot
	case 5:
		expectedBytes = report5_snapshot
	case 6:
		expectedBytes = report6_snapshot
	case 7:
		expectedBytes = report7_snapshot
	case 8:
		expectedBytes = report8_snapshot
	case 9:
		expectedBytes = report9_snapshot
	case 10:
		expectedBytes = report10_snapshot
	case 11:
		expectedBytes = report11_snapshot
	default:
		t.Fatalf("Unknown snapshot id: %d", id)
	}
	expectedReport := &systemreports.BaseReport{}
	if err := json.Unmarshal(expectedBytes, expectedReport); err != nil {
		t.Error(fmt.Sprintf("Could not decode report%d_snapshot.json ", id), err)
	}
	expectedReport.Timestamp = actual.Timestamp

	if expectedReport.Errors == nil {
		expectedReport.Errors = make([]string, 0)
	}
	if actual.Errors == nil {
		actual.Errors = make([]string, 0)
	}
	assert.Equal(t, expectedReport, actual, "Snapshot id: %d is different than expected", id)
}

func TestSystemReportEndpointSetOrDefault(t *testing.T) {
	tt := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid value is set",
			input: "/k8s/sysreport-test",
			want:  "/k8s/sysreport-test",
		},
		{
			name:  "empty value is set to default",
			input: "",
			want:  "/k8s/sysreport",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			systemReportEndpoint.SetOrDefault(tc.input)

			got := systemReportEndpoint.Get()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSystemReportEndpointGetOrDefault(t *testing.T) {
	tt := []struct {
		name          string
		previousValue string
		want          string
		wantAfter     string
	}{
		{
			name:          "previously set value is returned",
			previousValue: "/k8s/sysreport-test",
			want:          "/k8s/sysreport-test",
			wantAfter:     "/k8s/sysreport-test",
		},
		{
			name:          "no previous value returns default",
			previousValue: "",
			want:          "/k8s/sysreport",
			wantAfter:     "/k8s/sysreport",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			systemReportEndpoint.Set(tc.previousValue)

			got := systemReportEndpoint.GetOrDefault()
			gotAfter := systemReportEndpoint.Get()
			assert.Equal(t, tc.want, got)
			assert.Equalf(t, tc.wantAfter, gotAfter, "default value has not been set after getting with default")
		})
	}
}

func TestSystemReportEndpointIsEmpty(t *testing.T) {
	tt := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "empty string as endpoint value is considered empty",
			value: "",
			want:  true,
		},
		{
			name:  "non-empty string as endpoint value is considered empty",
			value: "/k8s/sysreport",
			want:  false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			systemReportEndpoint.Set(tc.value)

			got := systemReportEndpoint.IsEmpty()

			assert.Equal(t, tc.want, got)
		})
	}
}
