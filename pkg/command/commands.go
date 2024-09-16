package command

type OperatorCommandType string

const (
	OperatorCommandNodeNameLabelKey string = "nodeName" // holds the node name label, on which to run the command (optional)

	// command types will be defined here
	OperatorCommandTypeResponse OperatorCommandType = "response"
)

// ResponseCommand
type ResponseAction string

const (
	ResponseActionKill    ResponseAction = "Kill"
	ResponseActionStop    ResponseAction = "Stop"
	ResponseActionPause   ResponseAction = "Pause"
	ResponseActionUnpause ResponseAction = "Unpause"
)

type ResponseCommand struct {
	Namespace     string         `json:"namespace,omitempty"`
	PodName       string         `json:"podName,omitempty"`
	ContainerName string         `json:"containerName,omitempty"`
	Pid           *uint32        `json:"pid,omitempty"`
	Action        ResponseAction `json:"action,omitempty"`
}
