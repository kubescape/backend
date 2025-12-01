package v1

import (
	"context"
	"testing"
	"time"

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
		expectError bool
	}{
		{
			name:        "valid grpc URL",
			grpcURL:     "grpc://storage.example.com:50051",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "test-cluster",
			expectError: false,
		},
		{
			name:        "valid grpcs URL",
			grpcURL:     "grpcs://storage.example.com:443",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "test-cluster",
			expectError: false,
		},
		{
			name:        "empty URL",
			grpcURL:     "",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "test-cluster",
			expectError: true,
		},
		{
			name:        "invalid scheme",
			grpcURL:     "http://storage.example.com:50051",
			accountID:   "test-account",
			accessKey:   "test-key",
			cluster:     "test-cluster",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewStorageClient(tt.grpcURL, tt.accountID, tt.accessKey, tt.cluster)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.accountID, client.GetAccountID())
				assert.Equal(t, tt.accessKey, client.GetAccessKey())
				assert.Equal(t, tt.cluster, client.GetCluster())
				assert.NotNil(t, client.GetGRPCConfig())
			}
		})
	}
}

func TestStorageClient_SetAccountIDAndAccessKey(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "account1", "key1", "cluster1")
	require.NoError(t, err)

	assert.Equal(t, "account1", client.GetAccountID())
	assert.Equal(t, "key1", client.GetAccessKey())
	assert.Equal(t, "cluster1", client.GetCluster())

	client.SetAccountID("account2")
	assert.Equal(t, "account2", client.GetAccountID())

	client.SetAccessKey("key2")
	assert.Equal(t, "key2", client.GetAccessKey())

	client.SetCluster("cluster2")
	assert.Equal(t, "cluster2", client.GetCluster())
}

func TestStorageClient_SendContainerProfile(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	mockClient := &mockStorageServiceClient{
		sendContainerProfileFunc: func(ctx context.Context, in *proto.SendContainerProfileRequest, opts ...grpc.CallOption) (*proto.SendContainerProfileResponse, error) {
			// customer_guid and cluster are now sent via metadata, not in request
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

	mockClient := &mockStorageServiceClient{
		getProfileFunc: func(ctx context.Context, in *proto.GetProfileRequest, opts ...grpc.CallOption) (*proto.GetProfileResponse, error) {
			assert.Equal(t, "ApplicationProfile", in.Kind)
			assert.Equal(t, "default", in.Namespace)
			assert.Equal(t, "test-app", in.Name)
			// customer_guid and cluster are now sent via metadata, not in request
			return &proto.GetProfileResponse{
				Success:            true,
				ApplicationProfile: &v1beta1.ApplicationProfile{},
			}, nil
		},
	}
	client.protoClient = mockClient

	resp, err := client.GetApplicationProfile(context.Background(), "default", "test-app")
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestStorageClient_NotConnected(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	resp, err := client.SendContainerProfile(context.Background(), &v1beta1.ContainerProfile{})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "not connected")
}

func TestStorageClient_ListApplicationProfiles(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	mockClient := &mockStorageServiceClient{
		listApplicationProfilesFunc: func(ctx context.Context, in *proto.ListApplicationProfilesRequest, opts ...grpc.CallOption) (*proto.ListApplicationProfilesResponse, error) {
			assert.Equal(t, "default", in.Namespace)
			// customer_guid and cluster are now sent via metadata, not in request
			return &proto.ListApplicationProfilesResponse{
				Success: true,
				ApplicationProfiles: []*v1beta1.ApplicationProfile{
					{}, // Spec is nil
					{}, // Spec is nil
				},
			}, nil
		},
	}
	client.protoClient = mockClient

	list, err := client.ListApplicationProfiles(context.Background(), "default")
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Len(t, list.Items, 2)
}

func TestStorageClient_ListNetworkNeighborhoods(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	mockClient := &mockStorageServiceClient{
		listNetworkNeighborhoodsFunc: func(ctx context.Context, in *proto.ListNetworkNeighborhoodsRequest, opts ...grpc.CallOption) (*proto.ListNetworkNeighborhoodsResponse, error) {
			assert.Equal(t, "kube-system", in.Namespace)
			// customer_guid and cluster are now sent via metadata, not in request
			return &proto.ListNetworkNeighborhoodsResponse{
				Success: true,
				NetworkNeighborhoods: []*v1beta1.NetworkNeighborhood{
					{}, // Spec is nil
					{}, // Spec is nil
					{}, // Spec is nil
				},
			}, nil
		},
	}
	client.protoClient = mockClient

	list, err := client.ListNetworkNeighborhoods(context.Background(), "kube-system")
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Len(t, list.Items, 3)
}

func TestStorageClient_ListApplicationProfiles_NotConnected(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	list, err := client.ListApplicationProfiles(context.Background(), "default")
	assert.Error(t, err)
	assert.Nil(t, list)
	assert.Contains(t, err.Error(), "not connected")
}

func TestStorageClient_ListNetworkNeighborhoods_NotConnected(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "test-account", "test-key", "test-cluster")
	require.NoError(t, err)

	list, err := client.ListNetworkNeighborhoods(context.Background(), "default")
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
