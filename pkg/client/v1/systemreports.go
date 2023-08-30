package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	httputils "github.com/armosec/utils-go/httputils"
	v1 "github.com/kubescape/backend/pkg/server/v1"
	"github.com/kubescape/backend/pkg/server/v1/systemreports"
	"github.com/kubescape/backend/pkg/utils"
)

var (
	_ IReportSender = &BaseReportSender{}

	systemReportEndpoint = &sysEndpoint{}

	MAX_RETRIES int           = 3
	RETRY_DELAY time.Duration = time.Second * 5
)

type IReportSender interface {
	systemreports.IReporter

	Send() (int, string, error) //send logic here

	/*
		SendAsRoutine
		@input:
		collector []string - leave as empty (a way to hold all previous failed reports and send them in bulk)
		progressNext bool - increase actionID, sometimes u send parallel jobs that have the same order - (vuln scanning a cluster for eg. all wl scans have the same order)
		errChan - chan to allow the goroutine to return the errors inside
	*/
	SendAsRoutine(bool, chan<- error) //goroutine wrapper

	// set methods
	SendAction(action string, sendReport bool, errChan chan<- error)
	SendError(err error, sendReport bool, initErrors bool, errChan chan<- error)
	SendStatus(status string, sendReport bool, errChan chan<- error)
	SendDetails(details string, sendReport bool, errChan chan<- error)
	SendWarning(warning string, sendReport bool, initWarnings bool, errChan chan<- error)
}

type BaseReportSender struct {
	eventReceiverUrl string
	httpClient       httputils.IHttpClient
	report           *systemreports.BaseReport
}

type sysEndpoint struct {
	value string
	mu    sync.RWMutex
}

func NewBaseReportSender(eventReceiverUrl string, httpClient httputils.IHttpClient, report *systemreports.BaseReport) *BaseReportSender {
	return &BaseReportSender{
		eventReceiverUrl: eventReceiverUrl,
		httpClient:       httpClient,
		report:           report,
	}
}

func (e *sysEndpoint) IsEmpty() bool {
	return e.Get() == ""
}

func (e *sysEndpoint) Set(value string) {
	e.mu.Lock()
	e.value = value
	e.mu.Unlock()
}

func (e *sysEndpoint) Get() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.value
}

// SetOrDefault sets the system report endpoint to the provided value for valid
// inputs, or to a default value on invalid ones
//
// An empty input is considered invalid, and would thus be set to a default endpoint
func (e *sysEndpoint) SetOrDefault(value string) {
	if value == "" {
		value = v1.ReporterSystemReportPath
	}
	e.Set(value)
}

// GetOrDefault returns the value of the current system report endpoint if it
// is set. If not set, it sets the value to a default and returns the newly set
// value
func (e *sysEndpoint) GetOrDefault() string {
	current := e.Get()
	if current == "" {
		e.SetOrDefault("")
	}
	return e.Get()
}

// Send - send http request. returns-> http status code, return message (jobID/OK), http/go error
func (s *BaseReportSender) Send() (int, string, error) {
	scheme, host, err := utils.ParseHost(s.eventReceiverUrl)
	if err != nil {
		return 500, fmt.Sprintf("invalid url: %s", s.eventReceiverUrl), err
	}
	url := url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   systemReportEndpoint.GetOrDefault(),
	}

	s.report.Timestamp = time.Now()
	if s.report.ActionID == "" {
		s.report.ActionID = "1"
		s.report.ActionIDN = 1
	}
	reqBody, err := json.Marshal(s.report)

	if err != nil {
		return 500, "Couldn't marshall report object", err
	}
	var resp *http.Response
	var bodyAsStr string
	for i := 0; i < MAX_RETRIES; i++ {
		resp, err = httputils.HttpPost(s.httpClient, url.String(), map[string]string{"Content-Type": "application/json"}, reqBody)
		bodyAsStr = "body could not be fetched"
		retry := err != nil
		if resp != nil {
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				retry = true
			}
			if resp.Body != nil {
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					bodyAsStr = string(body)
				}
				resp.Body.Close()
			}
		}
		if !retry {
			break
		}
		//else err != nil
		e := fmt.Errorf("attempt #%d %s - Failed posting report. Url: '%s', reason: '%s' report: '%s' response: '%s'", i, s.report.GetReportID(), url.String(), err.Error(), string(reqBody), bodyAsStr)

		if i == MAX_RETRIES-1 {
			return 500, e.Error(), err
		}
		//wait 5 secs between retries
		time.Sleep(RETRY_DELAY)
	}
	//first successful report gets it's jobID/proccessID
	if len(s.report.JobID) == 0 && bodyAsStr != "ok" && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.report.JobID = bodyAsStr
	}
	return resp.StatusCode, bodyAsStr, nil

}

// The caller must read the errChan, to prevent the goroutine from waiting in memory forever
func (sender *BaseReportSender) SendAsRoutine(progressNext bool, errChan chan<- error) {
	sender.report.Mutex.Lock()
	defer sender.report.Mutex.Unlock()

	if err := sender.unprotectedSendAsRoutine(progressNext); err != nil {
		errChan <- err
	}
}

// internal send as routine without mutex lock
func (sender *BaseReportSender) unprotectedSendAsRoutine(progressNext bool) error {
	defer recover()
	status, body, err := sender.Send()
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		err := fmt.Errorf("failed to send report. Status: %d Body:%s", status, body)
		return err
	}
	if progressNext {
		sender.report.NextActionID()
	}
	return nil
}

