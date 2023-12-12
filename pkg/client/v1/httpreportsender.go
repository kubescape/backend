package v1

import (
	"fmt"
	"io"
	"net/http"
	"time"

	httputils "github.com/armosec/utils-go/httputils"
)

type IHttpSender interface {
	Send(serverURL string, headers map[string]string, reqBody []byte) (int, string, error)
}

type HttpReportSender struct {
	httpClient httputils.IHttpClient
}

// Send sends an HTTP request to a server and returns the HTTP status code, return message, and any errors.
func (s *HttpReportSender) Send(serverURL string, headers map[string]string, reqBody []byte) (int, string, error) {
	var resp *http.Response
	var err error
	var bodyAsStr string
	for i := 0; i < MAX_RETRIES; i++ {
		resp, err = httputils.HttpPost(s.httpClient, serverURL, headers, reqBody)
		bodyAsStr = "body could not be fetched"
		retry := err != nil
		if resp != nil {
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				retry = true
			}
			if resp.Body != nil {
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					bodyAsStr = string(body)
				}
				resp.Body.Close()
			}
		}
		if !retry {
			break
		}

		if i == MAX_RETRIES-1 {
			return 500, "", err
		}
		time.Sleep(RETRY_DELAY)
	}
	if resp == nil {
		return 500, bodyAsStr, fmt.Errorf("failed to send report, empty response: %w", err)
	}
	return resp.StatusCode, bodyAsStr, nil
}

type HttpReportSenderMock struct {
}

func (sm *HttpReportSenderMock) Send(serverURL string, headers map[string]string, reqBody []byte) (int, string, error) {
	return 200, "ok", nil
}
