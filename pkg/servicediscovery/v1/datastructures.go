package v1

type ServiceDiscoveryServer struct {
	host   string
	scheme string
	path   string
}

type ServiceDiscoveryResponse struct {
	EventReceiverHttpUrl      string `json:"event-receiver-http"`
	EventReceiverWebsocketUrl string `json:"event-receiver-ws"`
	GatewayUrl                string `json:"gateway"`
	ApiServerUrl              string `json:"api-server"`
	MetricsUrl                string `json:"metrics"`
}
