package v1

import (
	"fmt"
	"net/url"

	"github.com/kubescape/backend/pkg/servicediscovery/schema"
	"github.com/kubescape/backend/pkg/utils"
)

func NewServiceDiscoveryClientV1(url string) *ServiceDiscoveryClientV1 {
	scheme, host := utils.ParseHost(url)
	return &ServiceDiscoveryClientV1{scheme: scheme, host: host, path: ServiceDiscoveryPathV1}
}

func (sds *ServiceDiscoveryClientV1) GetServiceDiscoveryUrl() string {
	u := url.URL{
		Host:   sds.host,
		Scheme: sds.scheme,
		Path:   sds.path,
	}
	return u.String()
}

func (sds *ServiceDiscoveryClientV1) GetHost() string {
	return sds.host
}
func (sds *ServiceDiscoveryClientV1) GetPath() string {
	return sds.path
}

func (sds *ServiceDiscoveryClientV1) GetScheme() string {
	return sds.scheme
}

func (sds *ServiceDiscoveryClientV1) ParseResponse(response any) (schema.IBackendServices, error) {
	if res, ok := response.(ServicesV1); ok {
		return &res, nil
	}
	return nil, fmt.Errorf("invalid response")
}

func NewServiceDiscoveryServerV1(services ServicesV1) *ServiceDiscoveryServerV1 {
	return &ServiceDiscoveryServerV1{version: ApiVersion, services: services}
}

func (sds *ServiceDiscoveryServerV1) GetResponse() any {
	return sds.services
}

func (sds *ServiceDiscoveryServerV1) GetVersion() string {
	return sds.version
}

func (sds *ServiceDiscoveryServerV1) GetCachedResponse() ([]byte, bool) {
	return sds.cachedResponse, sds.cachedResponse != nil
}

func (sds *ServiceDiscoveryServerV1) CacheResponse(response []byte) {
	if sds.cachedResponse == nil {
		sds.cachedResponse = response
	}
}

func (s *ServicesV1) SetReportReceiverHttpUrl(val string) {
	s.EventReceiverHttpUrl = val
}

func (s *ServicesV1) SetReportReceiverWebsocketUrl(val string) {
	s.EventReceiverWebsocketUrl = val
}

func (s *ServicesV1) SetGatewayUrl(val string) {
	s.GatewayUrl = val
}

func (s *ServicesV1) SetApiServerUrl(val string) {
	s.ApiServerUrl = val
}

func (s *ServicesV1) SetMetricsUrl(val string) {
	s.MetricsUrl = val
}

func (s *ServicesV1) GetReportReceiverHttpUrl() string {
	return s.EventReceiverHttpUrl
}

func (s *ServicesV1) GetReportReceiverWebsocketUrl() string {
	return s.EventReceiverWebsocketUrl
}

func (s *ServicesV1) GetGatewayUrl() string {
	return s.GatewayUrl
}

func (s *ServicesV1) GetApiServerUrl() string {
	return s.ApiServerUrl
}

func (s *ServicesV1) GetMetricsUrl() string {
	return s.MetricsUrl
}
