package types

const (
	OperatorCommandGroup  string = "kubescape.io"
	OperatorCommandKind   string = "OperatorCommand"
	OperatorCommandPlural string = "operatorcommands"
)

type ResponseType struct {
	PodName       string  `json:"podName,omitempty"`
	NameSpace     string  `json:"namespace,omitempty"`
	ContainerName string  `json:"containerName,omitempty"`
	Pid           *uint32 `json:"pid,omitempty"`
	Action        string  `json:"action,omitempty"`
}
