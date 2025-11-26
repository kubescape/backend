package v1

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/kubescape/backend/pkg/client/v1/proto"
	backendv1 "github.com/kubescape/backend/pkg/server/v1"
	"github.com/kubescape/storage/pkg/apis/softwarecomposition/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Default gRPC ports
const (
	DefaultGRPCPort  = 50051 // Non-secure gRPC
	DefaultGRPCSPort = 50052 // Secure gRPC
)

// GRPCConfig represents the parsed gRPC connection configuration
type GRPCConfig struct {
	IsSecure bool
	Host     string
	Port     int
	URL      string
}

// StorageClient provides a gRPC client for the Kubescape storage server
type StorageClient struct {
	*StorageClientOptions
	accountID   string
	accessKey   string
	address     string // host:port format
	grpcConfig  *GRPCConfig
	conn        *grpc.ClientConn
	protoClient proto.StorageServiceClient
	metadata    metadata.MD
}

// ParseGRPCURL parses a gRPC URL and returns the configuration
func ParseGRPCURL(grpcURL string) (*GRPCConfig, error) {
	// Parse the URL
	parsedURL, err := url.Parse(grpcURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	// Check if it's a valid gRPC scheme
	var isSecure bool
	var defaultPort int
	switch strings.ToLower(parsedURL.Scheme) {
	case "grpc":
		isSecure = false
		defaultPort = DefaultGRPCPort
	case "grpcs":
		isSecure = true
		defaultPort = DefaultGRPCSPort
	default:
		return nil, fmt.Errorf("invalid scheme: %s, expected 'grpc' or 'grpcs'", parsedURL.Scheme)
	}

	// Extract host and port
	host := parsedURL.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing hostname in URL")
	}

	var port int
	portStr := parsedURL.Port()
	if portStr == "" {
		// Use default port if not specified
		port = defaultPort
	} else {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %s", portStr)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("port number out of range: %d", port)
		}
	}

	return &GRPCConfig{
		IsSecure: isSecure,
		Host:     host,
		Port:     port,
		URL:      grpcURL,
	}, nil
}

// String returns a string representation of the config
func (c *GRPCConfig) String() string {
	secureStr := "insecure"
	if c.IsSecure {
		secureStr = "secure"
	}
	return fmt.Sprintf("Host: %s, Port: %d, Secure: %s, URL: %s",
		c.Host, c.Port, secureStr, c.URL)
}

// NewStorageClient creates a new StorageClient instance from a gRPC URL
// grpcURL is the full gRPC URL with scheme (e.g., "grpc://storage-server:50051" or "grpcs://storage.example.com:443")
// accountID is the customer GUID
// accessKey is the API access token
func NewStorageClient(grpcURL, accountID, accessKey string, opts ...StorageClientOption) (*StorageClient, error) {
	if grpcURL == "" {
		return nil, fmt.Errorf("gRPC URL cannot be empty")
	}

	// Parse the gRPC URL
	config, err := ParseGRPCURL(grpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gRPC URL: %w", err)
	}

	client := &StorageClient{
		StorageClientOptions: storageClientOptionsWithDefaults(opts),
		accountID:            accountID,
		accessKey:            accessKey,
		address:              config.URL,
		grpcConfig:           config,
	}

	// Create gRPC metadata with auth headers
	client.metadata = metadata.Pairs(
		backendv1.GrpcAccessKeyHeader, accessKey,
		backendv1.GrpcAccountKey, accountID,
	)

	return client, nil
}

// SetAccountID sets the customer account GUID
func (c *StorageClient) SetAccountID(value string) {
	c.accountID = value
	c.metadata = metadata.Pairs(
		backendv1.GrpcAccessKeyHeader, c.accessKey,
		backendv1.GrpcAccountKey, value,
	)
}

// SetAccessKey sets the API access key
func (c *StorageClient) SetAccessKey(value string) {
	c.accessKey = value
	c.metadata = metadata.Pairs(
		backendv1.GrpcAccessKeyHeader, value,
		backendv1.GrpcAccountKey, c.accountID,
	)
}

// GetAccountID returns the customer account GUID
func (c *StorageClient) GetAccountID() string {
	return c.accountID
}

// GetAccessKey returns the API access key
func (c *StorageClient) GetAccessKey() string {
	return c.accessKey
}

// GetAddress returns the storage server address
func (c *StorageClient) GetAddress() string {
	return c.address
}

// GetGRPCConfig returns the parsed gRPC configuration (if created from URL)
func (c *StorageClient) GetGRPCConfig() *GRPCConfig {
	return c.grpcConfig
}

