package v4

import "github.com/kubescape/backend/pkg/servicediscovery/schema"

type ServiceDiscoveryClientV4 struct {
	host   string
	scheme string
	path   string
}

type ServiceDiscoveryServerV4 struct {
	version        string
	services       ServicesV4
	cachedResponse []byte
}

type ServicesV4 struct {
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

type ServiceDiscoveryFileV4 struct {
	path string
}

type ServiceDiscoveryStreamV4 struct {
	data []byte
}
