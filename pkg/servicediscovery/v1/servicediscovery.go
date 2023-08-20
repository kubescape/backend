package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteServiceDiscoveryResponse writes the service discovery response to the HTTP response writer
// This is used by the service discovery server to respond to HTTP GET requests
func WriteServiceDiscoveryResponse(w http.ResponseWriter, serviceDiscovery ServiceDiscoveryResponse) {
	serviceMap := ServiceDiscoveryResponse{
		EventReceiverHttpUrl:      serviceDiscovery.GetReportReceiverHttpUrl(),
		EventReceiverWebsocketUrl: serviceDiscovery.GetReportReceiverWebsocketUrl(),
		GatewayUrl:                serviceDiscovery.GetGatewayUrl(),
		ApiServerUrl:              serviceDiscovery.GetApiServerUrl(),
		MetricsUrl:                serviceDiscovery.GetMetricsUrl(),
	}

	res, err := json.Marshal(serviceMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(res)
	w.WriteHeader(http.StatusOK)
}

// GetServices returns the services from the service discovery server via HTTP GET request
// This is used by the service discovery client to get the services from the service discovery server
func GetServices(ctx context.Context, server *ServiceDiscoveryServer) (IBackendServices, error) {
	response, err := http.Get(server.GetServiceDiscoveryUrl())
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("server (%s) responded: %v", server.host, response.StatusCode)
	}

	var serviceResponse ServiceDiscoveryResponse
	dec := json.NewDecoder(response.Body)
	if err = dec.Decode(&serviceResponse); err != nil {
		return nil, fmt.Errorf("server (%s) returned invalid response", server.host)
	}

	return &serviceResponse, nil
}
