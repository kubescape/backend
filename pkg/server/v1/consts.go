package v1

const (
	// API routes
	ApiServerAttackTracksPath                 = "/api/v1/attackTracks"
	ApiServerFrameworksPath                   = "/api/v1/frameworks"
	ApiServerExceptionsPath                   = "/api/v1/controlExceptions" // TODO: rename to controlExceptions
	ApiServerCustomerConfigPath               = "/api/v1/customerConfig"
	ApiServerVulnerabilitiesExceptionsPathOld = "/api/v1/armoVulnerabilityExceptions"
	ApiServerVulnerabilitiesExceptionsPath    = "/api/v1/vulnerabilityExceptions"

	// Reporter routes
	ReporterReportPath                  = "/k8s/v2/postureReport" // TODO: rename to postureReport
	ReporterVulnerabilitiesReportPath   = "/k8s/v2/containerScan"
	ReporterSystemReportPath            = "/k8s/sysreport"
	ReporterWebsocketClusterReportsPath = "/k8s/cluster-reports"

	// Gateway routes
	GatewayNotificationsPath = "/v1/waitfornotification"

	// default dummy account ID when not defined
	KubescapeFallbackCustomerGUID = "11111111-1111-1111-1111-111111111111"

	// URL query parameters
	QueryParamCustomerGUID        = "customerGUID"
	QueryParamScope               = "scope"
	QueryParamFrameworkName       = "frameworkName"
	QueryParamPolicyName          = "policyName"
	QueryParamClusterName         = "clusterName"
	QueryParamContextName         = "contextName"
	QueryParamReport              = "reportGUID"
	QueryParamJobID               = "jobID"
	QueryParamRegistryName        = "registryName"
	QueryParamGitRegoStoreVersion = "gitRegoStoreVersion"
	RegolibraryVersion            = "v2"

	AccessKeyHeader = "X-API-KEY"
)
