package v1

import (
	"encoding/json"
	"fmt"
	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/armoapi-go/identifiers"
	httputils "github.com/armosec/utils-go/httputils"
	v1 "github.com/kubescape/backend/pkg/server/v1"
	"github.com/kubescape/backend/pkg/utils"
	"io"
	"net/http"
	"net/url"
)

func constructCVEExceptionsURL(backendURL, customerGUID string, queryParams *url.Values) (*url.URL, error) {
	scheme, host, err := utils.ParseHost(backendURL)
	if err != nil {
		return nil, err
	}
	expURL := &url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   v1.ApiServerVulnerabilitiesExceptionsPathOld,
	}
	queryParams.Add(v1.QueryParamCustomerGUID, customerGUID)
	expURL.RawQuery = queryParams.Encode()
	return expURL, nil
}

func getCVEExceptionsURL(backendURL, customerGUID string, designators *identifiers.PortalDesignator) (*url.URL, error) {
	qValues := url.Values{}
	for k, v := range designators.Attributes {
		qValues.Add(k, v)
	}
	return constructCVEExceptionsURL(backendURL, customerGUID, &qValues)
}

func getCVEExceptionsURLByRawQuery(backendURL, customerGUID string, rawQuery *url.Values) (*url.URL, error) {
	return constructCVEExceptionsURL(backendURL, customerGUID, rawQuery)
}

func fetchCVEExceptions(url *url.URL, headers map[string]string) ([]armotypes.VulnerabilityExceptionPolicy, error) {
	var vulnerabilityExceptionPolicy []armotypes.VulnerabilityExceptionPolicy

	resp, err := httputils.HttpGet(http.DefaultClient, url.String(), headers)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetchCVEExceptions: resp.StatusCode %d", resp.StatusCode)
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

func GetCVEExceptionByDesignator(backendURL, customerGUID string, designators *identifiers.PortalDesignator, headers map[string]string) ([]armotypes.VulnerabilityExceptionPolicy, error) {
	url, err := getCVEExceptionsURL(backendURL, customerGUID, designators)
	if err != nil {
		return nil, err
	}
	return fetchCVEExceptions(url, headers)
}

func GetCVEExceptionByRawQuery(backendURL, customerGUID string, rawQuery *url.Values, headers map[string]string) ([]armotypes.VulnerabilityExceptionPolicy, error) {
	url, err := getCVEExceptionsURLByRawQuery(backendURL, customerGUID, rawQuery)
	if err != nil {
		return nil, err
	}
	return fetchCVEExceptions(url, headers)
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
