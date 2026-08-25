package v1

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/armosec/armoapi-go/armotypes"
	legacyv1beta1 "github.com/kubescape/backend/pkg/apis/softwarecomposition/v1beta1"
	"github.com/kubescape/backend/pkg/client/v1/proto"
	"github.com/kubescape/storage/pkg/apis/softwarecomposition/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Mock StorageServiceClient for testing
type mockStorageServiceClient struct {
	sendContainerProfileFunc       func(ctx context.Context, in *proto.SendContainerProfileRequest, opts ...grpc.CallOption) (*proto.SendContainerProfileResponse, error)
	getProfileFunc                 func(ctx context.Context, in *proto.GetProfileRequest, opts ...grpc.CallOption) (*proto.GetProfileResponse, error)
	listApplicationProfilesFunc    func(ctx context.Context, in *proto.ListApplicationProfilesRequest, opts ...grpc.CallOption) (*proto.ListApplicationProfilesResponse, error)
	listNetworkNeighborhoodsFunc   func(ctx context.Context, in *proto.ListNetworkNeighborhoodsRequest, opts ...grpc.CallOption) (*proto.ListNetworkNeighborhoodsResponse, error)
	putSBOMStreamFunc              func(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.PutSBOMChunk, proto.PutSBOMResponse], error)
	getSBOMStreamFunc              func(ctx context.Context, in *proto.GetSBOMRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.GetSBOMChunk], error)
	sendContainerProfileStreamFunc func(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.ContainerProfileChunk, proto.SendContainerProfileResponse], error)
	getContainerProfileStreamFunc  func(ctx context.Context, in *proto.GetContainerProfileStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.GetContainerProfileStreamChunk], error)
	patchSBOMAnnotationsFunc       func(ctx context.Context, in *proto.PatchSBOMAnnotationsRequest, opts ...grpc.CallOption) (*proto.PatchSBOMAnnotationsResponse, error)
}

func (m *mockStorageServiceClient) SendContainerProfile(ctx context.Context, in *proto.SendContainerProfileRequest, opts ...grpc.CallOption) (*proto.SendContainerProfileResponse, error) {
	if m.sendContainerProfileFunc != nil {
		return m.sendContainerProfileFunc(ctx, in, opts...)
	}
	return &proto.SendContainerProfileResponse{Success: true}, nil
}

func (m *mockStorageServiceClient) GetProfile(ctx context.Context, in *proto.GetProfileRequest, opts ...grpc.CallOption) (*proto.GetProfileResponse, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, in, opts...)
	}
	return &proto.GetProfileResponse{Success: true}, nil
}

func (m *mockStorageServiceClient) ListApplicationProfiles(ctx context.Context, in *proto.ListApplicationProfilesRequest, opts ...grpc.CallOption) (*proto.ListApplicationProfilesResponse, error) {
	if m.listApplicationProfilesFunc != nil {
		return m.listApplicationProfilesFunc(ctx, in, opts...)
	}
	return &proto.ListApplicationProfilesResponse{Success: true}, nil
}

func (m *mockStorageServiceClient) ListNetworkNeighborhoods(ctx context.Context, in *proto.ListNetworkNeighborhoodsRequest, opts ...grpc.CallOption) (*proto.ListNetworkNeighborhoodsResponse, error) {
	if m.listNetworkNeighborhoodsFunc != nil {
		return m.listNetworkNeighborhoodsFunc(ctx, in, opts...)
	}
	return &proto.ListNetworkNeighborhoodsResponse{Success: true}, nil
}

func (m *mockStorageServiceClient) PutSBOMStream(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.PutSBOMChunk, proto.PutSBOMResponse], error) {
	if m.putSBOMStreamFunc != nil {
		return m.putSBOMStreamFunc(ctx, opts...)
	}
	return nil, fmt.Errorf("PutSBOMStream not implemented in mock")
}

func (m *mockStorageServiceClient) GetSBOMStream(ctx context.Context, in *proto.GetSBOMRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.GetSBOMChunk], error) {
	if m.getSBOMStreamFunc != nil {
		return m.getSBOMStreamFunc(ctx, in, opts...)
	}
	return nil, fmt.Errorf("GetSBOMStream not implemented in mock")
}

func (m *mockStorageServiceClient) PatchSBOMAnnotations(ctx context.Context, in *proto.PatchSBOMAnnotationsRequest, opts ...grpc.CallOption) (*proto.PatchSBOMAnnotationsResponse, error) {
	if m.patchSBOMAnnotationsFunc != nil {
		return m.patchSBOMAnnotationsFunc(ctx, in, opts...)
	}
	return &proto.PatchSBOMAnnotationsResponse{Success: true}, nil
}

func (m *mockStorageServiceClient) SendContainerProfileStream(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.ContainerProfileChunk, proto.SendContainerProfileResponse], error) {
	if m.sendContainerProfileStreamFunc != nil {
		return m.sendContainerProfileStreamFunc(ctx, opts...)
	}
	return nil, fmt.Errorf("SendContainerProfileStream not implemented in mock")
}

func (m *mockStorageServiceClient) GetContainerProfileStream(ctx context.Context, in *proto.GetContainerProfileStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.GetContainerProfileStreamChunk], error) {
	if m.getContainerProfileStreamFunc != nil {
		return m.getContainerProfileStreamFunc(ctx, in, opts...)
	}
	return nil, fmt.Errorf("GetContainerProfileStream not implemented in mock")
}

