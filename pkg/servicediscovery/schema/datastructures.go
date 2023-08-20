package schema

// ServiceDiscoveryResponse is the response object that should be returned from the service discovery server
type ServiceDiscoveryResponse struct {
	Version  string `json:"version"`
	Response any    `json:"response"`
}
