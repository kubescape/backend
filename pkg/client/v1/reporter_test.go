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