func TestNewStorageClient(t *testing.T) {
	tests := []struct {
		name        string
		grpcURL     string
		accountID   string
		accessKey   string
		cluster     string
		hostType    string
		hostID      string
		expectError bool
	}{
		{
			name:        "valid grpc URL",
			grpcURL:     "grpc://storage.example.com:50051",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "test-cluster",
			hostType:    "",
			hostID:      "",
			expectError: false,
		},
		{
			name:        "valid grpcs URL",
			grpcURL:     "grpcs://storage.example.com:443",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "test-cluster",
			hostType:    "",
			hostID:      "",
			expectError: false,
		},
		{
			name:        "empty URL",
			grpcURL:     "",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "test-cluster",
			hostType:    "",
			hostID:      "",
			expectError: true,
		},
		{
			name:        "invalid scheme",
			grpcURL:     "http://storage.example.com:50051",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "test-cluster",
			hostType:    "",
			hostID:      "",
			expectError: true,
		},
		{
			name:        "kubernetes host type with cluster",
			grpcURL:     "grpc://storage.example.com:50051",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "my-k8s-cluster",
			hostType:    string(armotypes.HostTypeKubernetes),
			hostID:      "",
			expectError: false,
		},
		{
			name:        "ec2 host type with hostID",
			grpcURL:     "grpc://storage.example.com:50051",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "",
			hostType:    string(armotypes.HostTypeEc2),
			hostID:      "i-0123456789abcdef0",
			expectError: false,
		},
		{
			name:        "ecs-ec2 host type with hostID",
			grpcURL:     "grpc://storage.example.com:50051",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "my-ecs-cluster",
			hostType:    string(armotypes.HostTypeEcsEc2),
			hostID:      "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []StorageClientOption
			if tt.hostType != "" {
				opts = append(opts, WithHostType(tt.hostType))
			}
			if tt.hostID != "" {
				opts = append(opts, WithHostID(tt.hostID))
			}

			client, err := NewStorageClient(tt.grpcURL, tt.accountID, tt.accessKey, tt.cluster, opts...)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.accountID, client.GetAccountID())
				assert.Equal(t, tt.accessKey, client.GetAccessKey())
				assert.Equal(t, tt.cluster, client.GetCluster())
				assert.Equal(t, tt.hostType, client.hostType)
				assert.Equal(t, tt.hostID, client.hostID)
				assert.NotNil(t, client.GetGRPCConfig())
			}
		})
	}
}

func TestStorageClient_SetAccountIDAndAccessKey(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "account1", "key1", "cluster1", WithHostType(string(armotypes.HostTypeKubernetes)))
	require.NoError(t, err)

	assert.Equal(t, "account1", client.GetAccountID())
	assert.Equal(t, "key1", client.GetAccessKey())
	assert.Equal(t, "cluster1", client.GetCluster())
	assert.Equal(t, string(armotypes.HostTypeKubernetes), client.hostType)
	assert.Equal(t, "", client.hostID)

	client.SetAccountID("account2")
	assert.Equal(t, "account2", client.GetAccountID())

	client.SetAccessKey("key2")
	assert.Equal(t, "key2", client.GetAccessKey())

	client.SetCluster("cluster2")
	assert.Equal(t, "cluster2", client.GetCluster())
}

