package v1

import (
	"time"
)

// StorageClientOption allows to configure the behavior of the Storage client
type StorageClientOption func(*StorageClientOptions)

// StorageClientOptions holds all the configurable parts of the Storage client
type StorageClientOptions struct {
	callTimeout *time.Duration
	withTrace   bool
	hostType    string
	hostID      string
}

// WithCallTimeout sets the timeout for individual gRPC calls
// A value of 0 means no timeout.
// The default is 30 seconds.
func WithCallTimeout(timeout time.Duration) StorageClientOption {
	duration := timeout
	return func(o *StorageClientOptions) {
		o.callTimeout = &duration
	}
}

// WithStorageTrace toggles request/response tracing for debugging
func WithStorageTrace(enabled bool) StorageClientOption {
	return func(o *StorageClientOptions) {
		o.withTrace = enabled
	}
}

// WithHostType sets the host type (e.g., "kubernetes", "ec2", "ecs")
// If not set, defaults to "kubernetes" on the server side
func WithHostType(hostType string) StorageClientOption {
	return func(o *StorageClientOptions) {
		o.hostType = hostType
	}
}

// WithHostID sets the host ID (e.g., EC2 instance ID)
// Required for non-cluster-based host types
func WithHostID(hostID string) StorageClientOption {
	return func(o *StorageClientOptions) {
		o.hostID = hostID
	}
}

// storageClientOptionsWithDefaults sets defaults for the Storage client and applies overrides
func storageClientOptionsWithDefaults(opts []StorageClientOption) *StorageClientOptions {
	defaultCallTimeout := 30 * time.Second

	options := &StorageClientOptions{
		callTimeout: &defaultCallTimeout,
		withTrace:   false,
		hostType:    "",
		hostID:      "",
	}

	for _, apply := range opts {
		apply(options)
	}

	return options
}

// ========== Profile Query Options ==========

// ProfileOption allows to configure profile queries
type ProfileOption func(*ProfileOptions)

// ProfileOptions holds configuration for profile queries
type ProfileOptions struct {
	Region       string
	AWSAccountID string
}

// WithProfileRegion sets the region for non-k8s scoped resources
func WithProfileRegion(region string) ProfileOption {
	return func(o *ProfileOptions) {
		o.Region = region
	}
}

// WithProfileAWSAccountID sets the AWS account ID for non-k8s scoped resources
func WithProfileAWSAccountID(accountID string) ProfileOption {
	return func(o *ProfileOptions) {
		o.AWSAccountID = accountID
	}
}

// profileOptionsWithDefaults applies profile query options
func profileOptionsWithDefaults(opts []ProfileOption) *ProfileOptions {
	options := &ProfileOptions{
		Region:       "",
		AWSAccountID: "",
	}

	for _, apply := range opts {
		apply(options)
	}

	return options
}
