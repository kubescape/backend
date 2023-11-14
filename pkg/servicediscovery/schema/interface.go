package schema

import (
	"encoding/json"
	"io"
)

type IBackendServices interface {
	SetReportReceiverHttpUrl(string)
	SetReportReceiverWebsocketUrl(string)
	SetGatewayUrl(string)
	SetApiServerUrl(string)
	SetMetricsUrl(string)
	SetSynchronizerUrl(string)
	GetReportReceiverHttpUrl() string
	GetReportReceiverWebsocketUrl() string
	GetGatewayUrl() string
	GetApiServerUrl() string
	GetMetricsUrl() string
	GetSynchronizerUrl() string
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
