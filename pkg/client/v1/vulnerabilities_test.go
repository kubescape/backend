package v1

import (
	"testing"

	"github.com/armosec/armoapi-go/identifiers"
	"github.com/stretchr/testify/assert"
)

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