func (sender *BaseReportSender) SendError(err error, sendReport bool, initErrors bool, errChan chan<- error) {
	sender.report.Mutex.Lock()
	defer sender.report.Mutex.Unlock()

	if err != nil {
		e := fmt.Sprintf("Action: %s, Error: %s", sender.report.ActionName, err.Error())
		sender.report.Errors = append(sender.report.Errors, e)
	}
	sender.report.Status = systemreports.JobFailed

	if !sendReport {
		return
	}
	if err := sender.unprotectedSendAsRoutine(true); err != nil {
		errChan <- err
	}
}

func (sender *BaseReportSender) SendWarning(warnMsg string, sendReport bool, initWarnings bool, errChan chan<- error) {
	sender.report.Mutex.Lock()
	defer sender.report.Mutex.Unlock()

	if sender.report.Errors == nil {
		sender.report.Errors = make([]string, 0)
	}
	if len(warnMsg) != 0 {
		e := fmt.Sprintf("Action: %s, Error: %s", sender.report.ActionName, warnMsg)
		sender.report.Errors = append(sender.report.Errors, e)
	}
	sender.report.Status = systemreports.JobWarning

	if !sendReport {
		return
	}

	if err := sender.unprotectedSendAsRoutine(true); err != nil {
		errChan <- err
	}
}

func (sender *BaseReportSender) SendAction(actionName string, sendReport bool, errChan chan<- error) {
	sender.report.Mutex.Lock()
	defer sender.report.Mutex.Unlock()

	sender.report.DoSetActionName(actionName)
	if !sendReport {
		return
	}

	if err := sender.unprotectedSendAsRoutine(true); err != nil {
		errChan <- err
	}
}

func (sender *BaseReportSender) SendStatus(status string, sendReport bool, errChan chan<- error) {
	sender.report.Mutex.Lock()
	defer sender.report.Mutex.Unlock()

	sender.report.DoSetStatus(status)
	if !sendReport {
		return
	}

	if err := sender.unprotectedSendAsRoutine(true); err != nil {
		errChan <- err
	}
}

func (sender *BaseReportSender) SendDetails(details string, sendReport bool, errChan chan<- error) {
	sender.report.Mutex.Lock()
	defer sender.report.Mutex.Unlock()

	sender.report.DoSetDetails(details)
	if sendReport {
		return
	}

	if err := sender.unprotectedSendAsRoutine(true); err != nil {
		errChan <- err
	}
}

func (sender *BaseReportSender) GetReportID() string {
	return sender.report.GetReportID()
}

func (sender *BaseReportSender) AddError(errorString string) {
	sender.report.AddError(errorString)
}

func (sender *BaseReportSender) GetNextActionId() string {
	return sender.report.GetNextActionId()
}

func (sender *BaseReportSender) NextActionID() {
	sender.report.NextActionID()
}

func (sender *BaseReportSender) SimpleReportAnnotations(setParent bool, setCurrent bool) (string, string) {
	return sender.report.SimpleReportAnnotations(setParent, setCurrent)
}

func (sender *BaseReportSender) SetReporter(val string) {
	sender.report.SetReporter(val)
}

func (sender *BaseReportSender) SetStatus(val string) {
	sender.report.SetStatus(val)
}

func (sender *BaseReportSender) SetActionName(val string) {
	sender.report.SetActionName(val)
}

func (sender *BaseReportSender) SetTarget(val string) {
	sender.report.SetTarget(val)
}

func (sender *BaseReportSender) SetActionID(val string) {
	sender.report.SetActionID(val)
}

func (sender *BaseReportSender) SetJobID(val string) {
	sender.report.SetJobID(val)
}
func (sender *BaseReportSender) SetParentAction(val string) {
	sender.report.SetParentAction(val)
}

func (sender *BaseReportSender) SetTimestamp(val time.Time) {
	sender.report.SetTimestamp(val)
}

func (sender *BaseReportSender) SetActionIDN(val int) {
	sender.report.SetActionIDN(val)
}

func (sender *BaseReportSender) SetCustomerGUID(val string) {
	sender.report.SetCustomerGUID(val)
}
func (sender *BaseReportSender) SetDetails(val string) {
	sender.report.SetDetails(val)
}

func (sender *BaseReportSender) GetReporter() string {
	return sender.report.GetReporter()
}

func (sender *BaseReportSender) GetStatus() string {
	return sender.report.GetStatus()
}

func (sender *BaseReportSender) GetActionName() string {
	return sender.report.GetActionName()
}

func (sender *BaseReportSender) GetTarget() string {
	return sender.report.GetTarget()
}

func (sender *BaseReportSender) GetErrorList() []string {
	return sender.report.GetErrorList()
}

func (sender *BaseReportSender) GetActionID() string {
	return sender.report.GetActionID()
}

func (sender *BaseReportSender) GetJobID() string {
	return sender.report.GetJobID()
}

func (sender *BaseReportSender) GetParentAction() string {
	return sender.report.GetParentAction()
}

func (sender *BaseReportSender) GetTimestamp() time.Time {
	return sender.report.GetTimestamp()
}

func (sender *BaseReportSender) GetActionIDN() int {
	return sender.report.GetActionIDN()
}

func (sender *BaseReportSender) GetCustomerGUID() string {
	return sender.report.GetCustomerGUID()
}

func (sender *BaseReportSender) GetDetails() string {
	return sender.report.GetDetails()
}

func (sender *BaseReportSender) GetBaseReport() *systemreports.BaseReport {
	return sender.report
}
