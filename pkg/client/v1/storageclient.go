package v1

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/kubescape/backend/pkg/client/v1/proto"
	backendv1 "github.com/kubescape/backend/pkg/server/v1"
	"github.com/kubescape/storage/pkg/apis/softwarecomposition/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// sbomStreamChunkSize is the per-chunk byte budget used by PutSBOMStream
// and GetSBOMStream. Set well below the default 4 MiB gRPC message limit
// to leave headroom for framing overhead.
const sbomStreamChunkSize = 1 << 20 // 1 MiB

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
	cluster     string
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
// cluster is the cluster name
// opts allow configuring optional parameters like hostType, hostID, timeout, etc.
func NewStorageClient(grpcURL, accountID, accessKey, cluster string, opts ...StorageClientOption) (*StorageClient, error) {
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
		cluster:              cluster,
		address:              fmt.Sprintf("%s:%d", config.Host, config.Port),
		grpcConfig:           config,
	}

	client.refreshMetadata()

	return client, nil
}

// SetAccountID sets the customer account GUID
func (c *StorageClient) SetAccountID(value string) {
	c.accountID = value
	c.refreshMetadata()
}

// SetAccessKey sets the API access key
func (c *StorageClient) SetAccessKey(value string) {
	c.accessKey = value
	c.refreshMetadata()
}

// SetCluster sets the cluster name
func (c *StorageClient) SetCluster(value string) {
	c.cluster = value
	c.refreshMetadata()
}

