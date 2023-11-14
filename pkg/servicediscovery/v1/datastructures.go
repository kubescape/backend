package v1

import "github.com/kubescape/backend/pkg/servicediscovery/schema"

type ServiceDiscoveryClientV1 struct {
	host   string
	scheme string
	path   string
}

type ServiceDiscoveryServerV1 struct {
	version        string
	services       ServicesV1
	cachedResponse []byte
}

type ServicesV1 struct {
	schema.IBackendServices `json:"-"`

	EventReceiverHttpUrl      string `json:"event-receiver-http"`
	EventReceiverWebsocketUrl string `json:"event-receiver-ws"`
	GatewayUrl                string `json:"gateway"`
	ApiServerUrl              string `json:"api-server"`
	MetricsUrl                string `json:"metrics"`
}

type ServiceDiscoveryFileV1 struct {
	path string
}

type ServiceDiscoveryStreamV1 struct {
	data []byte
}
