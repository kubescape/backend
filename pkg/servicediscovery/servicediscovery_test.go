package servicediscovery

import (
	"flag"
	"fmt"
	"testing"

	v1 "github.com/kubescape/backend/pkg/servicediscovery/v1"
	"github.com/stretchr/testify/assert"
)

var testUrl string

func init() {
	flag.StringVar(&testUrl, "url", "", "Service Discovery Server To Test Against")
}

func TestServiceDiscoveryV1(t *testing.T) {
	flag.Parse()
	if testUrl == "" {
		t.Skip("skipping test because no URL was provided")
	}

	client := v1.NewServiceDiscoveryClientV1(testUrl)
	sdUrl := client.GetServiceDiscoveryUrl()
	t.Logf("testing URL: %s", sdUrl)
	services, err := GetServices(client)
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
