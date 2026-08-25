package v1

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	legacyv1beta1 "github.com/kubescape/backend/pkg/apis/softwarecomposition/v1beta1"
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

// SendContainerProfile sends a container profile to the storage server.
//
// Deprecated: use SendContainerProfileStream. The unary form is silently
// capped at gRPC's default 4 MiB message size; profiles with many or
// large entries can exceed this and fail on the wire. The streaming
// variant has no such bound.
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

// SendContainerProfileStream is the streaming replacement for
// SendContainerProfile. The profile is marshaled and sent to the server
// in chunks of sbomStreamChunkSize. Use this whenever you might write a
// large container profile (many entries, long stack traces, long paths).
func (c *StorageClient) SendContainerProfileStream(ctx context.Context, profile *v1beta1.ContainerProfile) (*proto.SendContainerProfileResponse, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}
	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	payload, err := profile.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ContainerProfile: %w", err)
	}

	// Note: we deliberately do NOT apply callTimeout here. Stream duration
	// depends on payload size / network speed; the per-call timeout was
	// chosen for short unary RPCs. Callers wanting a deadline should pass
	// a ctx with their own.
	ctx = c.withMetadata(ctx)

	stream, err := c.protoClient.SendContainerProfileStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open SendContainerProfileStream: %w", err)
	}

	// First chunk MUST set metadata. The metadata message is currently empty
	// (reserved for future per-call options); identifying fields continue to
	// be carried via gRPC metadata headers.
	cut := sbomStreamChunkSize
	if cut > len(payload) {
		cut = len(payload)
	}
	first := &proto.ContainerProfileChunk{
		Metadata:  &proto.ContainerProfileChunkMetadata{},
		BlobChunk: payload[:cut],
	}
	if err := stream.Send(first); err != nil {
		return nil, fmt.Errorf("failed to send first chunk: %w", err)
	}

	for offset := cut; offset < len(payload); offset += sbomStreamChunkSize {
		end := offset + sbomStreamChunkSize
		if end > len(payload) {
			end = len(payload)
		}
		if err := stream.Send(&proto.ContainerProfileChunk{BlobChunk: payload[offset:end]}); err != nil {
			return nil, fmt.Errorf("failed to send chunk: %w", err)
		}
	}

	return stream.CloseAndRecv()
}

// GetContainerProfileStream is the streaming replacement for the
// ContainerProfile branch of GetProfile. Use whenever the profile may
// exceed the default 4 MiB unary gRPC message limit. The chunks are
// reassembled and the marshaled bytes are unmarshaled internally — the
// caller receives a typed *v1beta1.ContainerProfile just as with the
// existing GetContainerProfile wrapper.
func (c *StorageClient) GetContainerProfileStream(ctx context.Context, namespace, name string, opts ...ProfileOption) (*v1beta1.ContainerProfile, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	profileOpts := profileOptionsWithDefaults(opts)

	req := &proto.GetContainerProfileStreamRequest{
		Namespace:              namespace,
		Name:                   name,
		Region:                 profileOpts.Region,
		CloudAccountIdentifier: profileOpts.CloudAccountIdentifier,
	}

	// Note: we deliberately do NOT apply callTimeout here. Stream duration
	// depends on payload size / network speed; the per-call timeout was
	// chosen for short unary RPCs. Callers wanting a deadline should pass
	// a ctx with their own.
	ctx = c.withMetadata(ctx)

	stream, err := c.protoClient.GetContainerProfileStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to open GetContainerProfileStream: %w", err)
	}

	// First chunk MUST carry metadata.
	firstChunk, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive first chunk: %w", err)
	}
	md := firstChunk.GetMetadata()
	if md == nil {
		return nil, fmt.Errorf("first GetContainerProfileStream chunk missing metadata")
	}
	if !md.Success {
		return nil, fmt.Errorf("failed to get container profile: %s (code: %v)", md.ErrorMessage, md.ErrorCode)
	}
	if !md.Exists {
		return nil, fmt.Errorf("container profile %s/%s not found", namespace, name)
	}

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
			return nil, fmt.Errorf("failed to receive chunk: %w", err)
		}
		buf = append(buf, chunk.BlobChunk...)
	}

	profile := &v1beta1.ContainerProfile{}
	if err := profile.Unmarshal(buf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ContainerProfile: %w", err)
	}
	return profile, nil
}