// refreshMetadata rebuilds the gRPC metadata with current credentials
func (c *StorageClient) refreshMetadata() {
	c.metadata = metadata.Pairs(
		backendv1.GrpcAccessKeyHeader, c.accessKey,
		backendv1.GrpcAccountKey, c.accountID,
		backendv1.GrpcClusterKey, c.cluster,
		backendv1.GrpcHostTypeKey, c.hostType,
		backendv1.GrpcHostIDKey, c.hostID,
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

// GetCluster returns the cluster name
func (c *StorageClient) GetCluster() string {
	return c.cluster
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
		// Use TLS with system CA certificates (for ingress-terminated TLS)
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
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

// withMetadata returns a context with auth metadata attached
func (c *StorageClient) withMetadata(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, c.metadata)
}

// SendContainerProfile sends a container profile to the storage server
func (c *StorageClient) SendContainerProfile(ctx context.Context, profile *v1beta1.ContainerProfile) (*proto.SendContainerProfileResponse, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	req := &proto.SendContainerProfileRequest{
		ContainerProfile: profile,
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
// For backward compatibility, region and cloudAccountIdentifier can be provided via ProfileOption
// Old way: GetApplicationProfile(ctx, "ns", "name")
// New way: GetApplicationProfile(ctx, "ns", "name", WithProfileRegion("us-east-1"), WithProfileCloudAccountIdentifier("123"))
func (c *StorageClient) GetApplicationProfile(ctx context.Context, namespace, name string, opts ...ProfileOption) (*v1beta1.ApplicationProfile, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	profileOpts := profileOptionsWithDefaults(opts)

	req := &proto.GetProfileRequest{
		Kind:                   "ApplicationProfile",
		Namespace:              namespace,
		Name:                   name,
		Region:                 profileOpts.Region,
		CloudAccountIdentifier: profileOpts.CloudAccountIdentifier,
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
// For backward compatibility, region and cloudAccountIdentifier can be provided via ProfileOption
// Old way: GetNetworkNeighborhood(ctx, "ns", "name")
// New way: GetNetworkNeighborhood(ctx, "ns", "name", WithProfileRegion("us-east-1"), WithProfileCloudAccountIdentifier("123"))
func (c *StorageClient) GetNetworkNeighborhood(ctx context.Context, namespace, name string, opts ...ProfileOption) (*v1beta1.NetworkNeighborhood, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	profileOpts := profileOptionsWithDefaults(opts)

	req := &proto.GetProfileRequest{
		Kind:                   "NetworkNeighborhood",
		Namespace:              namespace,
		Name:                   name,
		Region:                 profileOpts.Region,
		CloudAccountIdentifier: profileOpts.CloudAccountIdentifier,
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

// GetContainerProfile retrieves a ContainerProfile from the storage server
func (c *StorageClient) GetContainerProfile(ctx context.Context, namespace, name string, opts ...ProfileOption) (*v1beta1.ContainerProfile, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	profileOpts := profileOptionsWithDefaults(opts)

	req := &proto.GetProfileRequest{
		Kind:                   "ContainerProfile",
		Namespace:              namespace,
		Name:                   name,
		Region:                 profileOpts.Region,
		CloudAccountIdentifier: profileOpts.CloudAccountIdentifier,
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
		return nil, fmt.Errorf("failed to get container profile: %s (code: %v)", resp.ErrorMessage, resp.ErrorCode)
	}

	return resp.ContainerProfile, nil
}

// ListApplicationProfiles lists all ApplicationProfiles in a namespace (returns metadata only, nil Spec)
// For backward compatibility, region and cloudAccountIdentifier can be provided via ProfileOption
// Old way: ListApplicationProfiles(ctx, "ns", 100, "")
// New way: ListApplicationProfiles(ctx, "ns", 100, "", WithProfileRegion("us-east-1"), WithProfileCloudAccountIdentifier("123"))
func (c *StorageClient) ListApplicationProfiles(ctx context.Context, namespace string, limit int64, cont string, opts ...ProfileOption) (*v1beta1.ApplicationProfileList, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	profileOpts := profileOptionsWithDefaults(opts)

	req := &proto.ListApplicationProfilesRequest{
		Namespace:              namespace,
		Limit:                  limit,
		Cont:                   cont,
		Region:                 profileOpts.Region,
		CloudAccountIdentifier: profileOpts.CloudAccountIdentifier,
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

	// Set continue token for Kubernetes-style pagination
	if resp.Cont != "" {
		list.Continue = resp.Cont
	}

	return list, nil
}

// ListNetworkNeighborhoods lists all NetworkNeighborhoods in a namespace (returns metadata only, nil Spec)
// For backward compatibility, region and cloudAccountIdentifier can be provided via ProfileOption
// Old way: ListNetworkNeighborhoods(ctx, "ns", 100, "")
// New way: ListNetworkNeighborhoods(ctx, "ns", 100, "", WithProfileRegion("us-east-1"), WithProfileCloudAccountIdentifier("123"))
func (c *StorageClient) ListNetworkNeighborhoods(ctx context.Context, namespace string, limit int64, cont string, opts ...ProfileOption) (*v1beta1.NetworkNeighborhoodList, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	profileOpts := profileOptionsWithDefaults(opts)

	req := &proto.ListNetworkNeighborhoodsRequest{
		Namespace:              namespace,
		Limit:                  limit,
		Cont:                   cont,
		Region:                 profileOpts.Region,
		CloudAccountIdentifier: profileOpts.CloudAccountIdentifier,
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

	if resp.Cont != "" {
		list.Continue = resp.Cont
	}

	return list, nil
}

// PutSBOM uploads an SBOM to the storage server. Use for payloads small
// enough to fit in a single gRPC message (~4 MiB). For larger SBOMs, use
// PutSBOMStream instead.
func (c *StorageClient) PutSBOM(ctx context.Context, imageDigest, syftVersion string, source proto.SBOMSource, sbom *v1beta1.SBOMSyft) (*proto.PutSBOMResponse, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	req := &proto.PutSBOMRequest{
		ImageDigest: imageDigest,
		SyftVersion: syftVersion,
		Source:      source,
		Sbom:        sbom,
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	return c.protoClient.PutSBOM(ctx, req)
}

// PutSBOMStream uploads an SBOM to the storage server in chunks. The SBOM
// is marshaled to its proto wire form and split into ~1 MiB chunks; the
// first chunk carries metadata, subsequent chunks carry blob bytes only.
// Use when the SBOM is too large for unary PutSBOM (~4 MiB default gRPC
// message limit).
func (c *StorageClient) PutSBOMStream(ctx context.Context, imageDigest, syftVersion string, source proto.SBOMSource, sbom *v1beta1.SBOMSyft) (*proto.PutSBOMResponse, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	if sbom == nil {
		return nil, fmt.Errorf("sbom is nil")
	}

	payload, err := sbom.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SBOMSyft: %w", err)
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	stream, err := c.protoClient.PutSBOMStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open PutSBOMStream: %w", err)
	}

	// First chunk carries metadata + (optionally) the first slice of bytes.
	first := &proto.PutSBOMChunk{
		Metadata: &proto.PutSBOMChunkMetadata{
			ImageDigest: imageDigest,
			SyftVersion: syftVersion,
			Source:      source,
		},
	}
	cut := sbomStreamChunkSize
	if cut > len(payload) {
		cut = len(payload)
	}
	first.BlobChunk = payload[:cut]
	if err := stream.Send(first); err != nil {
		return nil, fmt.Errorf("failed to send first chunk: %w", err)
	}

	// Subsequent chunks carry blob bytes only.
	for offset := cut; offset < len(payload); offset += sbomStreamChunkSize {
		end := offset + sbomStreamChunkSize
		if end > len(payload) {
			end = len(payload)
		}
		if err := stream.Send(&proto.PutSBOMChunk{BlobChunk: payload[offset:end]}); err != nil {
			return nil, fmt.Errorf("failed to send chunk: %w", err)
		}
	}

	return stream.CloseAndRecv()
}

// GetSBOM probes for or fetches an SBOM by (image_digest, syft_version).
// Set metadataOnly=true to skip the blob and only check whether an SBOM
// exists. For SBOMs too large to fit in a single gRPC message, use
// GetSBOMStream when metadataOnly is false.
func (c *StorageClient) GetSBOM(ctx context.Context, imageDigest, syftVersion string, metadataOnly bool) (*proto.GetSBOMResponse, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	req := &proto.GetSBOMRequest{
		ImageDigest:  imageDigest,
		SyftVersion:  syftVersion,
		MetadataOnly: metadataOnly,
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	return c.protoClient.GetSBOM(ctx, req)
}

// GetSBOMStream fetches an SBOM in chunks. The first stream chunk returns
// metadata + status; subsequent chunks carry the marshaled SBOMSyft bytes,
// which are concatenated and unmarshaled by this method. Returns
// (metadata, nil) when the server reports exists=false or success=false —
// the caller should consult metadata.Success / metadata.Exists before
// using the returned SBOMSyft.
func (c *StorageClient) GetSBOMStream(ctx context.Context, imageDigest, syftVersion string) (*proto.GetSBOMChunkMetadata, *v1beta1.SBOMSyft, error) {
	if c.protoClient == nil {
		return nil, nil, fmt.Errorf("client is not connected")
	}

	req := &proto.GetSBOMRequest{
		ImageDigest: imageDigest,
		SyftVersion: syftVersion,
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	stream, err := c.protoClient.GetSBOMStream(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open GetSBOMStream: %w", err)
	}

	// First chunk MUST carry metadata.
	firstChunk, err := stream.Recv()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to receive first chunk: %w", err)
	}
	md := firstChunk.GetMetadata()
	if md == nil {
		return nil, nil, fmt.Errorf("first GetSBOMStream chunk missing metadata")
	}

	// On error or miss the server closes the stream after the metadata chunk.
	if !md.Success || !md.Exists {
		return md, nil, nil
	}

	// Accumulate blob bytes. The first chunk may have carried some payload
	// alongside the metadata.
	var buf []byte
	if len(firstChunk.BlobChunk) > 0 {
		buf = append(buf, firstChunk.BlobChunk...)
	}
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return md, nil, fmt.Errorf("failed to receive chunk: %w", err)
		}
		buf = append(buf, chunk.BlobChunk...)
	}

	sbom := &v1beta1.SBOMSyft{}
	if err := sbom.Unmarshal(buf); err != nil {
		return md, nil, fmt.Errorf("failed to unmarshal SBOMSyft: %w", err)
	}
	return md, sbom, nil
}
