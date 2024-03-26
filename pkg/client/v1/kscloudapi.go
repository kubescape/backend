package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	v1 "github.com/kubescape/backend/pkg/server/v1"
	"github.com/kubescape/backend/pkg/utils"
)

var (
	ErrAPINotPublic = errors.New("control api is not public")
)

// KSCloudAPI allows to access the API of the Kubescape Cloud offering.
type KSCloudAPI struct {
	*KsCloudOptions
	accountID    string
	accessKey    string
	apiHost      string
	apiScheme    string
	reportHost   string
	reportScheme string
}

// NewEmptyKSCloudAPI creates a new KSCloudAPI without any hosts set.
func NewEmptyKSCloudAPI(opts ...KSCloudOption) *KSCloudAPI {
	api := &KSCloudAPI{
		KsCloudOptions: ksCloudOptionsWithDefaults(opts),
	}

	return api
}

func NewKSCloudAPI(apiURL, reportURL, accountID, accessKey string, opts ...KSCloudOption) (*KSCloudAPI, error) {
	api := &KSCloudAPI{
		KsCloudOptions: ksCloudOptionsWithDefaults(opts),
		accountID:      accountID,
		accessKey:      accessKey,
	}

	if err := api.SetCloudAPIURL(apiURL); err != nil {
		return nil, err
	}

	if err := api.SetCloudReportURL(reportURL); err != nil {
		return nil, err
	}

	return api, nil
}

func (api *KSCloudAPI) SetAccountID(value string) {
	api.accountID = value
}

func (api *KSCloudAPI) SetAccessKey(value string) {
	api.accessKey = value
}

// GetAccountID returns the customer account's GUID.
func (api *KSCloudAPI) GetAccountID() string { return api.accountID }

func (api *KSCloudAPI) GetAccessKey() string { return api.accessKey }

func (api *KSCloudAPI) GetCloudReportURL() string {
	if api.reportHost == "" {
		return ""
	}

	return api.reportScheme + "://" + api.reportHost
}

func (api *KSCloudAPI) GetCloudAPIURL() string {
	if api.apiHost == "" {
		return ""
	}
	return api.apiScheme + "://" + api.apiHost
}

func (api *KSCloudAPI) SetCloudAPIURL(cloudAPIURL string) (err error) {
	if cloudAPIURL == "" {
		return nil
	}
	api.apiScheme, api.apiHost, err = utils.ParseHost(cloudAPIURL)
	return err
}

func (api *KSCloudAPI) SetCloudReportURL(cloudReportURL string) (err error) {
	if cloudReportURL == "" {
		return nil
	}

	api.reportScheme, api.reportHost, err = utils.ParseHost(cloudReportURL)
	return err
}

func (api *KSCloudAPI) GetAttackTracks() ([]AttackTrack, error) {
	rdr, _, err := api.get(api.getAttackTracksURL())
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	attackTracks, err := utils.Decode[[]AttackTrack](rdr)
	if err != nil {
		return nil, err
	}

	return attackTracks, nil
}

func (api *KSCloudAPI) getAttackTracksURL() string {
	return api.buildAPIURL(
		v1.ApiServerAttackTracksPath,
		append(
			api.paramsWithGUID(),
			v1.QueryParamGitRegoStoreVersion, v1.RegolibraryVersion,
		)...,
	)
}

// GetFramework retrieves a framework by name.
func (api *KSCloudAPI) GetFramework(frameworkName string) (*Framework, error) {
	rdr, _, err := api.get(api.getFrameworkURL(frameworkName))
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	framework, err := utils.Decode[Framework](rdr)
	if err != nil {
		return nil, err
	}

	return &framework, err
}

func (api *KSCloudAPI) getFrameworkURL(frameworkName string) string {
	if utils.IsNativeFramework(frameworkName) {
		// Native framework name is normalized as upper case, but for a custom framework the name remains unaltered
		frameworkName = strings.ToUpper(frameworkName)
	}

	return api.buildAPIURL(
		v1.ApiServerFrameworksPath,
		append(
			api.paramsWithGUID(),
			v1.QueryParamFrameworkName, frameworkName,
			v1.QueryParamGitRegoStoreVersion, v1.RegolibraryVersion,
		)...,
	)
}

