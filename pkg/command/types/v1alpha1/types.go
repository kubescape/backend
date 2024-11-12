package v1alpha1

import (
	"time"

	"github.com/armosec/armoapi-go/identifiers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OperatorCommandList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []OperatorCommand `json:"items"`
}

type OperatorCommand struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatorCommandSpec   `json:"spec,omitempty"`
	Status OperatorCommandStatus `json:"status,omitempty"`
}

type OperatorCommandSpec struct {
	GUID           string                         `json:"guid"`                     // GUID is a unique identifier for the command
	CommandType    string                         `json:"commandType"`              // CommandType is the type of the command
	CommandVersion string                         `json:"commandVersion,omitempty"` // CommandVersion is the version of the command
	Designators    []identifiers.PortalDesignator `json:"designators,omitempty"`    // Designators are the designators for the command
	Body           []byte                         `json:"body,omitempty"`           // Body is the body of the command
	TTL            time.Duration                  `json:"ttl,omitempty"`            // TTL is the time to live for the command
	Args           map[string]interface{}         `json:"args,omitempty"`           // Args are the arguments for the command
	CommandIndex   *int                           `json:"commandIndex,omitempty"`   // CommandIndex is the index of the command in the sequence
	CommandCount   *int                           `json:"commandCount,omitempty"`   // CommandCount is the total number of commands in the sequence
}

type OperatorCommandStatus struct {
	Started     bool                        `json:"started"`               // Started indicates if the command has started
	StartedAt   *metav1.Time                `json:"startedAt,omitempty"`   // StartedAt is the time at which the command was started
	Completed   bool                        `json:"completed"`             // Completed indicates if the command has completed
	CompletedAt *metav1.Time                `json:"completedAt,omitempty"` // CompletedAt is the time at which the command was completed
	Executer    string                      `json:"executer,omitempty"`    // Executer is the entity that executed the command
	Error       *OperatorCommandStatusError `json:"error,omitempty"`       // Error is the error that occurred during the execution of the command (if any)
	Payload     []byte                      `json:"payload,omitempty"`     // Payload is the response payload from execution of the command (if any)
}

type OperatorCommandStatusError struct {
	Reason    string `json:"reason,omitempty"`    // reason for the error (optional)
	Message   string `json:"message,omitempty"`   // error message (optional)
	ErrorCode int    `json:"errorCode,omitempty"` // error code (optional)
}
