package schema

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
	ParseResponse(any) (IBackendServices, error)
}

type IServiceDiscoveryServer interface {
	GetVersion() string
	GetResponse() any
	GetCachedResponse() ([]byte, bool)
	CacheResponse([]byte)
}
