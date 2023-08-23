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
