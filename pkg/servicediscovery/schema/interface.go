package schema

import "encoding/json"

type IBackendServices interface {
	SetReportReceiverHttpUrl(string)
	SetReportReceiverWebsocketUrl(string)
	SetGatewayUrl(string)
	SetApiServerUrl(string)
	SetMetricsUrl(string)
	GetReportReceiverHttpUrl() string
	GetReportReceiverWebsocketUrl() string
	GetGatewayUrl() string
	GetApiServerUrl() string
	GetMetricsUrl() string
}

type IServiceDiscoveryClient interface {
	GetHost() string
	GetScheme() string
	GetPath() string
	GetServiceDiscoveryUrl() string
	ParseResponse(json.RawMessage) (IBackendServices, error)
}

type IServiceDiscoveryServer interface {
	GetVersion() string
	GetResponse() json.RawMessage
	GetCachedResponse() ([]byte, bool)
	CacheResponse([]byte)
}
