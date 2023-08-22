package systemreports

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/francoispqt/gojay"
)

func BaseReportDiff(lhs, rhs *BaseReport) {
	if strings.Compare(lhs.JobID, rhs.JobID) != 0 {
		fmt.Printf("jobID: %v != %v\n", lhs.JobID, rhs.JobID)
	}
	if strings.Compare(lhs.Status, rhs.Status) != 0 {
		fmt.Printf("Status: %v != %v\n", lhs.Status, rhs.Status)
	}
	if strings.Compare(lhs.Reporter, rhs.Reporter) != 0 {
		fmt.Printf("Reporter: %v != %v\n", lhs.Reporter, rhs.Reporter)
	}
	if strings.Compare(lhs.Target, rhs.Target) != 0 {
		fmt.Printf("Target: %v != %v\n", lhs.Target, rhs.Target)
	}
	if strings.Compare(lhs.ActionID, rhs.ActionID) != 0 {
		fmt.Printf("ActionID: %v != %v\n", lhs.ActionID, rhs.ActionID)
	}
	if strings.Compare(lhs.ActionName, rhs.ActionName) != 0 {
		fmt.Printf("ActionName: %v != %v\n", lhs.ActionName, rhs.ActionName)
	}
	if strings.Compare(lhs.ParentAction, rhs.ParentAction) != 0 {
		fmt.Printf("%v != %v\n", lhs.ParentAction, rhs.ParentAction)
	}
	if lhs.Timestamp.Unix() != rhs.Timestamp.Unix() {
		fmt.Printf("Timestamp: %v != %v\n", lhs.Timestamp, rhs.Timestamp)
	}
	if lhs.ActionIDN != rhs.ActionIDN {
		fmt.Printf("ActionIDN: %v != %v\n", lhs.ActionIDN, rhs.ActionIDN)
	}
	if !reflect.DeepEqual(rhs.Errors, lhs.Errors) {
		fmt.Printf("Errors: %v != %v\n", lhs.Errors, rhs.Errors)
	}

}
func TestUnMarshallingSuccess(t *testing.T) {
	lhs := BaseReport{Reporter: "unit-test", Target: "unit-test-framework", JobID: "1", ActionID: "1", Status: "testing", ActionName: "Testing", ActionIDN: 1}
	rhs := &BaseReport{}
	lhs.AddError("1")
	lhs.AddError("2")
	lhs.Timestamp = time.Now()
	bolB, _ := json.Marshal(lhs)
	r := bytes.NewReader(bolB)

	er := gojay.NewDecoder(r).DecodeObject(rhs)
	if er != nil {
		t.Errorf("marshalling failed due to: %v", er.Error())
	}
	if !IsEqual(&lhs, rhs) {
		BaseReportDiff(&lhs, rhs)
		fmt.Printf("%+v\n", lhs)
		t.Errorf("%v", rhs)
	}

}

func TestUnMarshallingPartial(t *testing.T) {
	lhs := BaseReport{Reporter: "unit-test", Target: "unit-test-framework", JobID: "1", ActionID: "1", Status: "testing", ActionName: "Testing", ActionIDN: 1}
	rhs := &BaseReport{}

	lhs.Timestamp = time.Now()
	bolB, _ := json.Marshal(lhs)
	r := bytes.NewReader(bolB)

	er := gojay.NewDecoder(r).DecodeObject(rhs)
	if er != nil {
		t.Errorf("marshalling failed due to: %v", er.Error())
	}
	if !IsEqual(&lhs, rhs) {
		BaseReportDiff(&lhs, rhs)
		fmt.Printf("%+v\n", lhs)
		t.Errorf("%v", rhs)
	}

}

func TestSetAction(t *testing.T) {
	lhs := BaseReport{Reporter: "unit-test", Target: "unit-test-framework",
		JobID: "1", ActionID: "1", Status: "testing", ActionName: "Testing", ActionIDN: 1}
	newAct := "blabla"
	lhs.SetActionName(newAct)

	if lhs.ActionName != newAct {
		t.Errorf("wrong action name after set: %s", lhs.ActionName)
	}
}

func TestBaseReportTestConcurrentErrorAdding(t *testing.T) {
	a := &BaseReport{Reporter: "unit-test", Target: "unit-test-framework", Status: "started", JobID: "processid1", ActionID: "1"}
	var wg sync.WaitGroup
	for j := 0; j < 10; j++ {

		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				s := strconv.Itoa(i)
				a.AddError(s)
			}(i, &wg)
		}
		wg.Wait()

		if len(a.Errors) != 4 {
			t.Errorf("an inconsistency error occurred at round %d, expected 4 errors and got %v", j, a)
		}
		a.Errors = nil

	}
}
