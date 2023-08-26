package v1

import (
	"net/url"

	v1 "github.com/kubescape/backend/pkg/server/v1"
	"github.com/kubescape/backend/pkg/utils"
)

func GetRootGatewayUrl(gatewayUrl string) (*url.URL, error) {
	scheme, host, err := utils.ParseHost(gatewayUrl)
	if err != nil {
		return nil, err
	}
	// if no scheme is specified, calling ParseHost default to https, so we need to change it to wss
	if scheme == "https" {
		scheme = "wss"
	}
	urlBase := &url.URL{
		Host:   host,
		Scheme: scheme,
		Path:   v1.GatewayNotificationsPath,
	}
	return urlBase, nil
}