func TestStorageClient_SendContainerProfile(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster", WithHostType(string(armotypes.HostTypeKubernetes)))
	require.NoError(t, err)

	mockClient := &mockStorageServiceClient{
		sendContainerProfileFunc: func(ctx context.Context, in *proto.SendContainerProfileRequest, opts ...grpc.CallOption) (*proto.SendContainerProfileResponse, error) {
			assert.NotNil(t, in.ContainerProfile)
			return &proto.SendContainerProfileResponse{Success: true}, nil
		},
	}
	client.protoClient = mockClient

	resp, err := client.SendContainerProfile(context.Background(), &v1beta1.ContainerProfile{})
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestStorageClient_GetProfile(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	tests := []struct {
		name                   string
		kind                   armotypes.ProfileKind
		namespace              string
		profileName            string
		region                 string
		cloudAccountIdentifier string
	}{
		{
			name:                   "ApplicationProfile without region and cloudAccountIdentifier (k8s)",
			kind:                   armotypes.ApplicationProfileKind,
			namespace:              "default",
			profileName:            "my-app",
			region:                 "",
			cloudAccountIdentifier: "",
		},
		{
			name:                   "ApplicationProfile with region and cloudAccountIdentifier (ECS/EC2)",
			kind:                   armotypes.ApplicationProfileKind,
			namespace:              "",
			profileName:            "ecs-task-profile",
			region:                 "us-east-1",
			cloudAccountIdentifier: "123456789012",
		},
		{
			name:                   "NetworkNeighborhood without region and cloudAccountIdentifier (k8s)",
			kind:                   armotypes.NetworkNeighborhoodKind,
			namespace:              "kube-system",
			profileName:            "core-dns-nn",
			region:                 "",
			cloudAccountIdentifier: "",
		},
		{
			name:                   "NetworkNeighborhood with region and cloudAccountIdentifier (ECS/EC2)",
			kind:                   armotypes.NetworkNeighborhoodKind,
			namespace:              "",
			profileName:            "ec2-instance-nn",
			region:                 "us-west-2",
			cloudAccountIdentifier: "987654321098",
		},
		{
			name:                   "ContainerProfile k8s",
			kind:                   armotypes.ContainerProfileKind,
			namespace:              "default",
			profileName:            "my-container",
			region:                 "",
			cloudAccountIdentifier: "",
		},
		{
			name:                   "ContainerProfile ECS/EC2",
			kind:                   armotypes.ContainerProfileKind,
			namespace:              "",
			profileName:            "ecs-container",
			region:                 "us-east-1",
			cloudAccountIdentifier: "123456789012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			mockClient := &mockStorageServiceClient{
				getProfileFunc: func(ctx context.Context, in *proto.GetProfileRequest, opts ...grpc.CallOption) (*proto.GetProfileResponse, error) {
					assert.Equal(t, string(tt.kind), in.Kind)
					assert.Equal(t, tt.namespace, in.Namespace)
					assert.Equal(t, tt.profileName, in.Name)
					assert.Equal(t, tt.region, in.Region)
					assert.Equal(t, tt.cloudAccountIdentifier, in.CloudAccountIdentifier)
					switch tt.kind {
					case armotypes.ApplicationProfileKind:
						return &proto.GetProfileResponse{
							Success:            true,
							ApplicationProfile: &legacyv1beta1.ApplicationProfile{},
						}, nil
					case armotypes.ContainerProfileKind:
						return &proto.GetProfileResponse{
							Success:          true,
							ContainerProfile: &v1beta1.ContainerProfile{},
						}, nil
					default:
						return &proto.GetProfileResponse{
							Success:             true,
							NetworkNeighborhood: &legacyv1beta1.NetworkNeighborhood{},
						}, nil
					}
				},
			}
			client.protoClient = mockClient

			switch tt.kind {
			case armotypes.ApplicationProfileKind:
				resp, err := client.GetApplicationProfile(context.Background(), tt.namespace, tt.profileName, WithProfileRegion(tt.region), WithProfileCloudAccountIdentifier(tt.cloudAccountIdentifier))
				require.NoError(t, err)
				assert.NotNil(t, resp)
			case armotypes.ContainerProfileKind:
				resp, err := client.GetContainerProfile(context.Background(), tt.namespace, tt.profileName, WithProfileRegion(tt.region), WithProfileCloudAccountIdentifier(tt.cloudAccountIdentifier))
				require.NoError(t, err)
				assert.NotNil(t, resp)
			default:
				resp, err := client.GetNetworkNeighborhood(context.Background(), tt.namespace, tt.profileName, WithProfileRegion(tt.region), WithProfileCloudAccountIdentifier(tt.cloudAccountIdentifier))
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestStorageClient_NotConnected(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster", WithHostType(string(armotypes.HostTypeKubernetes)))
	require.NoError(t, err)

	resp, err := client.SendContainerProfile(context.Background(), &v1beta1.ContainerProfile{})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "not connected")
}

func TestStorageClient_ListApplicationProfiles(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	tests := []struct {
		name                   string
		namespace              string
		limit                  int64
		cont                   string
		region                 string
		cloudAccountIdentifier string
		expectedLen            int
	}{
		{
			name:                   "k8s with namespace, no region/cloudAccountIdentifier",
			namespace:              "default",
			limit:                  10,
			cont:                   "next-token",
			region:                 "",
			cloudAccountIdentifier: "",
			expectedLen:            2,
		},
		{
			name:                   "ECS/EC2 with region and cloudAccountIdentifier, empty namespace",
			namespace:              "",
			limit:                  25,
			cont:                   "cont-token",
			region:                 "us-east-1",
			cloudAccountIdentifier: "123456789012",
			expectedLen:            3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			mockClient := &mockStorageServiceClient{
				listApplicationProfilesFunc: func(ctx context.Context, in *proto.ListApplicationProfilesRequest, opts ...grpc.CallOption) (*proto.ListApplicationProfilesResponse, error) {
					assert.Equal(t, tt.namespace, in.Namespace)
					assert.Equal(t, tt.limit, in.Limit)
					assert.Equal(t, tt.cont, in.Cont)
					assert.Equal(t, tt.region, in.Region)
					assert.Equal(t, tt.cloudAccountIdentifier, in.CloudAccountIdentifier)
					profiles := make([]*legacyv1beta1.ApplicationProfile, tt.expectedLen)
					for i := range profiles {
						profiles[i] = &legacyv1beta1.ApplicationProfile{}
					}
					return &proto.ListApplicationProfilesResponse{
						Success:             true,
						ApplicationProfiles: profiles,
					}, nil
				},
			}
			client.protoClient = mockClient

			list, err := client.ListApplicationProfiles(context.Background(), tt.namespace, tt.limit, tt.cont, WithProfileRegion(tt.region), WithProfileCloudAccountIdentifier(tt.cloudAccountIdentifier))
			require.NoError(t, err)
			assert.NotNil(t, list)
			assert.Len(t, list.Items, tt.expectedLen)
		})
	}
}

func TestStorageClient_ListNetworkNeighborhoods(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	tests := []struct {
		name                   string
		namespace              string
		limit                  int64
		cont                   string
		region                 string
		cloudAccountIdentifier string
		expectedLen            int
	}{
		{
			name:                   "k8s with namespace, no region/cloudAccountIdentifier",
			namespace:              "kube-system",
			limit:                  25,
			cont:                   "cont-token",
			region:                 "",
			cloudAccountIdentifier: "",
			expectedLen:            3,
		},
		{
			name:                   "ECS/EC2 with region and cloudAccountIdentifier, empty namespace",
			namespace:              "",
			limit:                  50,
			cont:                   "next-page",
			region:                 "us-west-2",
			cloudAccountIdentifier: "987654321098",
			expectedLen:            2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			mockClient := &mockStorageServiceClient{
				listNetworkNeighborhoodsFunc: func(ctx context.Context, in *proto.ListNetworkNeighborhoodsRequest, opts ...grpc.CallOption) (*proto.ListNetworkNeighborhoodsResponse, error) {
					assert.Equal(t, tt.namespace, in.Namespace)
					assert.Equal(t, tt.limit, in.Limit)
					assert.Equal(t, tt.cont, in.Cont)
					assert.Equal(t, tt.region, in.Region)
					assert.Equal(t, tt.cloudAccountIdentifier, in.CloudAccountIdentifier)
					neighborhoods := make([]*legacyv1beta1.NetworkNeighborhood, tt.expectedLen)
					for i := range neighborhoods {
						neighborhoods[i] = &legacyv1beta1.NetworkNeighborhood{}
					}
					return &proto.ListNetworkNeighborhoodsResponse{
						Success:              true,
						NetworkNeighborhoods: neighborhoods,
					}, nil
				},
			}
			client.protoClient = mockClient

			list, err := client.ListNetworkNeighborhoods(context.Background(), tt.namespace, tt.limit, tt.cont, WithProfileRegion(tt.region), WithProfileCloudAccountIdentifier(tt.cloudAccountIdentifier))
			require.NoError(t, err)
			assert.NotNil(t, list)
			assert.Len(t, list.Items, tt.expectedLen)
		})
	}
}

func TestStorageClient_ListApplicationProfiles_NotConnected(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	list, err := client.ListApplicationProfiles(context.Background(), "default", 0, "")
	assert.Error(t, err)
	assert.Nil(t, list)
	assert.Contains(t, err.Error(), "not connected")
}

func TestStorageClient_ListNetworkNeighborhoods_NotConnected(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	list, err := client.ListNetworkNeighborhoods(context.Background(), "default", 0, "")
	assert.Error(t, err)
	assert.Nil(t, list)
	assert.Contains(t, err.Error(), "not connected")
}

func TestStorageClientOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := storageClientOptionsWithDefaults(nil)
		assert.NotNil(t, opts.callTimeout)
		assert.Equal(t, 30*time.Second, *opts.callTimeout)
		assert.False(t, opts.withTrace)
		assert.Empty(t, opts.hostType)
		assert.Empty(t, opts.hostID)
	})

	t.Run("with custom timeout", func(t *testing.T) {
		opts := storageClientOptionsWithDefaults([]StorageClientOption{
			WithCallTimeout(60 * time.Second),
		})
		assert.NotNil(t, opts.callTimeout)
		assert.Equal(t, 60*time.Second, *opts.callTimeout)
	})

	t.Run("with trace enabled", func(t *testing.T) {
		opts := storageClientOptionsWithDefaults([]StorageClientOption{
			WithStorageTrace(true),
		})
		assert.True(t, opts.withTrace)
	})

	t.Run("with host type and host ID", func(t *testing.T) {
		opts := storageClientOptionsWithDefaults([]StorageClientOption{
			WithHostType("ecs"),
			WithHostID("i-0123456789abcdef0"),
		})
		assert.Equal(t, "ecs", opts.hostType)
		assert.Equal(t, "i-0123456789abcdef0", opts.hostID)
	})

	t.Run("with all options", func(t *testing.T) {
		opts := storageClientOptionsWithDefaults([]StorageClientOption{
			WithCallTimeout(45 * time.Second),
			WithStorageTrace(true),
			WithHostType("ec2"),
			WithHostID("i-fedcba9876543210"),
		})
		assert.Equal(t, 45*time.Second, *opts.callTimeout)
		assert.True(t, opts.withTrace)
		assert.Equal(t, "ec2", opts.hostType)
		assert.Equal(t, "i-fedcba9876543210", opts.hostID)
	})
}

func TestProfileOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := profileOptionsWithDefaults(nil)
		assert.Empty(t, opts.Region)
		assert.Empty(t, opts.CloudAccountIdentifier)
	})

	t.Run("with region", func(t *testing.T) {
		opts := profileOptionsWithDefaults([]ProfileOption{
			WithProfileRegion("us-east-1"),
		})
		assert.Equal(t, "us-east-1", opts.Region)
		assert.Empty(t, opts.CloudAccountIdentifier)
	})

	t.Run("with cloud account identifier", func(t *testing.T) {
		opts := profileOptionsWithDefaults([]ProfileOption{
			WithProfileCloudAccountIdentifier("123456789012"),
		})
		assert.Empty(t, opts.Region)
		assert.Equal(t, "123456789012", opts.CloudAccountIdentifier)
	})

	t.Run("with both region and cloud account identifier", func(t *testing.T) {
		opts := profileOptionsWithDefaults([]ProfileOption{
			WithProfileRegion("eu-west-1"),
			WithProfileCloudAccountIdentifier("987654321098"),
		})
		assert.Equal(t, "eu-west-1", opts.Region)
		assert.Equal(t, "987654321098", opts.CloudAccountIdentifier)
	})
}

func TestParseGRPCURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		expectedCfg *GRPCConfig
	}{
		{
			name:        "grpc with port",
			url:         "grpc://aaa:123",
			expectError: false,
			expectedCfg: &GRPCConfig{IsSecure: false, Host: "aaa", Port: 123},
		},
		{
			name:        "grpcs with port",
			url:         "grpcs://aaaa:1234",
			expectError: false,
			expectedCfg: &GRPCConfig{IsSecure: true, Host: "aaaa", Port: 1234},
		},
		{
			name:        "grpc without port defaults to 50051",
			url:         "grpc://example.com",
			expectError: false,
			expectedCfg: &GRPCConfig{IsSecure: false, Host: "example.com", Port: 50051},
		},
		{
			name:        "grpcs without port defaults to 50052",
			url:         "grpcs://secure.example.com",
			expectError: false,
			expectedCfg: &GRPCConfig{IsSecure: true, Host: "secure.example.com", Port: 50052},
		},
		{
			name:        "invalid scheme",
			url:         "http://test:123",
			expectError: true,
		},
		{
			name:        "invalid port",
			url:         "grpc://test:invalid",
			expectError: true,
		},
		{
			name:        "missing hostname",
			url:         "grpc://",
			expectError: true,
		},
		{
			name:        "port out of range",
			url:         "grpc://test:99999",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseGRPCURL(tt.url)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.expectedCfg.IsSecure, config.IsSecure)
				assert.Equal(t, tt.expectedCfg.Host, config.Host)
				assert.Equal(t, tt.expectedCfg.Port, config.Port)
				assert.Equal(t, tt.url, config.URL)
			}
		})
	}
}

