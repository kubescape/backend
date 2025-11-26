package v3

import "github.com/kubescape/backend/pkg/servicediscovery/schema"

type ServiceDiscoveryClientV3 struct {
	host   string
	scheme string
	path   string
}

type ServiceDiscoveryServerV3 struct {
	version        string
	services       ServicesV3
	cachedResponse []byte
}

type ServicesV3 struct {
	schema.IBackendServices `json:"-"`

	EventReceiverHttpUrl      string `json:"event-receiver-http"`
	EventReceiverWebsocketUrl string `json:"event-receiver-ws"`
	ApiServerUrl              string `json:"api-server"`
	MetricsUrl                string `json:"metrics"`
	SynchronizerUrl           string `json:"synchronizer"`
	GrpcServerUrl             string `json:"grpc-server"`
}

type ServiceDiscoveryFileV3 struct {
	path string
}

type ServiceDiscoveryStreamV3 struct {
	data []byte
}
