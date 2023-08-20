package servicediscovery

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kubescape/backend/pkg/servicediscovery/schema"
	"golang.org/x/tools/go/packages"
)

// WriteServiceDiscoveryResponse writes the service discovery response to the HTTP response writer
// This is used by the service discovery server to respond to HTTP GET requests
func WriteServiceDiscoveryResponse(w http.ResponseWriter, sds schema.IServiceDiscoveryServer) {
	if cachedResponse, exist := sds.GetCachedResponse(); exist {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(cachedResponse)
		w.WriteHeader(http.StatusOK)
		return
	}

	serviceMap := schema.ServiceDiscoveryResponse{
		Version:  sds.GetVersion(),
		Response: sds.GetResponse(),
	}

	res, err := json.Marshal(serviceMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sds.CacheResponse(res)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(res)
	w.WriteHeader(http.StatusOK)
}

// GetServices returns the services from the service discovery server via HTTP GET request
// This is used by the service discovery client to get the services from the service discovery server
func GetServices(sdc schema.IServiceDiscoveryClient) (schema.IBackendServices, error) {
	response, err := http.Get(sdc.GetServiceDiscoveryUrl())
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("server (%s) responded: %v", sdc.GetHost(), response.StatusCode)
	}

	var serviceResponse schema.ServiceDiscoveryResponse
	dec := json.NewDecoder(response.Body)
	if err = dec.Decode(&serviceResponse); err != nil {
		return nil, fmt.Errorf("server (%s) returned invalid response", sdc.GetHost())
	}

	if !VersionImplementationExist(serviceResponse.Version) {
		return nil, fmt.Errorf("server (%s) returned invalid version (%s)", sdc.GetHost(), serviceResponse.Version)
	}

	return sdc.ParseResponse(serviceResponse.Response)
}

func VersionImplementationExist(version string) bool {
	dir := fmt.Sprintf("./%s", version)
	cfg := &packages.Config{Mode: packages.NeedName, Dir: dir}
	pkgs, err := packages.Load(cfg, dir)
	if err != nil || len(pkgs) == 0 {
		return false
	}
	return true
}
