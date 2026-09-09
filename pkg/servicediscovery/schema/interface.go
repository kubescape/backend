package schema

import (
	"encoding/json"
	"io"
)

type IBackendServices interface {
	SetReportReceiverHttpUrl(string)
	// deprecated - use v1 or v2
	SetReportReceiverWebsocketUrl(string)
	// deprecated - use v1 or v2
	SetGatewayUrl(string)
	SetApiServerUrl(string)
	SetStorageUrl(string)
	SetMetricsUrl(string)
	SetSynchronizerUrl(string)
	// v3 only: OTLP/gRPC endpoint for agent event telemetry; empty when not configured
	SetOtelEventsUrl(string)
	GetReportReceiverHttpUrl() string
	// deprecated - use v1 or v2
	GetReportReceiverWebsocketUrl() string
	// deprecated - use v1 or v2
	GetGatewayUrl() string
	GetApiServerUrl() string
	GetMetricsUrl() string
	GetSynchronizerUrl() string
	GetStorageUrl() string
	// v3 only: OTLP/gRPC endpoint for agent event telemetry; empty when not configured
	GetOtelEventsUrl() string
}

type IServiceDiscoveryClient interface {
	IServiceDiscoveryServiceGetter
	GetHost() string
	GetScheme() string
	GetPath() string
	GetServiceDiscoveryUrl() string
}

type IServiceDiscoveryServer interface {
	GetVersion() string
	GetResponse() json.RawMessage
	GetCachedResponse() ([]byte, bool)
	CacheResponse([]byte)
}

type IServiceDiscoveryServiceGetter interface {
	Get() (io.Reader, error)
	ParseResponse(json.RawMessage) (IBackendServices, error)
}
