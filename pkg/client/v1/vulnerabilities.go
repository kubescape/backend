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
	scheme, host, err := utils.ParseHost(backendURL)
	if err != nil {
		return nil, err
	}
	expURL := &url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   v1.ApiServerVulnerabilitiesExceptionsPathOld,
	}
	qValues := expURL.Query()
	for k, v := range designators.Attributes {
		qValues.Add(k, v)
	}
	qValues.Add(v1.QueryParamCustomerGUID, customerGUID)

	expURL.RawQuery = qValues.Encode()
	return expURL, nil
}

func getCVEExceptionByDEsignator(backendURL, customerGUID string, designators *identifiers.PortalDesignator, headers map[string]string) ([]armotypes.VulnerabilityExceptionPolicy, error) {

	var vulnerabilityExceptionPolicy []armotypes.VulnerabilityExceptionPolicy

	url, err := getCVEExceptionsURL(backendURL, customerGUID, designators)
	if err != nil {
		return nil, err
	}

	resp, err := httputils.HttpGet(http.DefaultClient, url.String(), headers)
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

func GetCVEExceptionByDesignator(baseURL, customerGUID string, designators *identifiers.PortalDesignator, headers map[string]string) ([]armotypes.VulnerabilityExceptionPolicy, error) {
	vulnerabilityExceptionPolicyList, err := getCVEExceptionByDEsignator(baseURL, customerGUID, designators, headers)
	if err != nil {
		return nil, err
	}
	return vulnerabilityExceptionPolicyList, nil
}

func GetVulnerabilitiesReportURL(eventReceiverUrl, customerGUID string) (*url.URL, error) {
	scheme, host, err := utils.ParseHost(eventReceiverUrl)
	if err != nil {
		return nil, err
	}

	urlBase := &url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   v1.ReporterVulnerabilitiesReportPath,
	}
	q := urlBase.Query()
	q.Add(armotypes.CustomerGuidQuery, customerGUID)
	urlBase.RawQuery = q.Encode()
	return urlBase, nil
}

func GetSystemReportURL(eventReceiverUrl, customerGUID string) (*url.URL, error) {
	scheme, host, err := utils.ParseHost(eventReceiverUrl)
	if err != nil {
		return nil, err
	}
	urlBase := &url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   v1.ReporterSystemReportPath,
	}

	q := urlBase.Query()
	q.Add(armotypes.CustomerGuidQuery, customerGUID)
	urlBase.RawQuery = q.Encode()
	return urlBase, nil
}