// GetApplicationProfile retrieves an aggregated ApplicationProfile from the storage server
// For backward compatibility, region and cloudAccountIdentifier can be provided via ProfileOption
// Old way: GetApplicationProfile(ctx, "ns", "name")
// New way: GetApplicationProfile(ctx, "ns", "name", WithProfileRegion("us-east-1"), WithProfileCloudAccountIdentifier("123"))
func (c *StorageClient) GetApplicationProfile(ctx context.Context, namespace, name string, opts ...ProfileOption) (*legacyv1beta1.ApplicationProfile, error) {
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
func (c *StorageClient) GetNetworkNeighborhood(ctx context.Context, namespace, name string, opts ...ProfileOption) (*legacyv1beta1.NetworkNeighborhood, error) {
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

// GetContainerProfile retrieves a ContainerProfile from the storage server.
//
// Deprecated: use GetContainerProfileStream. The unary form goes through
// GetProfile, which is capped at gRPC's default 4 MiB message size;
// profiles with many or large entries can exceed this and fail on the
// wire. The streaming variant has no such bound.
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
func (c *StorageClient) ListApplicationProfiles(ctx context.Context, namespace string, limit int64, cont string, opts ...ProfileOption) (*legacyv1beta1.ApplicationProfileList, error) {
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
	items := make([]legacyv1beta1.ApplicationProfile, len(resp.ApplicationProfiles))
	for i, p := range resp.ApplicationProfiles {
		if p != nil {
			items[i] = *p
		}
	}

	list := &legacyv1beta1.ApplicationProfileList{
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
func (c *StorageClient) ListNetworkNeighborhoods(ctx context.Context, namespace string, limit int64, cont string, opts ...ProfileOption) (*legacyv1beta1.NetworkNeighborhoodList, error) {
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
	items := make([]legacyv1beta1.NetworkNeighborhood, len(resp.NetworkNeighborhoods))
	for i, p := range resp.NetworkNeighborhoods {
		if p != nil {
			items[i] = *p
		}
	}

	list := &legacyv1beta1.NetworkNeighborhoodList{
		Items: items,
	}

	if resp.Cont != "" {
		list.Continue = resp.Cont
	}

	return list, nil
}

// PutSBOMStream uploads an SBOM identified by (image_digest, syft_version,
// source). The payload is the marshaled SBOMSyft proto, read from r and
// sent to the server in chunks of sbomStreamChunkSize. Callers with a
// typed *v1beta1.SBOMSyft should use MarshalSBOM to obtain r.
//
// The underlying RPC is client-streaming; the caller never needs to know
// the payload size in advance.
func (c *StorageClient) PutSBOMStream(ctx context.Context, imageDigest, syftVersion string, source proto.SBOMSource, r io.Reader) (*proto.PutSBOMResponse, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}
	if r == nil {
		return nil, fmt.Errorf("payload reader is nil")
	}

	// Note: we deliberately do NOT apply callTimeout here. Stream duration
	// depends on payload size / network speed; the per-call timeout was
	// chosen for short unary RPCs. Callers wanting a deadline should pass
	// a ctx with their own.
	ctx = c.withMetadata(ctx)

	stream, err := c.protoClient.PutSBOMStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open PutSBOMStream: %w", err)
	}

	// First chunk MUST set metadata, and MAY carry the first slice of bytes.
	buf := make([]byte, sbomStreamChunkSize)
	n, readErr := io.ReadFull(r, buf)
	if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read payload: %w", readErr)
	}
	first := &proto.PutSBOMChunk{
		Metadata: &proto.PutSBOMChunkMetadata{
			ImageDigest: imageDigest,
			SyftVersion: syftVersion,
			Source:      source,
		},
		BlobChunk: buf[:n],
	}
	if err := stream.Send(first); err != nil {
		return nil, fmt.Errorf("failed to send first chunk: %w", err)
	}

	// Subsequent chunks carry blob bytes only. ReadFull returns
	// io.ErrUnexpectedEOF on a short final read, which is normal — we send
	// whatever bytes we got and then stop on the next iteration.
	for readErr == nil {
		n, readErr = io.ReadFull(r, buf)
		if n > 0 {
			if err := stream.Send(&proto.PutSBOMChunk{BlobChunk: buf[:n]}); err != nil {
				return nil, fmt.Errorf("failed to send chunk: %w", err)
			}
		}
	}
	if readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read payload: %w", readErr)
	}

	return stream.CloseAndRecv()
}

