// Package v1beta1 vendors the ApplicationProfile and NetworkNeighborhood
// CRD types that github.com/kubescape/storage removed in
// https://github.com/kubescape/storage/pull/351 (storage v0.0.303), where
// they were superseded by the unified ContainerProfile type.
//
// The backend's gRPC API (see pkg/client/v1/proto/storage_service.proto)
// still exposes these types on the wire for backward compatibility with
// existing clients, so their shape - including field ordering and
// protobuf tags - is preserved here verbatim from
// github.com/kubescape/storage v0.0.302, the last version that defined
// them. Sub-types that storage still exports (ExecCalls, OpenCalls,
// SingleSeccompProfile, HTTPEndpoint, RulePolicy, IdentifiedCallStack,
// NetworkNeighbor) are reused directly from the storage package rather
// than duplicated here.
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	storagev1beta1 "github.com/kubescape/storage/pkg/apis/softwarecomposition/v1beta1"
)

// ApplicationProfile represents a learned application profile for a workload.
//
// Deprecated: removed from github.com/kubescape/storage in favor of
// ContainerProfile. Kept here for gRPC wire compatibility.
type ApplicationProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ApplicationProfileSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ApplicationProfileStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type ApplicationProfileSpec struct {
	Architectures []string `json:"architectures" protobuf:"bytes,1,rep,name=architectures"`
	// +patchMergeKey=name
	// +patchStrategy=merge
	Containers []ApplicationProfileContainer `json:"containers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	// +patchMergeKey=name
	// +patchStrategy=merge
	InitContainers []ApplicationProfileContainer `json:"initContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,3,rep,name=initContainers"`
	// +patchMergeKey=name
	// +patchStrategy=merge
	EphemeralContainers []ApplicationProfileContainer `json:"ephemeralContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,4,rep,name=ephemeralContainers"`
}

type ApplicationProfileContainer struct {
	Name         string   `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Capabilities []string `json:"capabilities" protobuf:"bytes,2,rep,name=capabilities"`
	// +patchMergeKey=path
	// +patchStrategy=merge
	Execs []storagev1beta1.ExecCalls `json:"execs" patchStrategy:"merge" patchMergeKey:"path" protobuf:"bytes,3,rep,name=execs"`
	// +patchMergeKey=path
	// +patchStrategy=merge
	Opens          []storagev1beta1.OpenCalls          `json:"opens" patchStrategy:"merge" patchMergeKey:"path" protobuf:"bytes,4,rep,name=opens"`
	Syscalls       []string                            `json:"syscalls" protobuf:"bytes,5,rep,name=syscalls"`
	SeccompProfile storagev1beta1.SingleSeccompProfile `json:"seccompProfile,omitempty" protobuf:"bytes,6,opt,name=seccompProfile"`
	// +patchStrategy=merge
	// +patchMergeKey=endpoint
	Endpoints []storagev1beta1.HTTPEndpoint `json:"endpoints" patchStrategy:"merge" patchMergeKey:"endpoint" protobuf:"bytes,7,rep,name=endpoints"`
	ImageID   string                        `json:"imageID" protobuf:"bytes,8,opt,name=imageID"`
	ImageTag  string                        `json:"imageTag" protobuf:"bytes,9,opt,name=imageTag"`
	// +patchStrategy=merge
	// +patchMergeKey=ruleId
	PolicyByRuleId       map[string]storagev1beta1.RulePolicy `json:"rulePolicies" protobuf:"bytes,10,rep,name=rulePolicies" patchStrategy:"merge" patchMergeKey:"ruleId"`
	IdentifiedCallStacks []storagev1beta1.IdentifiedCallStack `json:"identifiedCallStacks" protobuf:"bytes,11,rep,name=identifiedCallStacks"`
}

type ApplicationProfileStatus struct {
}

type ApplicationProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ApplicationProfile `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// NetworkNeighborhood represents a list of network communications for a specific workload.
//
// Deprecated: removed from github.com/kubescape/storage in favor of
// ContainerProfile. Kept here for gRPC wire compatibility.
type NetworkNeighborhood struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec NetworkNeighborhoodSpec `json:"spec" protobuf:"bytes,2,req,name=spec"`
}

type NetworkNeighborhoodSpec struct {
	metav1.LabelSelector `json:",inline" protobuf:"bytes,3,opt,name=labelSelector"`
	Containers           []NetworkNeighborhoodContainer `json:"containers" protobuf:"bytes,4,rep,name=containers"`
	InitContainers       []NetworkNeighborhoodContainer `json:"initContainers" protobuf:"bytes,5,rep,name=initContainers"`
	EphemeralContainers  []NetworkNeighborhoodContainer `json:"ephemeralContainers" protobuf:"bytes,6,rep,name=ephemeralContainers"`
}

type NetworkNeighborhoodContainer struct {
	Name    string                           `json:"name" protobuf:"bytes,1,req,name=name"`
	Ingress []storagev1beta1.NetworkNeighbor `json:"ingress" protobuf:"bytes,2,rep,name=ingress"`
	Egress  []storagev1beta1.NetworkNeighbor `json:"egress" protobuf:"bytes,3,rep,name=egress"`
}

type NetworkNeighborhoodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []NetworkNeighborhood `json:"items" protobuf:"bytes,2,rep,name=items"`
}
