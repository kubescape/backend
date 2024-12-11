package types

import "github.com/armosec/armoapi-go/armotypes"

type ClusterConfig struct {
	ClusterName         string `json:"clusterName"`         // cluster name defined manually or from the cluster context
	AccountID           string `json:"accountID"`           // use accountID instead of customerGUID
	GatewayWebsocketURL string `json:"gatewayWebsocketURL"` // in-cluster gateway component websocket url
	GatewayRestURL      string `json:"gatewayRestURL"`      // in-cluster gateway component REST API url
	KubevulnURL         string `json:"kubevulnURL"`         // in-cluster kubevuln component REST API url
	KubescapeURL        string `json:"kubescapeURL"`        // in-cluster kubescape component REST API url
	armotypes.InstallationData
}
