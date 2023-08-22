package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetReporterClusterReportsWebsocketUrl(t *testing.T) {
	url, err := GetReporterClusterReportsWebsocketUrl("wss://some-host", "abc", "cccc1")
	assert.NoError(t, err)
	assert.Equal(t, "wss://some-host/k8s/cluster-reports?clusterName=cccc1&customerGUID=abc", url.String())
	url, err = GetReporterClusterReportsWebsocketUrl("wss://some-host", "abc", "cccc1")
	assert.NoError(t, err)
	assert.Equal(t, "wss://some-host/k8s/cluster-reports?clusterName=cccc1&customerGUID=abc", url.String())
}

func Test_GetRegistryRepositoriesUrl(t *testing.T) {
	url, err := GetRegistryRepositoriesUrl("https://some-host", "00000-aaaaa", "quay.io", "1234")
	assert.NoError(t, err)
	assert.Equal(t, "https://some-host/k8s/registryRepositories?customerGUID=00000-aaaaa&jobID=1234&registryName=quay.io", url.String())
}

func Test_GetPostureReportUrl(t *testing.T) {
	assert.Panics(t, func() {
		_, _ = GetPostureReportUrl("https://some-host", "invalid-customer-uuid", "cluster-123", "report-123")
	})

	url, err := GetPostureReportUrl("https://some-host", "11111111-1111-1111-1111-111111111111", "cluster-123", "report-123")
	assert.NoError(t, err)
	assert.Equal(t, "https://some-host/k8s/v2/postureReport?clusterName=cluster-123&contextName=cluster-123&customerGUID=11111111-1111-1111-1111-111111111111&reportGUID=report-123", url.String())
}
