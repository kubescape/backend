package versioncheck

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/semver"
)

func TestCheckLatestVersion_Semver_Compare(t *testing.T) {
	assert.Equal(t, -1, semver.Compare("v2.0.150", "v2.0.151"))
	assert.Equal(t, 0, semver.Compare("v2.0.150", "v2.0.150"))
	assert.Equal(t, 1, semver.Compare("v2.0.150", "v2.0.149"))
	assert.Equal(t, -1, semver.Compare("v2.0.150", "v3.0.150"))

}

func TestNormalizeVersion(t *testing.T) {
	assert.Equal(t, "v4.0.6", normalizeVersion("4.0.6"))
	assert.Equal(t, "v3.0.15", normalizeVersion("v3.0.15"))
	assert.Equal(t, "", normalizeVersion(""))
}

func TestSemverCompare_WithNormalization(t *testing.T) {
	result := semver.Compare(
		normalizeVersion("4.0.6"),
		normalizeVersion("v3.0.15"),
	)

	assert.Equal(t, 1, result)
}

func TestSemverCompare_WithoutNormalization(t *testing.T) {
	result := semver.Compare("4.0.6", "v3.0.15")

	assert.Equal(t, -1, result)
}

func TestCheckLatestVersion(t *testing.T) {
	type args struct {
		ctx         context.Context
		versionData *VersionCheckRequest
		versionURL  string
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "Get latest version",
			args: args{
				ctx:         context.Background(),
				versionData: &VersionCheckRequest{},
				versionURL:  "https://version-check.ks-services.co",
			},
			err: nil,
		},
		{
			name: "Failed to get latest version",
			args: args{
				ctx:         context.Background(),
				versionData: &VersionCheckRequest{},
				versionURL:  "https://example.com",
			},
			err: fmt.Errorf("failed to get latest version"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VersionCheckHandler{
				versionURL: tt.args.versionURL,
			}
			err := v.CheckLatestVersion(tt.args.ctx, tt.args.versionData)

			assert.Equal(t, tt.err, err)
		})
	}
}

func TestVersionCheckHandler_getLatestVersion(t *testing.T) {
	t.Run("Get latest version", func(t *testing.T) {
		// Live canary against the deployed version-check service. A request
		// with an empty clientVersion MUST be told the latest version -- this
		// guards the ksgf1v1 regression where empty/unparseable client
		// versions silently received no clientUpdate. We assert a *valid*
		// version rather than a hardcoded one so routine kubescape releases
		// (which bump the reference) don't break this test.
		v := &VersionCheckHandler{
			versionURL: "https://version-check.ks-services.co",
		}
		got, err := v.getLatestVersion(&VersionCheckRequest{Client: "kubescape"})
		if err != nil {
			t.Fatalf("getLatestVersion() unexpected error = %v", err)
		}
		if got == nil {
			t.Fatal("getLatestVersion() returned a nil response")
		}
		assert.Equal(t, "kubescape", got.Client)
		assert.NotEmpty(t, got.ClientUpdate, "empty clientVersion must be told the latest version")
		assert.True(t, semver.IsValid(normalizeVersion(got.ClientUpdate)),
			"ClientUpdate %q must be a valid semantic version", got.ClientUpdate)
	})

	t.Run("Failed to get latest version", func(t *testing.T) {
		v := &VersionCheckHandler{
			versionURL: "https://example.com",
		}
		got, err := v.getLatestVersion(&VersionCheckRequest{})
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestGetTriggerSource(t *testing.T) {
	// Running in github actions pipeline
	os.Setenv("GITHUB_ACTIONS", "true")
	source := getTriggerSource()
	assert.Equal(t, "pipeline", source)

	os.Args[0] = "ksserver"
	source = getTriggerSource()
	assert.Equal(t, "microservice", source)
}
