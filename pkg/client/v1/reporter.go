package v1

import (
	"net/url"

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
