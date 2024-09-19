package versioncheck

import (
	"context"
	"fmt"
	"os"
	"reflect"
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
	type fields struct {
		versionURL string
	}
	type args struct {
		versionData *VersionCheckRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *VersionCheckResponse
		wantErr bool
	}{
		{
			name: "Get latest version",
			fields: fields{
				versionURL: "https://version-check.ks-services.co",
			},
			args: args{
				versionData: &VersionCheckRequest{
					Client: "kubescape",
				},
			},
			want: &VersionCheckResponse{
				Client:       "kubescape",
				ClientUpdate: "v3.0.15",
			},
			wantErr: false,
		},
		{
			name: "Failed to get latest version",
			fields: fields{
				versionURL: "https://example.com",
			},
			args: args{
				versionData: &VersionCheckRequest{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VersionCheckHandler{
				versionURL: tt.fields.versionURL,
			}
			got, err := v.getLatestVersion(tt.args.versionData)
			if (err != nil) != tt.wantErr {
				t.Errorf("VersionCheckHandler.getLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VersionCheckHandler.getLatestVersion() = %v, want %v", got, tt.want)
			}
		})
	}
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
