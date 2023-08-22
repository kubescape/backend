package v1

import (
	"testing"

	"github.com/armosec/armoapi-go/identifiers"
	"github.com/stretchr/testify/assert"
)

func Test_getCVEExceptionsURL(t *testing.T) {
	url, err := getCVEExceptionsURL("http://localhost:8080", &identifiers.PortalDesignator{
		Attributes: map[string]string{
			identifiers.AttributeCustomerGUID: "test123",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/api/v1/vulnerabilitiesExceptions?customerGUID=test123", url.String())
}

func Test_GetVulnerabilitiesReportURL(t *testing.T) {
	url := GetVulnerabilitiesReportURL("https://localhost:8080", "123-abc")
	assert.Equal(t, "https://localhost:8080/k8s/v2/containerScan?customerGUID=123-abc", url.String())
}

func Test_GetSystemReportURL(t *testing.T) {
	url := GetSystemReportURL("https://localhost:8080")
	assert.Equal(t, "https://localhost:8080/k8s/sysreport", url.String())
}
