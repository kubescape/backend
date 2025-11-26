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

// storageClientOptionsWithDefaults sets defaults for the Storage client and applies overrides
func storageClientOptionsWithDefaults(opts []StorageClientOption) *StorageClientOptions {
	defaultCallTimeout := 30 * time.Second

	options := &StorageClientOptions{
		callTimeout: &defaultCallTimeout,
		withTrace:   false,
	}

	for _, apply := range opts {
		apply(options)
	}

	return options
}
