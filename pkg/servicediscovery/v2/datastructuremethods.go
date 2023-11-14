package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/kubescape/backend/pkg/servicediscovery/schema"
	"github.com/kubescape/backend/pkg/utils"
)

func NewServiceDiscoveryClientV2(url string) (*ServiceDiscoveryClientV2, error) {
	scheme, host, err := utils.ParseHost(url)
	if err != nil {
		return nil, err
	}
	return &ServiceDiscoveryClientV2{scheme: scheme, host: host, path: ServiceDiscoveryPathV2}, nil
}

func (sds *ServiceDiscoveryClientV2) GetServiceDiscoveryUrl() string {
	u := url.URL{
		Host:   sds.host,
		Scheme: sds.scheme,
		Path:   sds.path,
	}
	return u.String()
}

func (sds *ServiceDiscoveryClientV2) GetHost() string {
	return sds.host
}
func (sds *ServiceDiscoveryClientV2) GetPath() string {
	return sds.path
}

func (sds *ServiceDiscoveryClientV2) GetScheme() string {
	return sds.scheme
}

func (sds *ServiceDiscoveryClientV2) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV2
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}

func (sds *ServiceDiscoveryClientV2) Get() (io.Reader, error) {
	response, err := http.Get(sds.GetServiceDiscoveryUrl())
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("server (%s) responded: %v", sds.GetHost(), response.StatusCode)
	}
	return response.Body, nil
}

func NewServiceDiscoveryServerV2(services ServicesV2) *ServiceDiscoveryServerV2 {
	return &ServiceDiscoveryServerV2{version: ApiVersion, services: services}
}

func (sds *ServiceDiscoveryServerV2) GetResponse() json.RawMessage {
	resp, _ := json.Marshal(sds.services)
	return resp
}

func (sds *ServiceDiscoveryServerV2) GetVersion() string {
	return sds.version
}

func (sds *ServiceDiscoveryServerV2) GetCachedResponse() ([]byte, bool) {
	return sds.cachedResponse, sds.cachedResponse != nil
}

func (sds *ServiceDiscoveryServerV2) CacheResponse(response []byte) {
	if sds.cachedResponse == nil {
		sds.cachedResponse = response
	}
}

func (s *ServicesV2) SetReportReceiverHttpUrl(val string) {
	s.EventReceiverHttpUrl = val
}

func (s *ServicesV2) SetReportReceiverWebsocketUrl(val string) {
	s.EventReceiverWebsocketUrl = val
}

func (s *ServicesV2) SetApiServerUrl(val string) {
	s.ApiServerUrl = val
}

func (s *ServicesV2) SetMetricsUrl(val string) {
	s.MetricsUrl = val
}

func (s *ServicesV2) GetReportReceiverHttpUrl() string {
	return s.EventReceiverHttpUrl
}

func (s *ServicesV2) GetReportReceiverWebsocketUrl() string {
	return s.EventReceiverWebsocketUrl
}

func (s *ServicesV2) GetApiServerUrl() string {
	return s.ApiServerUrl
}

func (s *ServicesV2) GetMetricsUrl() string {
	return s.MetricsUrl
}

func (s *ServicesV2) SetSynchronizerUrl(val string) {
	s.SynchronizerUrl = val
}

func (s *ServicesV2) GetSynchronizerUrl() string {
	return s.SynchronizerUrl
}

func (s *ServicesV2) SetGatewayUrl(val string) {
	s.GatewayUrl = val
}

func (s *ServicesV2) GetGatewayUrl() string {
	return s.GatewayUrl
}

func NewServiceDiscoveryFileV2(path string) *ServiceDiscoveryFileV2 {
	return &ServiceDiscoveryFileV2{path: path}
}

func (s *ServiceDiscoveryFileV2) Get() (io.Reader, error) {
	jsonFile, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file (%s): %v", s.path, err)
	}
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file (%s): %v", s.path, err)
	}
	jsonFile.Close()

	return bytes.NewReader(data), nil
}

func (s *ServiceDiscoveryFileV2) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV2
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}

func NewServiceDiscoveryStreamV2(data []byte) *ServiceDiscoveryStreamV2 {
	return &ServiceDiscoveryStreamV2{data: data}
}

func (s *ServiceDiscoveryStreamV2) Get() (io.Reader, error) {
	return bytes.NewReader(s.data), nil
}

func (s *ServiceDiscoveryStreamV2) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV2
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}
