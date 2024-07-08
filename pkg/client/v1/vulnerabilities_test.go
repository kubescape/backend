package v1

import (
	url2 "net/url"
	"testing"

	"github.com/armosec/armoapi-go/identifiers"
	"github.com/stretchr/testify/assert"
)

func Test_getCVEExceptionsURLByRawQuery(t *testing.T) {
	url, err := getCVEExceptionsURLByRawQuery("http://localhost:8080", "abc", &url2.Values{
		"scope.namespace": []string{"kube-system", "*/*"},
		"scope.cluster":   []string{"c1", "c2"},
		"scope.name":      []string{"n1", "*/*"},
		"scope.kind":      []string{"deployment"},
		"scope.other":     []string{""},
	})
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/api/v1/armoVulnerabilityExceptions?customerGUID=abc&scope.cluster=c1&scope.cluster=c2&scope.kind=deployment&scope.name=n1&scope.name=%2A%2F%2A&scope.namespace=kube-system&scope.namespace=%2A%2F%2A&scope.other=", url.String())
}

func Test_getCVEExceptionsURL(t *testing.T) {
	url, err := getCVEExceptionsURL("http://localhost:8080", "abc", &identifiers.PortalDesignator{
		Attributes: map[string]string{
			"key1": "val2",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/api/v1/armoVulnerabilityExceptions?customerGUID=abc&key1=val2", url.String())
}

func Test_GetVulnerabilitiesReportURL(t *testing.T) {
	url, err := GetVulnerabilitiesReportURL("https://localhost:8080", "123-abc")
	assert.NoError(t, err)

	assert.Equal(t, "https://localhost:8080/k8s/v2/containerScan?customerGUID=123-abc", url.String())
}

func Test_GetSystemReportURL(t *testing.T) {
	url, err := GetSystemReportURL("https://localhost:8080", "123")
	assert.NoError(t, err)
	assert.Equal(t, "https://localhost:8080/k8s/sysreport?customerGUID=123", url.String())
}
