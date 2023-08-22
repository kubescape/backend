package systemreports

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (report *BaseReport) NextActionID() {
	report.ActionIDN++
	report.ActionID = report.GetNextActionId()
}
func (report *BaseReport) SimpleReportAnnotations(setParent bool, setCurrent bool) (string, string) {

	nextActionID := report.GetNextActionId()

	jobs := JobsAnnotations{LastActionID: nextActionID}
	if setParent {
		jobs.ParentJobID = report.JobID
	}
	if setCurrent {
		jobs.CurrJobID = report.JobID
	}
	jsonAsString, _ := json.Marshal(jobs)
	return string(jsonAsString), nextActionID
	//ok
}

func (report *BaseReport) GetNextActionId() string {
	return strconv.Itoa(report.ActionIDN)
}

func (report *BaseReport) AddError(er string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	if report.Errors == nil {
		report.Errors = make([]string, 0)
	}
	report.Errors = append(report.Errors, er)
}

func (report *BaseReport) GetReportID() string {
	return fmt.Sprintf("%s::%s::%s (verbose:  %s::%s)", report.Target, report.JobID, report.ActionID, report.ParentAction, report.ActionName)
}

// ======================================== SEND WRAPPER =======================================

// ============================================ SET ============================================

func (report *BaseReport) SetReporter(reporter string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.doSetReporter(reporter)
}
func (report *BaseReport) doSetReporter(reporter string) {
	report.Reporter = strings.ToTitle(reporter)
}

func (report *BaseReport) SetStatus(status string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.DoSetStatus(status)
}
func (report *BaseReport) DoSetStatus(status string) {
	report.Status = status
}

func (report *BaseReport) SetActionName(actionName string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.DoSetActionName(actionName)
}
func (report *BaseReport) DoSetActionName(actionName string) {
	report.ActionName = actionName
}

func (report *BaseReport) SetDetails(details string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.DoSetDetails(details)
}
func (report *BaseReport) DoSetDetails(details string) {
	report.Details = details
}

func (report *BaseReport) SetTarget(target string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.doSetTarget(target)
}
func (report *BaseReport) doSetTarget(target string) {
	report.Target = target
}

func (report *BaseReport) SetActionID(actionID string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.doSetActionID(actionID)
}
func (report *BaseReport) doSetActionID(actionID string) {
	report.ActionID = actionID
}

func (report *BaseReport) SetJobID(jobID string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.doSetJobID(jobID)
}
func (report *BaseReport) doSetJobID(jobID string) {
	report.JobID = jobID
}

func (report *BaseReport) SetParentAction(parentAction string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.doSetParentAction(parentAction)
}
func (report *BaseReport) doSetParentAction(parentAction string) {
	report.ParentAction = parentAction
}

func (report *BaseReport) SetCustomerGUID(customerGUID string) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.doSetCustomerGUID(customerGUID)
}
func (report *BaseReport) doSetCustomerGUID(customerGUID string) {
	report.CustomerGUID = customerGUID
}

func (report *BaseReport) SetActionIDN(actionIDN int) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.doSetActionIDN(actionIDN)
}
func (report *BaseReport) doSetActionIDN(actionIDN int) {
	report.ActionIDN = actionIDN
	report.ActionID = strconv.Itoa(report.ActionIDN)
}

func (report *BaseReport) SetTimestamp(timestamp time.Time) {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	report.doSetTimestamp(timestamp)
}
func (report *BaseReport) doSetTimestamp(timestamp time.Time) {
	report.Timestamp = timestamp
}

// ============================================ GET ============================================
func (report *BaseReport) GetActionName() string {
	return report.ActionName
}

func (report *BaseReport) GetStatus() string {
	return report.Status
}

func (report *BaseReport) GetErrorList() []string {
	return report.Errors
}

func (report *BaseReport) GetTarget() string {
	return report.Target
}

func (report *BaseReport) GetReporter() string {
	return report.Reporter
}

func (report *BaseReport) GetActionID() string {
	return report.ActionID
}

func (report *BaseReport) GetJobID() string {
	return report.JobID
}

func (report *BaseReport) GetParentAction() string {
	return report.ParentAction
}

func (report *BaseReport) GetCustomerGUID() string {
	return report.CustomerGUID
}

func (report *BaseReport) GetActionIDN() int {
	return report.ActionIDN
}

func (report *BaseReport) GetTimestamp() time.Time {
	return report.Timestamp
}

func (report *BaseReport) GetDetails() string {
	return report.Details
}