// storageRoundTripServer is a minimal proto.StorageServiceServer impl that
// records what PutSBOMStream receives and serves what GetSBOMStream should return.
// It exercises the full marshal/unmarshal path of the streaming RPCs
// end-to-end through bufconn — catching wire-shape bugs the prior
// mock-based tests would miss.
type storageRoundTripServer struct {
	proto.UnimplementedStorageServiceServer
	mu sync.Mutex

	// SBOM upload state: captured from the most recent PutSBOMStream call.
	receivedMetadata *proto.PutSBOMChunkMetadata
	receivedBytes    []byte

	// SBOM download config: returned from the next GetSBOMStream call.
	// If serveExists is false the server closes the stream after the
	// metadata chunk.
	serveExists     bool
	serveMetadata   *proto.SBOMMetadata
	serveBytes      []byte
	serveChunkSize  int           // 0 → send the whole payload in one chunk
	serveChunkDelay time.Duration // sleep before each chunk; for timeout tests

	// ContainerProfile upload state: captured from the most recent
	// SendContainerProfileStream call.
	cpReceivedBytes []byte

	// ContainerProfile download config: returned from the next
	// GetContainerProfileStream call.
	cpServeExists    bool
	cpServeBytes     []byte
	cpServeChunkSize int

	// PatchSBOMAnnotations state and config
	receivedPatchReq *proto.PatchSBOMAnnotationsRequest
	patchResp        *proto.PatchSBOMAnnotationsResponse
	patchErr         error
	patchDelay       time.Duration
}

func (s *storageRoundTripServer) PutSBOMStream(stream grpc.ClientStreamingServer[proto.PutSBOMChunk, proto.PutSBOMResponse]) error {
	var buf []byte
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		s.mu.Lock()
		if chunk.Metadata != nil && s.receivedMetadata == nil {
			s.receivedMetadata = chunk.Metadata
		}
		s.mu.Unlock()
		buf = append(buf, chunk.BlobChunk...)
	}
	s.mu.Lock()
	s.receivedBytes = buf
	s.mu.Unlock()
	return stream.SendAndClose(&proto.PutSBOMResponse{Success: true})
}

func (s *storageRoundTripServer) GetSBOMStream(req *proto.GetSBOMRequest, stream grpc.ServerStreamingServer[proto.GetSBOMChunk]) error {
	s.mu.Lock()
	exists := s.serveExists
	metadata := s.serveMetadata
	payload := s.serveBytes
	chunkSize := s.serveChunkSize
	chunkDelay := s.serveChunkDelay
	s.mu.Unlock()

	// First chunk MUST carry the metadata header.
	first := &proto.GetSBOMChunk{
		Metadata: &proto.GetSBOMChunkMetadata{
			Success:      true,
			Exists:       exists,
			SbomMetadata: metadata,
		},
	}
	if err := stream.Send(first); err != nil {
		return err
	}

	if !exists || req.MetadataOnly {
		return nil
	}

	if chunkSize <= 0 || chunkSize >= len(payload) {
		if chunkDelay > 0 {
			time.Sleep(chunkDelay)
		}
		return stream.Send(&proto.GetSBOMChunk{BlobChunk: payload})
	}
	for offset := 0; offset < len(payload); offset += chunkSize {
		end := offset + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		if chunkDelay > 0 {
			time.Sleep(chunkDelay)
		}
		if err := stream.Send(&proto.GetSBOMChunk{BlobChunk: payload[offset:end]}); err != nil {
			return err
		}
	}
	return nil
}

func (s *storageRoundTripServer) SendContainerProfileStream(stream grpc.ClientStreamingServer[proto.ContainerProfileChunk, proto.SendContainerProfileResponse]) error {
	var buf []byte
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		buf = append(buf, chunk.BlobChunk...)
	}
	s.mu.Lock()
	s.cpReceivedBytes = buf
	s.mu.Unlock()
	return stream.SendAndClose(&proto.SendContainerProfileResponse{Success: true})
}

func (s *storageRoundTripServer) GetContainerProfileStream(req *proto.GetContainerProfileStreamRequest, stream grpc.ServerStreamingServer[proto.GetContainerProfileStreamChunk]) error {
	s.mu.Lock()
	exists := s.cpServeExists
	payload := s.cpServeBytes
	chunkSize := s.cpServeChunkSize
	s.mu.Unlock()

	first := &proto.GetContainerProfileStreamChunk{
		Metadata: &proto.GetContainerProfileStreamChunkMetadata{
			Success: true,
			Exists:  exists,
		},
	}
	if err := stream.Send(first); err != nil {
		return err
	}
	if !exists {
		return nil
	}
	if chunkSize <= 0 || chunkSize >= len(payload) {
		return stream.Send(&proto.GetContainerProfileStreamChunk{BlobChunk: payload})
	}
	for offset := 0; offset < len(payload); offset += chunkSize {
		end := offset + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		if err := stream.Send(&proto.GetContainerProfileStreamChunk{BlobChunk: payload[offset:end]}); err != nil {
			return err
		}
	}
	return nil
}

