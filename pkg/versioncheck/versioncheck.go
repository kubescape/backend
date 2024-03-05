package versioncheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/armosec/utils-go/boolutils"
	"github.com/kubescape/backend/pkg/utils"
	"github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
	"github.com/kubescape/kubescape/v3/core/cautils/getter"
	"github.com/mattn/go-isatty"
	"go.opentelemetry.io/otel"
	"golang.org/x/mod/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const SKIP_VERSION_CHECK_DEPRECATED_ENV = "KUBESCAPE_SKIP_UPDATE_CHECK"
const SKIP_VERSION_CHECK_ENV = "KS_SKIP_UPDATE_CHECK"
const CLIENT_ENV = "KS_CLIENT"

var BuildNumber string
var Client string
var LatestReleaseVersion string

const UnknownBuildNumber = "unknown"

type IVersionCheckHandler interface {
	CheckLatestVersion(context.Context, *VersionCheckRequest) error
}

func NewIVersionCheckHandler(ctx context.Context) IVersionCheckHandler {
	if BuildNumber == "" {
		logger.L().Ctx(ctx).Warning("Unknown build number: this might affect your scan results. Please ensure that you are running the latest version.")
	}

	if v, ok := os.LookupEnv(CLIENT_ENV); ok && v != "" {
		Client = v
	}

	if v, ok := os.LookupEnv(SKIP_VERSION_CHECK_ENV); ok && boolutils.StringToBool(v) {
		return NewVersionCheckHandlerMock()
	} else if v, ok := os.LookupEnv(SKIP_VERSION_CHECK_DEPRECATED_ENV); ok && boolutils.StringToBool(v) {
		return NewVersionCheckHandlerMock()
	}
	return NewVersionCheckHandler()
}

type VersionCheckHandlerMock struct {
}

func NewVersionCheckHandlerMock() *VersionCheckHandlerMock {
	return &VersionCheckHandlerMock{}
}

type VersionCheckHandler struct {
	versionURL string
}
type VersionCheckRequest struct {
	AccountID        string `json:"accountID"`        // account id
	Client           string `json:"client"`           // kubescape
	ClientBuild      string `json:"clientBuild"`      // client build environment
	ClientVersion    string `json:"clientVersion"`    // kubescape version
	ClusterID        string `json:"clusterID"`        // cluster id
	Framework        string `json:"framework"`        // framework name
	FrameworkVersion string `json:"frameworkVersion"` // framework version
	HelmChartVersion string `json:"helmChartVersion"` // helm chart version
	Nodes            int    `json:"nodes"`            // number of nodes
	ScanningTarget   string `json:"target"`           // Deprecated
	ScanningContext  string `json:"context"`          // scanning context- cluster/file/gitURL/localGit/dir
	TriggeredBy      string `json:"triggeredBy"`      // triggered by - cli/ ci / microservice
}

type VersionCheckResponse struct {
	Client          string `json:"client"`          // kubescape
	ClientUpdate    string `json:"clientUpdate"`    // kubescape latest version
	Framework       string `json:"framework"`       // framework name
	FrameworkUpdate string `json:"frameworkUpdate"` // framework latest version
	Message         string `json:"message"`         // alert message
}

func NewVersionCheckHandler() *VersionCheckHandler {
	return &VersionCheckHandler{
		versionURL: "https://version-check.ks-services.co",
	}
}

func getTriggerSource() string {
	if strings.Contains(os.Args[0], "ksserver") {
		return "microservice"
	}

	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		// non-interactive shell
		return "pipeline"
	}

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return "pipeline"
	}

	return "cli"
}

func NewVersionCheckRequest(accountID, buildNumber, frameworkName, frameworkVersion, scanningContext string, k8sClient kubernetes.Interface) *VersionCheckRequest {
	if buildNumber == "" {
		buildNumber = UnknownBuildNumber
	}

	if scanningContext == "" {
		scanningContext = "unknown"
	}

	if Client == "" {
		Client = "local-build"
	}

	return &VersionCheckRequest{
		AccountID:        accountID,
		Client:           "kubescape",
		ClientBuild:      Client,
		ClientVersion:    buildNumber,
		ClusterID:        generateClusterID(k8sClient),
		Framework:        frameworkName,
		FrameworkVersion: frameworkVersion,
		HelmChartVersion: getHelmChartVersion(),
		Nodes:            getNodeCount(k8sClient),
		ScanningContext:  scanningContext,
		TriggeredBy:      getTriggerSource(),
	}
}

// copilot suggests to use the uid of a service to generate a cluster id
func generateClusterID(k8sClient kubernetes.Interface) string {
	if k8sClient == nil {
		return ""
	}
	svc, err := k8sClient.CoreV1().Services("default").Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err != nil {
		return ""
	}
	return string(svc.UID)
}

func getHelmChartVersion() string {
	return os.Getenv("HELM_RELEASE")
}

func getNodeCount(k8sClient kubernetes.Interface) int {
	if k8sClient == nil {
		return 0
	}
	list, err := k8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0
	}
	return len(list.Items)
}

func (v *VersionCheckHandlerMock) CheckLatestVersion(_ context.Context, _ *VersionCheckRequest) error {
	logger.L().Info("Skipping version check")
	return nil
}

func (v *VersionCheckHandler) CheckLatestVersion(ctx context.Context, versionData *VersionCheckRequest) error {
	ctx, span := otel.Tracer("").Start(ctx, "versionCheckHandler.CheckLatestVersion")
	defer span.End()
	defer func() {
		if err := recover(); err != nil {
			logger.L().Ctx(ctx).Warning("failed to get latest version", helpers.Interface("error", err))
		}
	}()

	latestVersion, err := v.getLatestVersion(versionData)
	if err != nil || latestVersion == nil {
		return fmt.Errorf("failed to get latest version")
	}

	LatestReleaseVersion = latestVersion.ClientUpdate

	if latestVersion.ClientUpdate != "" {
		if BuildNumber != "" && semver.Compare(BuildNumber, LatestReleaseVersion) == -1 {
			logger.L().Ctx(ctx).Warning(warningMessage(LatestReleaseVersion))
		}
	}

	// TODO - Enable after supporting framework version
	// if latestVersion.FrameworkUpdate != "" {
	// 	fmt.Println(warningMessage(latestVersion.Framework, latestVersion.FrameworkUpdate))
	// }

	if latestVersion.Message != "" {
		logger.L().Info(latestVersion.Message)
	}

	return nil
}

func (v *VersionCheckHandler) getLatestVersion(versionData *VersionCheckRequest) (*VersionCheckResponse, error) {

	reqBody, err := json.Marshal(*versionData)
	if err != nil {
		return nil, fmt.Errorf("in 'CheckLatestVersion' failed to json.Marshal, reason: %s", err.Error())
	}

	rdr, _, err := getter.HTTPPost(http.DefaultClient, v.versionURL, reqBody, map[string]string{"Content-Type": "application/json"})

	vResp, err := utils.Decode[*VersionCheckResponse](rdr)
	if err != nil {
		return nil, err
	}

	return vResp, nil
}

func warningMessage(release string) string {
	return fmt.Sprintf("current version '%s' is not updated to the latest release: '%s'", BuildNumber, release)
}
