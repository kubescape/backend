package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildURL(t *testing.T) {
	t.Parallel()

	ks, err := NewKSCloudAPI(
		"api.example.com",
		"report.example.com",
		"",
		"",
	)
	require.NoError(t, err)

	t.Run("should build API URL with query params on https host", func(t *testing.T) {
		require.Equal(t,
			"https://api.example.com/path?q1=v1&q2=v2",
			ks.buildAPIURL("/path", "q1", "v1", "q2", "v2"),
		)
	})

	t.Run("should build API URL with query params on http host", func(t *testing.T) {
		ku, err := NewKSCloudAPI("http://api.example.com", "", "", "")

		require.NoError(t, err)
		require.Equal(t,
			"http://api.example.com/path?q1=v1&q2=v2",
			ku.buildAPIURL("/path", "q1", "v1", "q2", "v2"),
		)
	})

	t.Run("should panic when params are not provided in pairs", func(t *testing.T) {
		require.Panics(t, func() {
			// notice how the linter detects wrong args
			_ = ks.buildAPIURL("/path", "q1", "v1", "q2") //nolint:staticcheck
		})
	})

	t.Run("should build report URL with query params on https host", func(t *testing.T) {
		require.Equal(t,
			"https://report.example.com/path?q1=v1&q2=v2",
			ks.buildReportURL("/path", "q1", "v1", "q2", "v2"),
		)
	})
}