func (s *storageRoundTripServer) PatchSBOMAnnotations(ctx context.Context, req *proto.PatchSBOMAnnotationsRequest) (*proto.PatchSBOMAnnotationsResponse, error) {
	s.mu.Lock()
	s.receivedPatchReq = req
	resp := s.patchResp
	err := s.patchErr
	delay := s.patchDelay
	s.mu.Unlock()

	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}

	mergedAnnotations := make(map[string]string)
	if s.serveMetadata != nil && s.serveMetadata.Annotations != nil {
		for k, v := range s.serveMetadata.Annotations {
			mergedAnnotations[k] = v
		}
	}
	for k, v := range req.Set {
		mergedAnnotations[k] = v
	}
	for _, k := range req.Delete {
		delete(mergedAnnotations, k)
	}

	return &proto.PatchSBOMAnnotationsResponse{
		Success: true,
		SbomMetadata: &proto.SBOMMetadata{
			ImageDigest: req.ImageDigest,
			SyftVersion: req.SyftVersion,
			Annotations: mergedAnnotations,
		},
	}, nil
}

// startBufconnStorageServer starts the round-trip server on an in-memory
// bufconn listener and returns a client connected to it. Optional client
// options (e.g. WithCallTimeout) are forwarded to NewStorageClient.
func startBufconnStorageServer(t *testing.T, opts ...StorageClientOption) (*storageRoundTripServer, *StorageClient, func()) {
	t.Helper()
	const bufsize = 1024 * 1024
	lis := bufconn.Listen(bufsize)
	srv := grpc.NewServer()
	rtSrv := &storageRoundTripServer{}
	proto.RegisterStorageServiceServer(srv, rtSrv)
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client, err := NewStorageClient("grpc://example.com:50051", "test-account", "test-key", "test-cluster", opts...)
	require.NoError(t, err)
	client.conn = conn
	client.protoClient = proto.NewStorageServiceClient(conn)

	cleanup := func() {
		_ = conn.Close()
		srv.Stop()
	}
	return rtSrv, client, cleanup
}

// sampleSBOMSyft constructs a non-trivial SBOMSyft for round-trip tests
// so the proto Marshal / Unmarshal path actually has fields to round-trip.
func sampleSBOMSyft() *v1beta1.SBOMSyft {
	return &v1beta1.SBOMSyft{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SBOMSyft",
			APIVersion: "spdx.softwarecomposition.kubescape.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sha256-abc123",
			Namespace: "kubescape",
			Annotations: map[string]string{
				"image.name":   "library/nginx:latest",
				"syft.version": "1.0.0",
			},
		},
		Spec: v1beta1.SBOMSyftSpec{
			Metadata: v1beta1.SPDXMeta{
				Tool: v1beta1.ToolMeta{
					Name:    "syft",
					Version: "1.0.0",
				},
			},
			Syft: v1beta1.SyftDocument{
				SyftSource: v1beta1.SyftSource{Type: "image"},
			},
		},
	}
}

// TestStorageClient_SBOMRoundTrip is the end-to-end marshal/unmarshal test
// matthyx asked for: it stands up a real gRPC server over bufconn, ships
// a non-trivial SBOMSyft via the client's PutSBOM, then pulls it back via
// GetSBOM, and verifies semantic equality at every hop. Covers four
// shapes:
//   - Probe (metadata_only=true), exists path
//   - Probe (metadata_only=true), miss path
//   - Full fetch, payload fits in one chunk
//   - Full fetch, payload is split across many small chunks (forces the
//     reader to drain across multiple Recv calls)
func TestStorageClient_SBOMRoundTrip(t *testing.T) {
	const (
		imageDigest = "abc123def456"
		syftVersion = "1.0.0"
	)
	source := proto.SBOMSource_SBOM_SOURCE_WORKLOAD

	t.Run("PutSBOM marshals through MarshalSBOM and arrives intact", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		original := sampleSBOMSyft()
		reader, err := MarshalSBOM(original)
		require.NoError(t, err)

		resp, err := client.PutSBOMStream(context.Background(), imageDigest, syftVersion, source, reader)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		// Server received the right metadata header.
		require.NotNil(t, rtSrv.receivedMetadata)
		assert.Equal(t, imageDigest, rtSrv.receivedMetadata.ImageDigest)
		assert.Equal(t, syftVersion, rtSrv.receivedMetadata.SyftVersion)
		assert.Equal(t, source, rtSrv.receivedMetadata.Source)

		// Server-received bytes unmarshal to a SBOMSyft equal to the original.
		require.NotEmpty(t, rtSrv.receivedBytes)
		got := &v1beta1.SBOMSyft{}
		require.NoError(t, got.Unmarshal(rtSrv.receivedBytes))
		assert.Equal(t, original.Name, got.Name)
		assert.Equal(t, original.Namespace, got.Namespace)
		assert.Equal(t, original.Spec.Metadata.Tool.Name, got.Spec.Metadata.Tool.Name)
		assert.Equal(t, original.Spec.Metadata.Tool.Version, got.Spec.Metadata.Tool.Version)
		assert.Equal(t, original.Annotations["image.name"], got.Annotations["image.name"])
	})

	t.Run("GetSBOM metadata-only probe returns only metadata", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		rtSrv.serveExists = true
		rtSrv.serveMetadata = &proto.SBOMMetadata{
			ImageDigest: imageDigest,
			SyftVersion: syftVersion,
			// Annotations let agents branch (on status, tool-version,
			// node-name, scanner-memory-limit, ...) without downloading the
			// full blob; they MUST survive the metadata_only path.
			Annotations: map[string]string{
				"kubescape.io/status":               "completed",
				"kubescape.io/tool-version":         syftVersion,
				"kubescape.io/node-name":            "node-1",
				"kubescape.io/scanner-memory-limit": "1500Mi",
			},
		}

		md, reader, err := client.GetSBOMStream(context.Background(), imageDigest, syftVersion, true)
		require.NoError(t, err)
		assert.True(t, md.Success)
		assert.True(t, md.Exists)
		require.NotNil(t, md.SbomMetadata)
		assert.Equal(t, imageDigest, md.SbomMetadata.ImageDigest)
		assert.Equal(t, rtSrv.serveMetadata.Annotations, md.SbomMetadata.Annotations,
			"annotations must round-trip through the metadata_only path")
		assert.Nil(t, reader, "metadata-only probe must not return a reader")
	})

	t.Run("GetSBOM probe miss returns exists=false and no reader", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		rtSrv.serveExists = false

		md, reader, err := client.GetSBOMStream(context.Background(), imageDigest, syftVersion, true)
		require.NoError(t, err)
		assert.True(t, md.Success)
		assert.False(t, md.Exists)
		assert.Nil(t, reader)
	})

	t.Run("GetSBOM full fetch round-trips through UnmarshalSBOM (single chunk)", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		original := sampleSBOMSyft()
		payload, err := original.Marshal()
		require.NoError(t, err)
		rtSrv.serveExists = true
		rtSrv.serveMetadata = &proto.SBOMMetadata{ImageDigest: imageDigest, SyftVersion: syftVersion}
		rtSrv.serveBytes = payload
		rtSrv.serveChunkSize = 0 // single chunk

		md, reader, err := client.GetSBOMStream(context.Background(), imageDigest, syftVersion, false)
		require.NoError(t, err)
		require.NotNil(t, reader)
		defer reader.Close()
		assert.True(t, md.Exists)

		got, err := UnmarshalSBOM(reader)
		require.NoError(t, err)
		assert.Equal(t, original.Name, got.Name)
		assert.Equal(t, original.Spec.Metadata.Tool.Name, got.Spec.Metadata.Tool.Name)
	})

	t.Run("GetSBOM full fetch round-trips with payload split across many chunks", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		original := sampleSBOMSyft()
		payload, err := original.Marshal()
		require.NoError(t, err)
		rtSrv.serveExists = true
		rtSrv.serveMetadata = &proto.SBOMMetadata{ImageDigest: imageDigest, SyftVersion: syftVersion}
		rtSrv.serveBytes = payload
		// 7-byte chunks force the reader to drain over many Recv calls and
		// across reads smaller than a chunk — exercises the buffering logic.
		rtSrv.serveChunkSize = 7

		md, reader, err := client.GetSBOMStream(context.Background(), imageDigest, syftVersion, false)
		require.NoError(t, err)
		require.NotNil(t, reader)
		defer reader.Close()
		assert.True(t, md.Exists)

		// Read in 13-byte sips to also stress the partial-Read consumer path.
		var collected []byte
		sip := make([]byte, 13)
		for {
			n, err := reader.Read(sip)
			collected = append(collected, sip[:n]...)
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}
		assert.Equal(t, payload, collected, "bytes received over many chunks must match the originally marshaled payload")

		got := &v1beta1.SBOMSyft{}
		require.NoError(t, got.Unmarshal(collected))
		assert.Equal(t, original.Name, got.Name)
	})
}

