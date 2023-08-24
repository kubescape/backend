package v1

import (
	"net/url"

	"github.com/google/uuid"
	v1 "github.com/kubescape/backend/pkg/server/v1"
	"github.com/kubescape/backend/pkg/utils"
)

func GetReporterClusterReportsWebsocketUrl(eventReceiverWebsocketUrl, accountID, clusterName string) (*url.URL, error) {
	scheme, host, err := utils.ParseHost(eventReceiverWebsocketUrl)
	if err != nil {
		return nil, err
	}

	// if no scheme is specified, calling ParseHost default to https, so we need to change it to wss
	if scheme == "https" {
		scheme = "wss"
	}
	u := &url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   v1.ReporterWebsocketClusterReportsPath,
	}

	q := u.Query()
	q.Add(v1.QueryParamCustomerGUID, accountID)
	q.Add(v1.QueryParamClusterName, clusterName)
	u.RawQuery = q.Encode()
	u.ForceQuery = true

	return u, nil
}

func GetRegistryRepositoriesUrl(eventReceiverRestUrl, customerGUID, registryName, jobID string) (*url.URL, error) {
	scheme, host, err := utils.ParseHost(eventReceiverRestUrl)
	if err != nil {
		return nil, err
	}

	urlQuery := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   "k8s/registryRepositories",
	}
	query := url.Values{
		v1.QueryParamJobID:        []string{jobID},
		v1.QueryParamCustomerGUID: []string{customerGUID},
		v1.QueryParamRegistryName: []string{registryName},
	}
	urlQuery.RawQuery = query.Encode()
	return urlQuery, nil
}

func GetPostureReportUrl(eventReceiverRestUrl, customerGUID, contextName, reportID string) (*url.URL, error) {
	scheme, host, err := utils.ParseHost(eventReceiverRestUrl)
	if err != nil {
		return nil, err
	}

	urlObj := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   v1.ReporterReportPath,
	}

	q := urlObj.Query()
	q.Add(v1.QueryParamCustomerGUID, uuid.MustParse(customerGUID).String())
	q.Add(v1.QueryParamContextName, contextName)
	q.Add(v1.QueryParamClusterName, contextName) // deprecated
	q.Add(v1.QueryParamReport, reportID)         // TODO - do we add the reportID?
	urlObj.RawQuery = q.Encode()

	return &urlObj, nil
}
