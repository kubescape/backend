package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/armoapi-go/identifiers"
	httputils "github.com/armosec/utils-go/httputils"
	v1 "github.com/kubescape/backend/pkg/server/v1"
	"github.com/kubescape/backend/pkg/utils"
)

func getCVEExceptionsURL(backendURL, customerGUID string, designators *identifiers.PortalDesignator) (*url.URL, error) {
	scheme, host := utils.ParseHost(backendURL)
	expURL := &url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   v1.ApiServerVulnerabilitiesExceptionsPath,
	}
	qValues := expURL.Query()
	for k, v := range designators.Attributes {
		qValues.Add(k, v)
	}
	qValues.Add(v1.QueryParamCustomerGUID, customerGUID)

	expURL.RawQuery = qValues.Encode()
	return expURL, nil
}

func getCVEExceptionByDEsignator(backendURL, customerGUID string, designators *identifiers.PortalDesignator) ([]armotypes.VulnerabilityExceptionPolicy, error) {

	var vulnerabilityExceptionPolicy []armotypes.VulnerabilityExceptionPolicy

	url, err := getCVEExceptionsURL(backendURL, customerGUID, designators)
	if err != nil {
		return nil, err
	}

	resp, err := httputils.HttpGet(http.DefaultClient, url.String(), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("getCVEExceptionByDEsignator: resp.StatusCode %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bodyBytes, &vulnerabilityExceptionPolicy)
	if err != nil {
		return nil, err
	}

	return vulnerabilityExceptionPolicy, nil
}

func GetCVEExceptionByDesignator(baseURL, customerGUID string, designators *identifiers.PortalDesignator) ([]armotypes.VulnerabilityExceptionPolicy, error) {
	vulnerabilityExceptionPolicyList, err := getCVEExceptionByDEsignator(baseURL, customerGUID, designators)
	if err != nil {
		return nil, err
	}
	return vulnerabilityExceptionPolicyList, nil
}

func GetVulnerabilitiesReportURL(eventReceiverUrl, customerGUID string) *url.URL {
	scheme, host := utils.ParseHost(eventReceiverUrl)
	urlBase := &url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   v1.ReporterVulnerabilitiesReportPath,
	}
	q := urlBase.Query()
	q.Add(armotypes.CustomerGuidQuery, customerGUID)
	urlBase.RawQuery = q.Encode()
	return urlBase
}

func GetSystemReportURL(eventReceiverUrl string) *url.URL {
	scheme, host := utils.ParseHost(eventReceiverUrl)
	urlBase := &url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   v1.ReporterSystemReportPath,
	}
	return urlBase
}
