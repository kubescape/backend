package v1

import (
	"context"
	"testing"
	"time"

	"github.com/armosec/armoapi-go/armotypes"
	"github.com/kubescape/backend/pkg/client/v1/proto"
	"github.com/kubescape/storage/pkg/apis/softwarecomposition/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// Mock StorageServiceClient for testing
type mockStorageServiceClient struct {
	sendContainerProfileFunc     func(ctx context.Context, in *proto.SendContainerProfileRequest, opts ...grpc.CallOption) (*proto.SendContainerProfileResponse, error)
	getProfileFunc               func(ctx context.Context, in *proto.GetProfileRequest, opts ...grpc.CallOption) (*proto.GetProfileResponse, error)
	listApplicationProfilesFunc  func(ctx context.Context, in *proto.ListApplicationProfilesRequest, opts ...grpc.CallOption) (*proto.ListApplicationProfilesResponse, error)
	listNetworkNeighborhoodsFunc func(ctx context.Context, in *proto.ListNetworkNeighborhoodsRequest, opts ...grpc.CallOption) (*proto.ListNetworkNeighborhoodsResponse, error)
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
		name         string
		kind         armotypes.ProfileKind
		namespace    string
		profileName  string
		region       string
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
					if tt.kind == armotypes.ApplicationProfileKind {
						return &proto.GetProfileResponse{
							Success:            true,
							ApplicationProfile: &v1beta1.ApplicationProfile{},
						}, nil
					}
					return &proto.GetProfileResponse{
						Success:             true,
						NetworkNeighborhood: &v1beta1.NetworkNeighborhood{},
					}, nil
				},
			}
			client.protoClient = mockClient

			if tt.kind == armotypes.ApplicationProfileKind {
				resp, err := client.GetApplicationProfile(context.Background(), tt.namespace, tt.profileName, WithProfileRegion(tt.region), WithProfileCloudAccountIdentifier(tt.cloudAccountIdentifier))
				require.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
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
		name         string
		namespace    string
		limit        int64
		cont         string
		region       string
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
					profiles := make([]*v1beta1.ApplicationProfile, tt.expectedLen)
					for i := range profiles {
						profiles[i] = &v1beta1.ApplicationProfile{}
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
		name         string
		namespace    string
		limit        int64
		cont         string
		region       string
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
					neighborhoods := make([]*v1beta1.NetworkNeighborhood, tt.expectedLen)
					for i := range neighborhoods {
						neighborhoods[i] = &v1beta1.NetworkNeighborhood{}
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
