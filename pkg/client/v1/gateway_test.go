package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetRootGatewayUrl(t *testing.T) {
	url, err := GetRootGatewayUrl("wss://some-host")
	assert.NoError(t, err)
	assert.Equal(t, "wss://some-host/v1/waitfornotification", url.String())
	url, err = GetRootGatewayUrl("some-host")
	assert.NoError(t, err)
	assert.Equal(t, "wss://some-host/v1/waitfornotification", url.String())
}
