package v2

import "github.com/kubescape/backend/pkg/servicediscovery/schema"

type ServiceDiscoveryClientV2 struct {
	host   string
	scheme string
	path   string
}

type ServiceDiscoveryServerV2 struct {
	version        string
	services       ServicesV2
	cachedResponse []byte
}

type ServicesV2 struct {
	schema.IBackendServices `json:"-"`

	EventReceiverHttpUrl      string `json:"event-receiver-http"`
	EventReceiverWebsocketUrl string `json:"event-receiver-ws"`
	GatewayUrl                string `json:"gateway"`
	ApiServerUrl              string `json:"api-server"`
	MetricsUrl                string `json:"metrics"`
	SynchronizerUrl           string `json:"synchronizer"`
}

type ServiceDiscoveryFileV2 struct {
	path string
}

type ServiceDiscoveryStreamV2 struct {
	data []byte
}
