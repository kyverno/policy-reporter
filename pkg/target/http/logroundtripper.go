package http

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"go.uber.org/zap"
)

func NewLoggingRoundTripper(roundTripper http.RoundTripper) http.RoundTripper {
	return &logRoundTripper{roundTripper: roundTripper}
}

type logRoundTripper struct {
	roundTripper http.RoundTripper
}

func (rt *logRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	logger := zap.L()
	if logger.Core().Enabled(zap.DebugLevel) {
		if info, err := httputil.DumpRequest(req, true); err == nil {
			logger.Debug(fmt.Sprintf("Sending request: %s", string(info)))
			if err != nil {
				return nil, err
			}
		}
	}
	resp, err := rt.roundTripper.RoundTrip(req)
	if resp != nil {
		if logger.Core().Enabled(zap.DebugLevel) {
			if info, err := httputil.DumpResponse(resp, true); err == nil {
				logger.Debug(fmt.Sprintf("Received response: %s", string(info)))
			}
		}
	}
	return resp, err
}
