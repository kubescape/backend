package v1

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testUrl string

func init() {
	flag.StringVar(&testUrl, "url", "", "Service Discovery Server To Test Against")
}

func TestServiceDiscovery(t *testing.T) {
	flag.Parse()
	if testUrl == "" {
		t.Skip("skipping test because no URL was provided")
	}

	server := NewServiceDiscoveryServer(testUrl)
	sdUrl := server.GetServiceDiscoveryUrl()
	t.Logf("testing URL: %s", sdUrl)
	services, err := GetServices(context.Background(), server)
	if err != nil {
		assert.FailNowf(t, fmt.Sprintf("failed to get services from url: %s (HTTP GET)", sdUrl), err.Error())
	}

	assert.NotNil(t, services)
	assert.NotEmpty(t, services.GetApiServerUrl())
	assert.NotEmpty(t, services.GetGatewayUrl())
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetReportReceiverWebsocketUrl())
}
