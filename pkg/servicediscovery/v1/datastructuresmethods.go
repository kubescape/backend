package v1

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

func NewServiceDiscoveryClientV1(url string) (*ServiceDiscoveryClientV1, error) {
	scheme, host, err := utils.ParseHost(url)
	if err != nil {
		return nil, err
	}
	return &ServiceDiscoveryClientV1{scheme: scheme, host: host, path: ServiceDiscoveryPathV1}, nil
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

func (sds *ServiceDiscoveryClientV1) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV1
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}

func (sds *ServiceDiscoveryClientV1) Get() (io.Reader, error) {
	response, err := http.Get(sds.GetServiceDiscoveryUrl())
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("server (%s) responded: %v", sds.GetHost(), response.StatusCode)
	}
	return response.Body, nil
}

func NewServiceDiscoveryServerV1(services ServicesV1) *ServiceDiscoveryServerV1 {
	return &ServiceDiscoveryServerV1{version: ApiVersion, services: services}
}

func (sds *ServiceDiscoveryServerV1) GetResponse() json.RawMessage {
	resp, _ := json.Marshal(sds.services)
	return resp
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

func NewServiceDiscoveryFileV1(path string) *ServiceDiscoveryFileV1 {
	return &ServiceDiscoveryFileV1{path: path}
}

func (s *ServiceDiscoveryFileV1) Get() (io.Reader, error) {
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

func (s *ServiceDiscoveryFileV1) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV1
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}

func NewServiceDiscoveryStreamV1(data []byte) *ServiceDiscoveryStreamV1 {
	return &ServiceDiscoveryStreamV1{data: data}
}

func (s *ServiceDiscoveryStreamV1) Get() (io.Reader, error) {
	return bytes.NewReader(s.data), nil
}

func (s *ServiceDiscoveryStreamV1) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV1
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}