// TestStorageClient_PutSBOM_NilReader and the next test cover the trivial
// argument-validation paths that the bufconn round-trip doesn't exercise.
func TestStorageClient_PutSBOM_NilReader(t *testing.T) {
	_, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()

	_, err := client.PutSBOMStream(context.Background(), "abc", "1.0.0", proto.SBOMSource_SBOM_SOURCE_WORKLOAD, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestMarshalSBOM_NilInput(t *testing.T) {
	_, err := MarshalSBOM(nil)
	require.Error(t, err)
}

func TestUnmarshalSBOM_NilReader(t *testing.T) {
	_, err := UnmarshalSBOM(nil)
	require.Error(t, err)
}

// sampleContainerProfile is the CP counterpart of sampleSBOMSyft: a
// non-trivial ContainerProfile used to verify that the streaming RPCs
// marshal and unmarshal fields correctly end-to-end.
func sampleContainerProfile() *v1beta1.ContainerProfile {
	return &v1beta1.ContainerProfile{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContainerProfile",
			APIVersion: "spdx.softwarecomposition.kubescape.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-7d4f8b9c-abc12",
			Namespace: "default",
			Labels: map[string]string{
				"app.kubernetes.io/name": "nginx",
			},
		},
	}
}

// TestStorageClient_ContainerProfileStreamRoundTrip mirrors the SBOM
// round-trip test for the new CP streaming RPCs. Stands up a real gRPC
// server over bufconn and verifies the upload + download paths.
func TestStorageClient_ContainerProfileStreamRoundTrip(t *testing.T) {
	t.Run("SendContainerProfileStream marshals and arrives intact", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		original := sampleContainerProfile()
		resp, err := client.SendContainerProfileStream(context.Background(), original)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		require.NotEmpty(t, rtSrv.cpReceivedBytes)
		got := &v1beta1.ContainerProfile{}
		require.NoError(t, got.Unmarshal(rtSrv.cpReceivedBytes))
		assert.Equal(t, original.Name, got.Name)
		assert.Equal(t, original.Namespace, got.Namespace)
		assert.Equal(t, original.Labels["app.kubernetes.io/name"], got.Labels["app.kubernetes.io/name"])
	})

	t.Run("GetContainerProfileStream round-trips with payload split across chunks", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		original := sampleContainerProfile()
		payload, err := original.Marshal()
		require.NoError(t, err)
		rtSrv.cpServeExists = true
		rtSrv.cpServeBytes = payload
		rtSrv.cpServeChunkSize = 7 // small chunks stress the reassembly loop

		got, err := client.GetContainerProfileStream(context.Background(), original.Namespace, original.Name)
		require.NoError(t, err)
		assert.Equal(t, original.Name, got.Name)
		assert.Equal(t, original.Namespace, got.Namespace)
		assert.Equal(t, original.Labels["app.kubernetes.io/name"], got.Labels["app.kubernetes.io/name"])
	})

	t.Run("GetContainerProfileStream returns not-found error when row is absent", func(t *testing.T) {
		_, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		_, err := client.GetContainerProfileStream(context.Background(), "default", "missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestStorageClient_SendContainerProfileStream_NilProfile(t *testing.T) {
	_, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()

	_, err := client.SendContainerProfileStream(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

// TestStorageClient_GetSBOMStream_NotBoundedByCallTimeout proves that the
// streaming RPC is NOT clipped by WithCallTimeout — a regression matthyx
// flagged. We configure a very short callTimeout (well under any
// production stream latency) and have the server delay between chunks
// by an order of magnitude longer. If the stream were wrapped in
// context.WithTimeout(callTimeout) the call would fail with
// "context deadline exceeded"; success here demonstrates the timeout is
// only applied to unary RPCs.
//
// Test stays under 500ms wall time so it's safe in CI.
func TestStorageClient_GetSBOMStream_NotBoundedByCallTimeout(t *testing.T) {
	const (
		shortCallTimeout = 20 * time.Millisecond
		serverChunkDelay = 50 * time.Millisecond
		chunks           = 5
	)
	// Total stream time on the server ≈ chunks * serverChunkDelay = 250ms,
	// which is 12× the configured callTimeout. A unary timeout would have
	// fired within the first 20ms.

	rtSrv, client, cleanup := startBufconnStorageServer(t, WithCallTimeout(shortCallTimeout))
	defer cleanup()

	original := sampleSBOMSyft()
	payload, err := original.Marshal()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(payload), chunks, "payload must have enough bytes to split")

	rtSrv.serveExists = true
	rtSrv.serveMetadata = &proto.SBOMMetadata{ImageDigest: "abc", SyftVersion: "1.0.0"}
	rtSrv.serveBytes = payload
	rtSrv.serveChunkSize = len(payload) / chunks
	rtSrv.serveChunkDelay = serverChunkDelay

	start := time.Now()
	md, reader, err := client.GetSBOMStream(context.Background(), "abc", "1.0.0", false)
	require.NoError(t, err, "stream must not be clipped by callTimeout")
	require.NotNil(t, reader)
	defer reader.Close()
	assert.True(t, md.Exists)

	got, err := UnmarshalSBOM(reader)
	require.NoError(t, err)
	assert.Equal(t, original.Name, got.Name)

	elapsed := time.Since(start)
	assert.Greater(t, elapsed, shortCallTimeout,
		"sanity: the stream genuinely ran longer than callTimeout (otherwise the test proves nothing)")
}

func TestStorageClient_PatchSBOMAnnotations(t *testing.T) {
	const (
		imageDigest = "sha256:deadbeef"
		syftVersion = "1.0.0"
	)

	t.Run("successful patch merges set and delete", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		rtSrv.serveMetadata = &proto.SBOMMetadata{
			ImageDigest: imageDigest,
			SyftVersion: syftVersion,
			Annotations: map[string]string{
				"status":    "initializing",
				"node":      "node-1",
				"to-delete": "val",
			},
		}

		set := map[string]string{
			"status":  "completed",
			"new-key": "hello",
		}
		del := []string{"to-delete"}

		md, err := client.PatchSBOMAnnotations(context.Background(), imageDigest, syftVersion, set, del)
		require.NoError(t, err)
		require.NotNil(t, md)

		// Assert server received request
		require.NotNil(t, rtSrv.receivedPatchReq)
		assert.Equal(t, imageDigest, rtSrv.receivedPatchReq.ImageDigest)
		assert.Equal(t, syftVersion, rtSrv.receivedPatchReq.SyftVersion)
		assert.Equal(t, set, rtSrv.receivedPatchReq.Set)
		assert.Equal(t, del, rtSrv.receivedPatchReq.Delete)

		// Assert returned metadata reflections
		assert.Equal(t, "completed", md.Annotations["status"])
		assert.Equal(t, "node-1", md.Annotations["node"])
		assert.Equal(t, "hello", md.Annotations["new-key"])
		_, exists := md.Annotations["to-delete"]
		assert.False(t, exists)
	})

	t.Run("server error returns formatted failure", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()

		rtSrv.patchResp = &proto.PatchSBOMAnnotationsResponse{
			Success:      false,
			ErrorMessage: "row not found",
			ErrorCode:    proto.ErrorCode_ERROR_CODE_SBOM_NOT_FOUND,
		}

		md, err := client.PatchSBOMAnnotations(context.Background(), imageDigest, syftVersion, map[string]string{"k": "v"}, nil)
		require.Error(t, err)
		assert.Nil(t, md)
		assert.Contains(t, err.Error(), "row not found")
		assert.Contains(t, err.Error(), "ERROR_CODE_SBOM_NOT_FOUND")
	})

	t.Run("not connected returns error", func(t *testing.T) {
		client := &StorageClient{}
		md, err := client.PatchSBOMAnnotations(context.Background(), imageDigest, syftVersion, map[string]string{"k": "v"}, nil)
		require.Error(t, err)
		assert.Nil(t, md)
		assert.Contains(t, err.Error(), "client is not connected")
	})

	t.Run("respects call timeout", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t, WithCallTimeout(20*time.Millisecond))
		defer cleanup()

		rtSrv.patchDelay = 100 * time.Millisecond

		_, err := client.PatchSBOMAnnotations(context.Background(), imageDigest, syftVersion, map[string]string{"k": "v"}, nil)
		require.Error(t, err)
		assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
	})
}

