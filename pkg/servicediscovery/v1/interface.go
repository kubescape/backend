package v1

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
