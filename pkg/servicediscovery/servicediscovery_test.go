package servicediscovery

import (
	"flag"
	"fmt"
	"testing"

	"github.com/kubescape/backend/pkg/servicediscovery/schema"
	v1 "github.com/kubescape/backend/pkg/servicediscovery/v1"
	"github.com/stretchr/testify/assert"
)

var _ schema.IServiceDiscoveryServer = &v1.ServiceDiscoveryServerV1{}
var _ schema.IServiceDiscoveryClient = &v1.ServiceDiscoveryClientV1{}

var testUrl string

func init() {
	flag.StringVar(&testUrl, "url", "", "Service Discovery Server To Test Against")
}

func TestServiceDiscoveryClientV1(t *testing.T) {
	flag.Parse()
	if testUrl == "" {
		t.Skip("skipping test because no URL was provided")
	}

	client, err := v1.NewServiceDiscoveryClientV1(testUrl)
	if err != nil {
		t.Fatalf("failed to create client: %s", err.Error())
	}
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

func TestServiceDiscoveryFileV1(t *testing.T) {
	file := v1.NewServiceDiscoveryFileV1("testdata/v1.json")
	services, err := GetServices(file)
	if err != nil {
		assert.FailNowf(t, "failed to get services from file: %s", err.Error())
	}

	assert.NotNil(t, services)
	assert.NotEmpty(t, services.GetApiServerUrl())
	assert.NotEmpty(t, services.GetGatewayUrl())
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetReportReceiverWebsocketUrl())
}

func TestServiceDiscoveryStreamV1(t *testing.T) {
	stream := []byte("{\"version\": \"v1\",\"response\": {\"event-receiver-http\": \"https://er-test.com\",\"event-receiver-ws\": \"wss://er-test.com\",\"gateway\": \"https://gw.test.com\",\"api-server\": \"https://api.test.com\",\"metrics\": \"https://metrics.test.com\"}}")
	services, err := GetServices(
		v1.NewServiceDiscoveryStreamV1(stream),
	)
	if err != nil {
		assert.FailNowf(t, "failed to get services from stream: %s", err.Error())
	}

	assert.NotNil(t, services)
	assert.NotEmpty(t, services.GetApiServerUrl())
	assert.NotEmpty(t, services.GetGatewayUrl())
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetReportReceiverWebsocketUrl())
}
