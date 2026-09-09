package servicediscovery

import (
	"flag"
	"fmt"
	"testing"

	"github.com/kubescape/backend/pkg/servicediscovery/schema"
	v1 "github.com/kubescape/backend/pkg/servicediscovery/v1"
	v2 "github.com/kubescape/backend/pkg/servicediscovery/v2"
	v3 "github.com/kubescape/backend/pkg/servicediscovery/v3"
	v4 "github.com/kubescape/backend/pkg/servicediscovery/v4"
	"github.com/stretchr/testify/assert"
)

// v1
var _ schema.IServiceDiscoveryServer = &v1.ServiceDiscoveryServerV1{}
var _ schema.IServiceDiscoveryClient = &v1.ServiceDiscoveryClientV1{}

// v2
var _ schema.IServiceDiscoveryServer = &v2.ServiceDiscoveryServerV2{}
var _ schema.IServiceDiscoveryClient = &v2.ServiceDiscoveryClientV2{}

// v3
var _ schema.IServiceDiscoveryServer = &v3.ServiceDiscoveryServerV3{}
var _ schema.IServiceDiscoveryClient = &v3.ServiceDiscoveryClientV3{}

// v4
var _ schema.IServiceDiscoveryServer = &v4.ServiceDiscoveryServerV4{}
var _ schema.IServiceDiscoveryClient = &v4.ServiceDiscoveryClientV4{}

var testUrl string
var testVersion string

func init() {
	flag.StringVar(&testUrl, "url", "", "Service Discovery Server To Test Against")
	flag.StringVar(&testVersion, "version", "", "Service Discovery Version To Test Against")
}

func TestServiceDiscoveryClientV1(t *testing.T) {
	flag.Parse()
	if testUrl == "" {
		t.Skip("skipping test because no URL was provided")
	}

	if testVersion != "v1" {
		t.Skip()
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

func TestServiceDiscoveryClientV2(t *testing.T) {
	flag.Parse()
	if testUrl == "" {
		t.Skip("skipping test because no URL was provided")
	}
	if testVersion != "v2" {
		t.Skip()
	}

	client, err := v2.NewServiceDiscoveryClientV2(testUrl)
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
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetGatewayUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetReportReceiverWebsocketUrl())
	assert.NotEmpty(t, services.GetSynchronizerUrl())
}

func TestServiceDiscoveryFileV2(t *testing.T) {
	file := v2.NewServiceDiscoveryFileV2("testdata/v2.json")
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
	assert.NotEmpty(t, services.GetSynchronizerUrl())
}

func TestServiceDiscoveryStreamV2(t *testing.T) {
	stream := []byte("{\"version\": \"v2\",\"response\": {\"event-receiver-http\": \"https://er-test.com\",\"event-receiver-ws\": \"wss://er-test.com\",\"gateway\": \"https://gw.test.com\",\"api-server\": \"https://api.test.com\",\"metrics\": \"https://metrics.test.com\", \"synchronizer\": \"wss://synchronizer.test.com\"}}")
	services, err := GetServices(
		v2.NewServiceDiscoveryStreamV2(stream),
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
	assert.NotEmpty(t, services.GetSynchronizerUrl())
}

func TestServiceDiscoveryClientV3(t *testing.T) {
	flag.Parse()
	if testUrl == "" {
		t.Skip("skipping test because no URL was provided")
	}
	if testVersion != "v3" {
		t.Skip()
	}

	client, err := v3.NewServiceDiscoveryClientV3(testUrl)
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

	// deprecated methods
	assert.Panics(t, func() { services.GetReportReceiverWebsocketUrl() })
	assert.Panics(t, func() { services.GetGatewayUrl() })

	// supported methods
	assert.NotEmpty(t, services.GetApiServerUrl())
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetSynchronizerUrl())
	assert.NotEmpty(t, services.GetStorageUrl())

}

func TestServiceDiscoveryFileV3(t *testing.T) {
	file := v3.NewServiceDiscoveryFileV3("testdata/v3.json")
	services, err := GetServices(file)
	if err != nil {
		assert.FailNowf(t, "failed to get services from file: %s", err.Error())
	}

	assert.NotNil(t, services)

	// deprecated methods
	assert.Panics(t, func() { services.GetReportReceiverWebsocketUrl() })
	assert.Panics(t, func() { services.GetGatewayUrl() })

	// supported methods
	assert.NotEmpty(t, services.GetApiServerUrl())
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetSynchronizerUrl())
	assert.NotEmpty(t, services.GetStorageUrl())
	// otel-events lives in v4: v3 never carries it
	assert.Empty(t, services.GetOtelEventsUrl())
}

