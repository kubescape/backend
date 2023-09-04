package v1

import (
	"fmt"
	"io"
	"net/http"
	"time"

	httputils "github.com/armosec/utils-go/httputils"
)

type IHttpSender interface {
	Send(serverURL string, reqBody []byte) (int, string, error)
}

type HttpReportSender struct {
	httpClient httputils.IHttpClient
}

// Send - send http request. returns-> http status code, return message (jobID/OK), http/go error
func (s *HttpReportSender) Send(serverURL string, reqBody []byte) (int, string, error) {

	var resp *http.Response
	var bodyAsStr string
	for i := 0; i < MAX_RETRIES; i++ {
		resp, err := httputils.HttpPost(s.httpClient, serverURL, map[string]string{"Content-Type": "application/json"}, reqBody)
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
		//else err != nil
		e := fmt.Errorf("attempt #%d - Failed posting report. Url: '%s', reason: '%s' report: '%s' response: '%s'", i, serverURL, err.Error(), string(reqBody), bodyAsStr)

		if i == MAX_RETRIES-1 {
			return 500, e.Error(), err
		}
		//wait 5 secs between retries
		time.Sleep(RETRY_DELAY)
	}
	return resp.StatusCode, bodyAsStr, nil

}

type HttpReportSenderMock struct {
}

func (sm *HttpReportSenderMock) Send(serverURL string, reqBody []byte) (int, string, error) {
	return 200, "ok", nil
}