// GetFrameworks returns all registered frameworks.
func (api *KSCloudAPI) GetFrameworks() ([]Framework, error) {
	rdr, _, err := api.get(api.getListFrameworkURL())
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	frameworks, err := utils.Decode[[]Framework](rdr)
	if err != nil {
		return nil, err
	}

	return frameworks, err
}

func (api *KSCloudAPI) getListFrameworkURL() string {
	return api.buildAPIURL(
		v1.ApiServerFrameworksPath,
		append(
			api.paramsWithGUID(),
			v1.QueryParamGitRegoStoreVersion, v1.RegolibraryVersion,
		)...,
	)
}

// ListCustomFrameworks lists the names of all non-native frameworks that have been registered for this account.
func (api *KSCloudAPI) ListCustomFrameworks() ([]string, error) {
	frameworks, err := api.GetFrameworks()
	if err != nil {
		return nil, err
	}

	frameworkList := make([]string, 0, len(frameworks))
	for _, framework := range frameworks {
		if utils.IsNativeFramework(framework.Name) {
			continue
		}

		frameworkList = append(frameworkList, framework.Name)
	}

	return frameworkList, nil
}

// ListFrameworks list the names of all registered frameworks.
func (api *KSCloudAPI) ListFrameworks() ([]string, error) {
	frameworks, err := api.GetFrameworks()
	if err != nil {
		return nil, err
	}

	frameworkList := make([]string, 0, len(frameworks))
	for _, framework := range frameworks {
		name := framework.Name
		if utils.IsNativeFramework(framework.Name) {
			name = strings.ToLower(framework.Name)
		}

		frameworkList = append(frameworkList, name)
	}

	return frameworkList, nil
}

// GetExceptions returns exception policies.
func (api *KSCloudAPI) GetExceptions(clusterName string) ([]PostureExceptionPolicy, error) {
	rdr, _, err := api.get(api.getExceptionsURL(clusterName))
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	exceptions, err := utils.Decode[[]PostureExceptionPolicy](rdr)
	if err != nil {
		return nil, err
	}

	return exceptions, nil
}

func (api *KSCloudAPI) getExceptionsURL(clusterName string) string {
	return api.buildAPIURL(
		v1.ApiServerExceptionsPath,
		append(
			api.paramsWithGUID(),
			v1.QueryParamGitRegoStoreVersion, v1.RegolibraryVersion,
		)...,
	)
	// queryParamClusterName, clusterName, // TODO - fix customer name support in Armo BE
}

// GetAccountConfig yields the account configuration.
func (api *KSCloudAPI) GetAccountConfig(clusterName string) (*CustomerConfig, error) {
	if api.accountID == "" {
		return &CustomerConfig{}, nil
	}

	rdr, _, err := api.get(api.getAccountConfig(clusterName))
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	accountConfig, err := utils.Decode[CustomerConfig](rdr)
	if err != nil {
		// retry with default scope
		rdr, _, err = api.get(api.getAccountConfigDefault(clusterName))
		if err != nil {
			return nil, err
		}
		defer rdr.Close()

		accountConfig, err = utils.Decode[CustomerConfig](rdr)
		if err != nil {
			return nil, err
		}
	}

	return &accountConfig, nil
}

func (api *KSCloudAPI) getAccountConfig(clusterName string) string {
	params := api.paramsWithGUID()

	if clusterName != "" { // TODO - fix customer name support in Armo BE
		params = append(params, v1.QueryParamClusterName, clusterName)
	}

	return api.buildAPIURL(
		v1.ApiServerCustomerConfigPath,
		append(
			params,
			v1.QueryParamGitRegoStoreVersion, v1.RegolibraryVersion,
		)...,
	)
}

func (api *KSCloudAPI) getAccountConfigDefault(clusterName string) string {
	params := append(
		api.paramsWithGUID(),
		v1.QueryParamScope, "customer",
	)

	if clusterName != "" { // TODO - fix customer name support in Armo BE
		params = append(params, v1.QueryParamClusterName, clusterName)
	}

	return api.buildAPIURL(
		v1.ApiServerCustomerConfigPath,
		append(
			params,
			v1.QueryParamGitRegoStoreVersion, v1.RegolibraryVersion,
		)...,
	)
}