func TestServiceDiscoveryStreamV3(t *testing.T) {
	stream := []byte("{\"version\": \"v3\",\"response\": {\"storage\":\"https://grpc.test.com\",\"event-receiver-http\": \"https://er-test.com\",\"api-server\": \"https://api.test.com\",\"metrics\": \"https://metrics.test.com\", \"synchronizer\": \"wss://synchronizer.test.com\"}}")
	services, err := GetServices(
		v3.NewServiceDiscoveryStreamV3(stream),
	)
	if err != nil {
		assert.FailNowf(t, "failed to get services from stream: %s", err.Error())
	}

	assert.NotNil(t, services)
	// deprecated methods
	assert.Panics(t, func() { services.GetReportReceiverWebsocketUrl() })
	assert.Panics(t, func() { services.GetGatewayUrl() })

	// supported methods
	assert.NotEmpty(t, services.GetApiServerUrl())
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetSynchronizerUrl())
	assert.NotEmpty(t, services.GetStorageUrl())
	assert.Empty(t, services.GetOtelEventsUrl())
}

func TestServiceDiscoveryClientV4(t *testing.T) {
	flag.Parse()
	if testUrl == "" {
		t.Skip("skipping test because no URL was provided")
	}
	if testVersion != "v4" {
		t.Skip()
	}

	client, err := v4.NewServiceDiscoveryClientV4(testUrl)
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

	// deprecated methods
	assert.Panics(t, func() { services.GetReportReceiverWebsocketUrl() })
	assert.Panics(t, func() { services.GetGatewayUrl() })

	// supported methods
	assert.NotEmpty(t, services.GetApiServerUrl())
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetSynchronizerUrl())
	assert.NotEmpty(t, services.GetStorageUrl())
	// otel-events is optional: a backend without the collector omits it
}

func TestServiceDiscoveryFileV4(t *testing.T) {
	file := v4.NewServiceDiscoveryFileV4("testdata/v4.json")
	services, err := GetServices(file)
	if err != nil {
		assert.FailNowf(t, "failed to get services from file: %s", err.Error())
	}

	assert.NotNil(t, services)

	// deprecated methods
	assert.Panics(t, func() { services.GetReportReceiverWebsocketUrl() })
	assert.Panics(t, func() { services.GetGatewayUrl() })

	// supported methods
	assert.NotEmpty(t, services.GetApiServerUrl())
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetSynchronizerUrl())
	assert.NotEmpty(t, services.GetStorageUrl())
	assert.Equal(t, "otel-events.test.com:443", services.GetOtelEventsUrl())
}

