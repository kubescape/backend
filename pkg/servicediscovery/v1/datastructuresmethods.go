package v1

import (
	"net/url"

	"github.com/kubescape/backend/pkg/utils"
)

func (sd *ServiceDiscoveryResponse) SetReportReceiverHttpUrl(val string) {
	sd.EventReceiverHttpUrl = val
}

func (sd *ServiceDiscoveryResponse) SetReportReceiverWebsocketUrl(val string) {
	sd.EventReceiverWebsocketUrl = val
}

func (sd *ServiceDiscoveryResponse) SetGatewayUrl(val string) {
	sd.GatewayUrl = val
}

func (sd *ServiceDiscoveryResponse) SetApiServerUrl(val string) {
	sd.ApiServerUrl = val
}

func (sd *ServiceDiscoveryResponse) SetMetricsUrl(val string) {
	sd.MetricsUrl = val
}

func (sd *ServiceDiscoveryResponse) GetReportReceiverHttpUrl() string {
	return sd.EventReceiverHttpUrl
}

func (sd *ServiceDiscoveryResponse) GetReportReceiverWebsocketUrl() string {
	return sd.EventReceiverWebsocketUrl
}

func (sd *ServiceDiscoveryResponse) GetGatewayUrl() string {
	return sd.GatewayUrl
}

func (sd *ServiceDiscoveryResponse) GetApiServerUrl() string {
	return sd.ApiServerUrl
}

func (sd *ServiceDiscoveryResponse) GetMetricsUrl() string {
	return sd.MetricsUrl
}

func NewServiceDiscoveryServer(url string) *ServiceDiscoveryServer {
	scheme, host := utils.ParseHost(url)
	return &ServiceDiscoveryServer{scheme: scheme, host: host, path: ServiceDiscoveryPath}
}

func (sds *ServiceDiscoveryServer) GetServiceDiscoveryUrl() string {
	u := url.URL{
		Host:   sds.host,
		Scheme: sds.scheme,
		Path:   sds.path,
	}
	return u.String()
}