// GetSBOMStream probes for or fetches an SBOM by (image_digest, syft_version).
//
// Contract:
//   - Server-reported failure → returns a non-nil error along with the
//     metadata. Caller's `if err != nil { ... }` is enough; no need to
//     inspect metadata.Success separately.
//   - Row does not exist OR metadataOnly is true → returns (md, nil, nil).
//     Caller consults `md.Exists` to distinguish probe-hit from miss. On a
//     hit, md.SbomMetadata carries the indexed view including the SBOM's
//     annotations (md.SbomMetadata.Annotations), so callers that only need
//     to branch on annotations can stay on the metadata_only path instead
//     of downloading and unmarshaling the full blob.
//   - Row exists and metadataOnly is false → returns (md, reader, nil).
//     The reader streams the marshaled SBOMSyft proto bytes; use
//     UnmarshalSBOM (or read into your own buffer) to reconstruct the
//     typed object. The caller MUST Close the reader to release the
//     underlying gRPC stream.
//
// The underlying RPC is server-streaming; the caller never needs to know
// the response size in advance.
func (c *StorageClient) GetSBOMStream(ctx context.Context, imageDigest, syftVersion string, metadataOnly bool) (*proto.GetSBOMChunkMetadata, io.ReadCloser, error) {
	if c.protoClient == nil {
		return nil, nil, fmt.Errorf("client is not connected")
	}

	req := &proto.GetSBOMRequest{
		ImageDigest:  imageDigest,
		SyftVersion:  syftVersion,
		MetadataOnly: metadataOnly,
	}

	ctx = c.withMetadata(ctx)

	// Note: we deliberately do NOT apply callTimeout here, because the
	// caller drives the read cadence via the returned io.ReadCloser. A
	// per-call timeout would clip large downloads. Callers wanting a
	// timeout should pass a ctx with their own deadline.
	streamCtx, cancel := context.WithCancel(ctx)
	stream, err := c.protoClient.GetSBOMStream(streamCtx, req)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to open GetSBOMStream: %w", err)
	}

	// First chunk MUST carry metadata.
	firstChunk, err := stream.Recv()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to receive first chunk: %w", err)
	}
	md := firstChunk.GetMetadata()
	if md == nil {
		cancel()
		return nil, nil, fmt.Errorf("first GetSBOMStream chunk missing metadata")
	}

	// A server-side failure is a real error — surface it instead of letting
	// callers misread (md, nil, nil) as a clean miss.
	if !md.Success {
		cancel()
		return md, nil, fmt.Errorf("server reported failure: %s (code: %v)", md.ErrorMessage, md.ErrorCode)
	}

	// On miss or metadata-only request, the server closes the stream after
	// the metadata chunk. Drain and return nil reader.
	if !md.Exists || metadataOnly {
		cancel()
		return md, nil, nil
	}

	reader := &sbomStreamReader{
		stream:  stream,
		pending: firstChunk.BlobChunk,
		cancel:  cancel,
	}
	return md, reader, nil
}

// sbomStreamReader adapts a server-streaming GetSBOM RPC to an
// io.ReadCloser, lazily fetching the next chunk when the previous one is
// drained.
type sbomStreamReader struct {
	stream  grpc.ServerStreamingClient[proto.GetSBOMChunk]
	pending []byte // buffer of bytes from the most-recent chunk not yet read
	cancel  context.CancelFunc
	err     error // sticky error; once set, Read returns it on every call
}

func (r *sbomStreamReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	for len(r.pending) == 0 {
		chunk, err := r.stream.Recv()
		if err == io.EOF {
			r.err = io.EOF
			return 0, io.EOF
		}
		if err != nil {
			r.err = err
			return 0, err
		}
		r.pending = chunk.BlobChunk
	}
	n := copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

func (r *sbomStreamReader) Close() error {
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	return nil
}

// PatchSBOMAnnotations updates only the annotations of an existing SBOM row,
// atomically and without rewriting the S3 blob. Merge-patch semantics: keys in
// `set` are added or overwritten, keys in `del` are removed, all other
// annotations are left untouched. Returns the post-merge SBOMMetadata on success.
func (c *StorageClient) PatchSBOMAnnotations(ctx context.Context, imageDigest, syftVersion string, set map[string]string, del []string) (*proto.SBOMMetadata, error) {
	if c.protoClient == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	req := &proto.PatchSBOMAnnotationsRequest{
		ImageDigest: imageDigest,
		SyftVersion: syftVersion,
		Set:         set,
		Delete:      del,
	}

	ctx = c.withMetadata(ctx)

	if c.callTimeout != nil && *c.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.callTimeout)
		defer cancel()
	}

	resp, err := c.protoClient.PatchSBOMAnnotations(ctx, req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to patch SBOM annotations: %s (code: %v)", resp.ErrorMessage, resp.ErrorCode)
	}

	return resp.SbomMetadata, nil
}

// MarshalSBOM marshals a typed SBOMSyft to its proto wire bytes and
// returns a reader over them, ready for PutSBOM. The full marshaled blob
// is held in memory; for very large SBOMs callers may prefer to write
// pre-marshaled bytes to a temp file and pass an *os.File to PutSBOM.
func MarshalSBOM(sbom *v1beta1.SBOMSyft) (io.Reader, error) {
	if sbom == nil {
		return nil, fmt.Errorf("sbom is nil")
	}
	b, err := sbom.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SBOMSyft: %w", err)
	}
	return bytes.NewReader(b), nil
}

// UnmarshalSBOM reads marshaled SBOMSyft proto bytes from r and returns
// the decoded object. It reads r until EOF; it does not Close r.
func UnmarshalSBOM(r io.Reader) (*v1beta1.SBOMSyft, error) {
	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM bytes: %w", err)
	}
	sbom := &v1beta1.SBOMSyft{}
	if err := sbom.Unmarshal(b); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SBOMSyft: %w", err)
	}
	return sbom, nil
}