// Connect establishes a gRPC connection to the storage server
func (c *StorageClient) Connect() error {
	if c.conn != nil {
		return fmt.Errorf("client is already connected")
	}

	// Build dial options
	var dialOpts []grpc.DialOption

	// Determine if connection should be secure
	if c.grpcConfig != nil && c.grpcConfig.IsSecure {
		// TODO: Add TLS credentials support for grpcs://
		return fmt.Errorf("TLS support not yet implemented, use grpc:// scheme for insecure connections")
	} else {
		// Use insecure credentials
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(c.address, dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to connect to storage server: %w", err)
	}
	c.conn = conn
	c.protoClient = proto.NewStorageServiceClient(conn)

	return nil
}

// Close closes the gRPC connection
func (c *StorageClient) Close() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	c.protoClient = nil
	return err
}

// IsConnected returns true if the client is connected to the server
func (c *StorageClient) IsConnected() bool {
	return c.conn != nil
}

// GetConnection returns the underlying gRPC connection (for advanced usage)
func (c *StorageClient) GetConnection() *grpc.ClientConn {
	return c.conn
}

// withMetadata returns a context with auth metadata attached
func (c *StorageClient) withMetadata(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, c.metadata)
}

// SendContainerProfile sends a container profile to the storage server
func (c *StorageClient) SendContainerProfile(ctx context.Context, profile *v1beta1.ContainerProfile, cluster string) (*proto.SendContainerProfileResponse, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	req := &proto.SendContainerProfileRequest{
		ContainerProfile: profile,
		CustomerGuid:     c.accountID,
		Cluster:          cluster,
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	return c.protoClient.SendContainerProfile(ctx, req)
}

// GetApplicationProfile retrieves an aggregated ApplicationProfile from the storage server
func (c *StorageClient) GetApplicationProfile(ctx context.Context, namespace, name, cluster string) (*v1beta1.ApplicationProfile, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	req := &proto.GetProfileRequest{
		Kind:         "applicationProfile",
		Namespace:    namespace,
		Name:         name,
		CustomerGuid: c.accountID,
		Cluster:      cluster,
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	resp, err := c.protoClient.GetProfile(ctx, req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get application profile: %s (code: %v)", resp.ErrorMessage, resp.ErrorCode)
	}

	return resp.ApplicationProfile, nil
}

// GetNetworkNeighborhood retrieves an aggregated NetworkNeighborhood from the storage server
func (c *StorageClient) GetNetworkNeighborhood(ctx context.Context, namespace, name, cluster string) (*v1beta1.NetworkNeighborhood, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	req := &proto.GetProfileRequest{
		Kind:         "networkNeighborhood",
		Namespace:    namespace,
		Name:         name,
		CustomerGuid: c.accountID,
		Cluster:      cluster,
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	resp, err := c.protoClient.GetProfile(ctx, req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get network neighborhood: %s (code: %v)", resp.ErrorMessage, resp.ErrorCode)
	}

	return resp.NetworkNeighborhood, nil
}

// ListApplicationProfiles lists all ApplicationProfiles in a namespace (returns metadata only, nil Spec)
func (c *StorageClient) ListApplicationProfiles(ctx context.Context, namespace, cluster string) (*v1beta1.ApplicationProfileList, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	req := &proto.ListApplicationProfilesRequest{
		Namespace:    namespace,
		CustomerGuid: c.accountID,
		Cluster:      cluster,
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	resp, err := c.protoClient.ListApplicationProfiles(ctx, req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to list application profiles: %s (code: %v)", resp.ErrorMessage, resp.ErrorCode)
	}

	// Convert pointer slice to value slice for ApplicationProfileList
	items := make([]v1beta1.ApplicationProfile, len(resp.ApplicationProfiles))
	for i, p := range resp.ApplicationProfiles {
		if p != nil {
			items[i] = *p
		}
	}

	list := &v1beta1.ApplicationProfileList{
		Items: items,
	}

	return list, nil
}

// ListNetworkNeighborhoods lists all NetworkNeighborhoods in a namespace (returns metadata only, nil Spec)
func (c *StorageClient) ListNetworkNeighborhoods(ctx context.Context, namespace, cluster string) (*v1beta1.NetworkNeighborhoodList, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	req := &proto.ListNetworkNeighborhoodsRequest{
		Namespace:    namespace,
		CustomerGuid: c.accountID,
		Cluster:      cluster,
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	resp, err := c.protoClient.ListNetworkNeighborhoods(ctx, req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to list network neighborhoods: %s (code: %v)", resp.ErrorMessage, resp.ErrorCode)
	}

	// Convert pointer slice to value slice for NetworkNeighborhoodList
	items := make([]v1beta1.NetworkNeighborhood, len(resp.NetworkNeighborhoods))
	for i, p := range resp.NetworkNeighborhoods {
		if p != nil {
			items[i] = *p
		}
	}

	list := &v1beta1.NetworkNeighborhoodList{
		Items: items,
	}

	return list, nil
}

// GetProtoClient returns the underlying proto client (for advanced usage)
func (c *StorageClient) GetProtoClient() proto.StorageServiceClient {
	return c.protoClient
}
