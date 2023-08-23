package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func ReadString(rdr io.Reader, sizeHint int64) (string, error) {

	// if the response is empty, return an empty string
	if sizeHint < 0 {
		return "", nil
	}

	var b strings.Builder

	b.Grow(int(sizeHint))
	_, err := io.Copy(&b, rdr)

	return b.String(), err
}

// ParseHost picks a host from a hostname or an URL and detects the scheme.
//
// The default scheme is https. This may be altered by specifying an explicit http://hostname URL.
func ParseHost(host string) (string, string, error) {
	_, err := url.Parse(host)
	if err != nil {
		return "", "", err
	}

	if strings.HasPrefix(host, "ws://") {
		return "ws", strings.Replace(host, "ws://", "", 1), nil
	}

	if strings.HasPrefix(host, "wss://") {
		return "wss", strings.Replace(host, "wss://", "", 1), nil
	}

	if strings.HasPrefix(host, "http://") {
		return "http", strings.Replace(host, "http://", "", 1), nil
	}

	// default scheme
	return "https", strings.Replace(host, "https://", "", 1), nil
}

// ErrAPI reports an API error, with a cap on the length of the error message.
func ErrAPI(resp *http.Response) error {
	const maxSize = 1024

	reason := new(strings.Builder)
	if resp.Body != nil {
		size := min(resp.ContentLength, maxSize)
		if size > 0 {
			reason.Grow(int(size))
		}

		_, _ = io.CopyN(reason, resp.Body, size)
		defer resp.Body.Close()
	}

	return fmt.Errorf("http-error: '%s', reason: '%s'", resp.Status, reason.String())
}

func IsNativeFramework(framework string) bool {
	return contains([]string{"allcontrols", "nsa", "mitre"}, framework)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.EqualFold(v, str) {
			return true
		}
	}

	return false
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}
