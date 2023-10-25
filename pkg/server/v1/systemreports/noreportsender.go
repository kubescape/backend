package systemreports

import (
	"time"
)

var _ IReporter = &noReportSender{}

type noReportSender struct{}

// NewNoReportSender returns a Sender that sends no reports
//
// It stands in for a proper report sender when the client needs no reports to
// be sent
func NewNoReportSender() *noReportSender {
	return &noReportSender{}
}

func (s *noReportSender) GetReportID() string { return "" }

func (s *noReportSender) AddError(errorString string) {}
func (s *noReportSender) GetNextActionId() string     { return "" }
func (s *noReportSender) NextActionID()               {}

func (s *noReportSender) SimpleReportAnnotations(setParent bool, setCurrent bool) (string, string) {
	return "", ""
}

func (s *noReportSender) SetReporter(string)     {}
func (s *noReportSender) SetStatus(string)       {}
func (s *noReportSender) SetActionName(string)   {}
func (s *noReportSender) SetTarget(string)       {}
func (s *noReportSender) SetActionID(string)     {}
func (s *noReportSender) SetJobID(string)        {}
func (s *noReportSender) SetParentAction(string) {}
func (s *noReportSender) SetTimestamp(time.Time) {}
func (s *noReportSender) SetActionIDN(int)       {}
func (s *noReportSender) SetCustomerGUID(string) {}
func (s *noReportSender) SetDetails(string)      {}

func (s *noReportSender) GetReporter() string     { return "" }
func (s *noReportSender) GetStatus() string       { return "" }
func (s *noReportSender) GetActionName() string   { return "" }
func (s *noReportSender) GetTarget() string       { return "" }
func (s *noReportSender) GetErrorList() []string  { return []string{""} }
func (s *noReportSender) GetActionID() string     { return "" }
func (s *noReportSender) GetJobID() string        { return "" }
func (s *noReportSender) GetParentAction() string { return "" }
func (s *noReportSender) GetTimestamp() time.Time {
	return time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
}
func (s *noReportSender) GetActionIDN() int       { return -1 }
func (s *noReportSender) GetCustomerGUID() string { return "" }
func (s *noReportSender) GetDetails() string      { return "" }

func (s *noReportSender) Send() (int, string, error) { return 200, "", nil }

func (s *noReportSender) SendAsRoutine(bool) {}

func (s *noReportSender) SendAction(action string, sendReport bool)                      {}
func (s *noReportSender) SendError(err error, sendReport bool, initErrors bool)          {}
func (s *noReportSender) SendStatus(status string, sendReport bool)                      {}
func (s *noReportSender) SendDetails(details string, sendReport bool)                    {}
func (s *noReportSender) SendWarning(warning string, sendReport bool, initWarnings bool) {}
