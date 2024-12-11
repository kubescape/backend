package utils

// annotations added to the workload
const (
	ArmoPrefix        string = "armo"
	ArmoUpdate        string = ArmoPrefix + ".last-update"
	ArmoWlid          string = ArmoPrefix + ".wlid"
	ArmoSid           string = ArmoPrefix + ".sid"
	ArmoJobID         string = ArmoPrefix + ".job"
	ArmoJobIDPath     string = ArmoJobID + "/id"
	ArmoJobParentPath string = ArmoJobID + "/parent"
	ArmoJobActionPath string = ArmoJobID + "/action"
)
