package v1alpha1

import (
	"github.com/kubescape/backend/pkg/command/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	OperatorCommandVersion string = "v1alpha1"
)

var SchemaGroupVersionResource = schema.GroupVersionResource{
	Group:    types.OperatorCommandGroup,
	Version:  OperatorCommandVersion,
	Resource: types.OperatorCommandPlural,
}
