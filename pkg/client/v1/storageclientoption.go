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
	Region                 string
	CloudAccountIdentifier string
	KnownChecksum          string
}

// WithProfileRegion sets the region for non-k8s scoped resources
func WithProfileRegion(region string) ProfileOption {
	return func(o *ProfileOptions) {
		o.Region = region
	}
}

// WithProfileCloudAccountIdentifier sets the cloud account identifier for non-k8s scoped resources (e.g. AWS account ID, GCP project ID)
func WithProfileCloudAccountIdentifier(cloudAccountIdentifier string) ProfileOption {
	return func(o *ProfileOptions) {
		o.CloudAccountIdentifier = cloudAccountIdentifier
	}
}

// WithProfileKnownChecksum sets the content checksum of the profile the caller already holds.
// It is advisory: an empty value (the default) requests the body unconditionally, and a server
// that does not understand it streams the body as before. When the server recognises it and the
// checksums match, GetContainerProfileStream returns ErrProfileUnchanged instead of a profile.
// Only GetContainerProfileStream honours this option.
func WithProfileKnownChecksum(checksum string) ProfileOption {
	return func(o *ProfileOptions) {
		o.KnownChecksum = checksum
	}
}

// profileOptionsWithDefaults applies profile query options
func profileOptionsWithDefaults(opts []ProfileOption) *ProfileOptions {
	options := &ProfileOptions{
		Region:                 "",
		CloudAccountIdentifier: "",
		KnownChecksum:          "",
	}

	for _, apply := range opts {
		apply(options)
	}

	return options
}
