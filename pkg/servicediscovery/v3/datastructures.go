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

	EventReceiverHttpUrl string `json:"event-receiver-http"`
	ApiServerUrl         string `json:"api-server"`
	MetricsUrl           string `json:"metrics"`
	SynchronizerUrl      string `json:"synchronizer"`
	StorageUrl           string `json:"storage"`
	// OtelEventsUrl is the OTLP/gRPC endpoint (host:port, TLS) that in-cluster agents export
	// raw event telemetry to (the AI-Sandbox OTel collector). Optional: absent when the
	// backend has no such collector configured.
	OtelEventsUrl string `json:"otel-events,omitempty"`
}

type ServiceDiscoveryFileV3 struct {
	path string
}

type ServiceDiscoveryStreamV3 struct {
	data []byte
}