func TestServiceDiscoveryStreamV4(t *testing.T) {
	stream := []byte("{\"version\": \"v4\",\"response\": {\"storage\":\"https://grpc.test.com\",\"event-receiver-http\": \"https://er-test.com\",\"api-server\": \"https://api.test.com\",\"metrics\": \"https://metrics.test.com\", \"synchronizer\": \"wss://synchronizer.test.com\"}}")
	services, err := GetServices(
		v4.NewServiceDiscoveryStreamV4(stream),
	)
	if err != nil {
		assert.FailNowf(t, "failed to get services from stream: %s", err.Error())
	}

	assert.NotNil(t, services)
	// deprecated methods
	assert.Panics(t, func() { services.GetReportReceiverWebsocketUrl() })
	assert.Panics(t, func() { services.GetGatewayUrl() })

	// supported methods
	assert.NotEmpty(t, services.GetApiServerUrl())
	assert.NotEmpty(t, services.GetMetricsUrl())
	assert.NotEmpty(t, services.GetReportReceiverHttpUrl())
	assert.NotEmpty(t, services.GetSynchronizerUrl())
	assert.NotEmpty(t, services.GetStorageUrl())
	// otel-events is optional: absent in the stream above, so it parses to empty
	assert.Empty(t, services.GetOtelEventsUrl())
}

func TestServiceDiscoveryServerV4OtelEventsOmittedWhenUnset(t *testing.T) {
	withOtel := v4.NewServiceDiscoveryServerV4(v4.ServicesV4{
		EventReceiverHttpUrl: "https://er-test.com",
		ApiServerUrl:         "https://api.test.com",
		MetricsUrl:           "metrics.test.com:443",
		SynchronizerUrl:      "wss://synchronizer.test.com",
		StorageUrl:           "grpcs://grpc.test.com:443",
		OtelEventsUrl:        "otel-events.test.com:443",
	})
	assert.Contains(t, string(withOtel.GetResponse()), `"otel-events":"otel-events.test.com:443"`)

	withoutOtel := v4.NewServiceDiscoveryServerV4(v4.ServicesV4{
		EventReceiverHttpUrl: "https://er-test.com",
		ApiServerUrl:         "https://api.test.com",
		MetricsUrl:           "metrics.test.com:443",
		SynchronizerUrl:      "wss://synchronizer.test.com",
		StorageUrl:           "grpcs://grpc.test.com:443",
	})
	assert.NotContains(t, string(withoutOtel.GetResponse()), "otel-events")

	// round trip through the client parser
	services, err := GetServices(v4.NewServiceDiscoveryStreamV4([]byte(`{"version":"v4","response":` + string(withOtel.GetResponse()) + `}`)))
	assert.NoError(t, err)
	assert.Equal(t, "otel-events.test.com:443", services.GetOtelEventsUrl())
}

func TestV3ResponseNeverCarriesOtelEvents(t *testing.T) {
	server := v3.NewServiceDiscoveryServerV3(v3.ServicesV3{
		EventReceiverHttpUrl: "https://er-test.com",
		ApiServerUrl:         "https://api.test.com",
		MetricsUrl:           "metrics.test.com:443",
		SynchronizerUrl:      "wss://synchronizer.test.com",
		StorageUrl:           "grpcs://grpc.test.com:443",
	})
	assert.NotContains(t, string(server.GetResponse()), "otel-events")
	var services v3.ServicesV3
	services.SetOtelEventsUrl("ignored")
	assert.Empty(t, services.GetOtelEventsUrl())
}

func TestOtelEventsUrlIsEmptyOnV1AndV2(t *testing.T) {
	v1Services, err := GetServices(v1.NewServiceDiscoveryFileV1("testdata/v1.json"))
	assert.NoError(t, err)
	assert.NotPanics(t, func() { v1Services.SetOtelEventsUrl("ignored") })
	assert.Empty(t, v1Services.GetOtelEventsUrl())

	v2Services, err := GetServices(v2.NewServiceDiscoveryFileV2("testdata/v2.json"))
	assert.NoError(t, err)
	assert.NotPanics(t, func() { v2Services.SetOtelEventsUrl("ignored") })
	assert.Empty(t, v2Services.GetOtelEventsUrl())
}

func TestServiceDiscoveryV4ParseErrorIsSurfaced(t *testing.T) {
	_, err := GetServices(v4.NewServiceDiscoveryStreamV4([]byte(`{"version":"v4","response":"not-an-object"}`)))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid response")
	assert.Contains(t, err.Error(), "json")
}