// GetControlsInputs returns the controls inputs configured in the account configuration.
func (api *KSCloudAPI) GetControlsInputs(clusterName string) (map[string][]string, error) {
	accountConfig, err := api.GetAccountConfig(clusterName)
	if err != nil {
		return nil, err
	}

	return accountConfig.Settings.PostureControlInputs, nil
}

// GetControl is currently not exposed as a public API endpoint.
func (api *KSCloudAPI) GetControl(ID string) (*Control, error) {
	return nil, ErrAPINotPublic
}

// ListControls is currently not exposed as a public API endpoint.
func (api *KSCloudAPI) ListControls() ([]string, error) {
	return nil, ErrAPINotPublic
}

// SubmitReport uploads a posture report.
func (api *KSCloudAPI) SubmitReport(report *PostureReport) (string, error) {
	jazon, err := json.Marshal(report)
	if err != nil {
		return "", err
	}

	rdr, _, err := api.post(api.postReportURL(report.ClusterName, report.ReportID), jazon, WithContentJSON(true))
	if err != nil {
		return "", err
	}
	defer rdr.Close()

	b, err := io.ReadAll(rdr)
	if err == nil {
		return string(b), nil
	}
	return "", err
}

func (api *KSCloudAPI) postReportURL(cluster, reportID string) string {
	return api.buildReportURL(v1.ReporterReportPath,
		append(
			api.paramsWithGUID(),
			v1.QueryParamContextName, cluster,
			v1.QueryParamClusterName, cluster, // deprecated
			v1.QueryParamReport, reportID,
		)...,
	)
}

// defaultRequestOptions adds standard authentication headers to all requests
func (api *KSCloudAPI) defaultRequestOptions(opts []RequestOption) *RequestOptions {
	optionsWithDefaults := []RequestOption{
		withTrace(api.withTrace),
		WithContentJSON(true),
	}

	if api.accessKey != "" {
		optionsWithDefaults = append(optionsWithDefaults,
			WithHeaders(map[string]string{
				v1.AccessKeyHeader: api.accessKey,
			}))
	}

	optionsWithDefaults = append(optionsWithDefaults, opts...)

	return requestOptionsWithDefaults(optionsWithDefaults)
}

func (api *KSCloudAPI) get(fullURL string, opts ...RequestOption) (io.ReadCloser, int64, error) {
	o := api.defaultRequestOptions(opts)
	req, err := http.NewRequestWithContext(o.reqContext, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, err
	}

	return api.do(req, o)
}

func (api *KSCloudAPI) post(fullURL string, body []byte, opts ...RequestOption) (io.ReadCloser, int64, error) {
	o := api.defaultRequestOptions(opts)
	req, err := http.NewRequestWithContext(o.reqContext, http.MethodPost, fullURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}

	return api.do(req, o)
}

// func (api *KSCloudAPI) delete(fullURL string, opts ...RequestOption) (io.ReadCloser, int64, error) {
// 	o := api.defaultRequestOptions(opts)
// 	req, err := http.NewRequestWithContext(o.reqContext, http.MethodDelete, fullURL, nil)
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	return api.do(req, o)
// }

func (api *KSCloudAPI) do(req *http.Request, o *RequestOptions) (io.ReadCloser, int64, error) {
	o.setHeaders(req)
	o.traceReq(req)

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	o.traceResp(resp)

	if resp.StatusCode >= 400 {
		return nil, 0, utils.ErrAPI(resp)
	}

	return resp.Body, resp.ContentLength, err
}

func (api *KSCloudAPI) paramsWithGUID() []string {
	return append(make([]string, 0, 6),
		v1.QueryParamCustomerGUID, api.getCustomerGUIDFallBack(),
	)
}

func (api *KSCloudAPI) getCustomerGUIDFallBack() string {
	if api.accountID != "" {
		return api.accountID
	}
	return v1.KubescapeFallbackCustomerGUID
}
