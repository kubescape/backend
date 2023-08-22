package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		reporter := NewBaseReportSender("https://dummyeventreceiver.com", &http.Client{}, baseReport)
		reporter.SendError(err1, true, false, nil)
		reporter.SendError(err1, false, false, nil)
		reporter.SendError(err1, false, false, nil)
		reporter.SendError(err1, true, false, nil)
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

		errChan := make(chan error)
		reporter.SendError(err1, true, true, errChan)
		e := <-errChan
		assert.Error(t, e)
		done <- 0
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

		errChan1 := make(chan error)
		err2 := fmt.Errorf("dummy error1")
		reporter.SendError(err2, false, false, errChan1)
		e = <-errChan1
		assert.NoError(t, e)
		done <- 1
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

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
		compareSnapshot(snapshotNum, t, reporter.report)

		reporter.SendAsRoutine(true, nil)
		errChan2 := make(chan error)
		reporter.SendAsRoutine(true, errChan2)
		e = <-errChan2
		assert.Error(t, e)
		done <- 2

		errChan3 := make(chan error)
		reporter.SendError(nil, false, true, errChan3)
		e = <-errChan3
		assert.NoError(t, e)
		done <- 3
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

		reporter.SendStatus("status", true, nil)
		reporter.SendStatus("status", false, nil)
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

		errChan4 := make(chan error)
		reporter.SendStatus("status", true, errChan4)
		e = <-errChan4
		assert.Error(t, e)
		done <- 4

		errChan5 := make(chan error)
		reporter.SendStatus("status", false, errChan5)
		e = <-errChan5
		assert.NoError(t, e)
		done <- 5

		reporter.SendAction("action", true, nil)
		reporter.SendAction("action", false, nil)

		errChan6 := make(chan error)
		reporter.SendAction("action", true, errChan6)
		e = <-errChan6
		assert.Error(t, e)
		done <- 6

		errChan7 := make(chan error)
		reporter.SendAction("action", false, errChan7)
		e = <-errChan7
		assert.NoError(t, e)
		done <- 7
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

		reporter.SendDetails("details", true, nil)
		reporter.SendDetails("details", false, nil)

		errChan8 := make(chan error)
		reporter.SendDetails("details", true, errChan8)
		e = <-errChan8
		assert.Error(t, e)
		done <- 8

		errChan9 := make(chan error)
		reporter.SendDetails("details", false, errChan9)
		e = <-errChan9
		assert.NoError(t, e)
		done <- 9
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

		reporter.SendWarning("warning", true, false, nil)
		reporter.SendWarning("warning", false, false, nil)
		reporter.SendWarning("warning", false, false, nil)
		reporter.SendWarning("warning", true, false, nil)
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

		errChan10 := make(chan error)
		reporter.SendWarning("warning", true, false, errChan10)
		e = <-errChan10
		assert.Error(t, e)
		done <- 10

		errChan11 := make(chan error)
		reporter.SendWarning("warning", false, false, errChan11)
		e = <-errChan11
		assert.NoError(t, e)
		done <- 11
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

		errChan12 := make(chan error)
		reporter.SendWarning("warning", false, true, errChan12)
		e = <-errChan12
		assert.NoError(t, e)
		done <- 12
		snapshotNum++
		compareSnapshot(snapshotNum, t, reporter.report)

		//finally test a caller that forgets to read the error channel
		errChan14 := make(chan error)
		timelocked := time.Now()
		reporter.SendWarning("warning", false, true, errChan12)
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
