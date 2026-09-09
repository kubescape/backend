package v4

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

func NewServiceDiscoveryClientV4(url string) (*ServiceDiscoveryClientV4, error) {
	scheme, host, err := utils.ParseHost(url)
	if err != nil {
		return nil, err
	}
	return &ServiceDiscoveryClientV4{scheme: scheme, host: host, path: ServiceDiscoveryPathV4}, nil
}

func (sds *ServiceDiscoveryClientV4) GetServiceDiscoveryUrl() string {
	u := url.URL{
		Host:   sds.host,
		Scheme: sds.scheme,
		Path:   sds.path,
	}
	return u.String()
}

func (sds *ServiceDiscoveryClientV4) GetHost() string {
	return sds.host
}
func (sds *ServiceDiscoveryClientV4) GetPath() string {
	return sds.path
}

func (sds *ServiceDiscoveryClientV4) GetScheme() string {
	return sds.scheme
}

func (sds *ServiceDiscoveryClientV4) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV4
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}

func (sds *ServiceDiscoveryClientV4) Get() (io.Reader, error) {
	response, err := http.Get(sds.GetServiceDiscoveryUrl())
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("server (%s) responded: %v", sds.GetHost(), response.StatusCode)
	}
	return response.Body, nil
}

func NewServiceDiscoveryServerV4(services ServicesV4) *ServiceDiscoveryServerV4 {
	return &ServiceDiscoveryServerV4{version: ApiVersion, services: services}
}

func (sds *ServiceDiscoveryServerV4) GetResponse() json.RawMessage {
	resp, _ := json.Marshal(sds.services)
	return resp
}

func (sds *ServiceDiscoveryServerV4) GetVersion() string {
	return sds.version
}

func (sds *ServiceDiscoveryServerV4) GetCachedResponse() ([]byte, bool) {
	return sds.cachedResponse, sds.cachedResponse != nil
}

func (sds *ServiceDiscoveryServerV4) CacheResponse(response []byte) {
	if sds.cachedResponse == nil {
		sds.cachedResponse = response
	}
}

func (s *ServicesV4) SetReportReceiverHttpUrl(val string) {
	s.EventReceiverHttpUrl = val
}

// deprecated
func (s *ServicesV4) SetReportReceiverWebsocketUrl(val string) {
	panic("deprecated method called")
}

func (s *ServicesV4) SetApiServerUrl(val string) {
	s.ApiServerUrl = val
}

func (s *ServicesV4) SetMetricsUrl(val string) {
	s.MetricsUrl = val
}

func (s *ServicesV4) GetReportReceiverHttpUrl() string {
	return s.EventReceiverHttpUrl
}

// deprecated
func (s *ServicesV4) GetReportReceiverWebsocketUrl() string {
	panic("deprecated method called")
}

func (s *ServicesV4) GetApiServerUrl() string {
	return s.ApiServerUrl
}

func (s *ServicesV4) GetMetricsUrl() string {
	return s.MetricsUrl
}

func (s *ServicesV4) SetSynchronizerUrl(val string) {
	s.SynchronizerUrl = val
}

func (s *ServicesV4) GetSynchronizerUrl() string {
	return s.SynchronizerUrl
}

// deprecated
func (s *ServicesV4) SetGatewayUrl(val string) {
	panic("deprecated method called")
}

// deprecated
func (s *ServicesV4) GetGatewayUrl() string {
	panic("deprecated method called")
}

func (s *ServicesV4) SetStorageUrl(val string) {
	s.StorageUrl = val
}

func (s *ServicesV4) GetStorageUrl() string {
	return s.StorageUrl
}

func (s *ServicesV4) SetOtelEventsUrl(val string) {
	s.OtelEventsUrl = val
}

func (s *ServicesV4) GetOtelEventsUrl() string {
	return s.OtelEventsUrl
}

func NewServiceDiscoveryFileV4(path string) *ServiceDiscoveryFileV4 {
	return &ServiceDiscoveryFileV4{path: path}
}

func (s *ServiceDiscoveryFileV4) Get() (io.Reader, error) {
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

func (s *ServiceDiscoveryFileV4) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV4
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}

func NewServiceDiscoveryStreamV4(data []byte) *ServiceDiscoveryStreamV4 {
	return &ServiceDiscoveryStreamV4{data: data}
}

func (s *ServiceDiscoveryStreamV4) Get() (io.Reader, error) {
	return bytes.NewReader(s.data), nil
}

func (s *ServiceDiscoveryStreamV4) ParseResponse(response json.RawMessage) (schema.IBackendServices, error) {
	var services ServicesV4
	if err := json.Unmarshal(response, &services); err == nil {
		return &services, nil
	}

	return nil, fmt.Errorf("invalid response")
}
