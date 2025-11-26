package v3

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

func NewServiceDiscoveryClientV3(url string) (*ServiceDiscoveryClientV3, error) {
	scheme, host, err := utils.ParseHost(url)
	if err != nil {
		return nil, err
	}
	return &ServiceDiscoveryClientV3{scheme: scheme, host: host, path: ServiceDiscoveryPathV3}, nil
}

func (sds *ServiceDiscoveryClientV3) GetServiceDiscoveryUrl() string {
	u := url.URL{
		Host:   sds.host,
		Scheme: sds.scheme,
		Path:   sds.path,
	}
	return u.String()
}

func (sds *ServiceDiscoveryClientV3) GetHost() string {
	return sds.host
}
func (sds *ServiceDiscoveryClientV3) GetPath() string {
	return sds.path
}

func (sds *ServiceDiscoveryClientV3) GetScheme() string {
	return sds.scheme
}

func (sds *ServiceDiscoveryClientV3) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV3
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}

func (sds *ServiceDiscoveryClientV3) Get() (io.Reader, error) {
	response, err := http.Get(sds.GetServiceDiscoveryUrl())
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("server (%s) responded: %v", sds.GetHost(), response.StatusCode)
	}
	return response.Body, nil
}

func NewServiceDiscoveryServerV3(services ServicesV3) *ServiceDiscoveryServerV3 {
	return &ServiceDiscoveryServerV3{version: ApiVersion, services: services}
}

func (sds *ServiceDiscoveryServerV3) GetResponse() json.RawMessage {
	resp, _ := json.Marshal(sds.services)
	return resp
}

func (sds *ServiceDiscoveryServerV3) GetVersion() string {
	return sds.version
}

func (sds *ServiceDiscoveryServerV3) GetCachedResponse() ([]byte, bool) {
	return sds.cachedResponse, sds.cachedResponse != nil
}

func (sds *ServiceDiscoveryServerV3) CacheResponse(response []byte) {
	if sds.cachedResponse == nil {
		sds.cachedResponse = response
	}
}

func (s *ServicesV3) SetReportReceiverHttpUrl(val string) {
	s.EventReceiverHttpUrl = val
}

// deprecated
func (s *ServicesV3) SetReportReceiverWebsocketUrl(val string) {
	panic("deprecated method called")
}

func (s *ServicesV3) SetApiServerUrl(val string) {
	s.ApiServerUrl = val
}

func (s *ServicesV3) SetMetricsUrl(val string) {
	s.MetricsUrl = val
}

func (s *ServicesV3) GetReportReceiverHttpUrl() string {
	return s.EventReceiverHttpUrl
}

// deprecated
func (s *ServicesV3) GetReportReceiverWebsocketUrl() string {
	panic("deprecated method called")
}

func (s *ServicesV3) GetApiServerUrl() string {
	return s.ApiServerUrl
}

func (s *ServicesV3) GetMetricsUrl() string {
	return s.MetricsUrl
}

func (s *ServicesV3) SetSynchronizerUrl(val string) {
	s.SynchronizerUrl = val
}

func (s *ServicesV3) GetSynchronizerUrl() string {
	return s.SynchronizerUrl
}

// deprecated
func (s *ServicesV3) SetGatewayUrl(val string) {
	panic("deprecated method called")
}

// deprecated
func (s *ServicesV3) GetGatewayUrl() string {
	panic("deprecated method called")
}

func (s *ServicesV3) SetStorageUrl(val string) {
	s.StorageUrl = val
}

func (s *ServicesV3) GetStorageUrl() string {
	return s.StorageUrl
}

func NewServiceDiscoveryFileV3(path string) *ServiceDiscoveryFileV3 {
	return &ServiceDiscoveryFileV3{path: path}
}

func (s *ServiceDiscoveryFileV3) Get() (io.Reader, error) {
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

func (s *ServiceDiscoveryFileV3) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV3
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}

func NewServiceDiscoveryStreamV3(data []byte) *ServiceDiscoveryStreamV3 {
	return &ServiceDiscoveryStreamV3{data: data}
}

func (s *ServiceDiscoveryStreamV3) Get() (io.Reader, error) {
	return bytes.NewReader(s.data), nil
}

func (s *ServiceDiscoveryStreamV3) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV3
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}
