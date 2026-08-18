/*
Copyright 2026 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dapr/go-sdk/version"
)

// Anonymous usage reporting. The SDK reports SDK version, OS, architecture and
// Go version once per process so maintainers can see adoption; the Go module
// proxy reports nothing back to the project. No application data is collected.
// See the "Usage Analytics" section of the README, including how to opt out.
const (
	analyticsEndpoint = "https://dapr.gateway.scarf.sh/dapr-go-sdk"
	analyticsTimeout  = 2 * time.Second
)

var analyticsOnce sync.Once

// reportAnalytics sends a single usage event per process, in the background.
// It never blocks the caller and never surfaces an error: analytics must not
// affect application startup, and a blocked or air-gapped network is a normal
// condition, not a fault.
func reportAnalytics() {
	analyticsOnce.Do(func() {
		if analyticsDisabled() {
			return
		}
		go sendAnalyticsEvent()
	})
}

// analyticsDisabled honors the cross-ecosystem DO_NOT_TRACK convention, Scarf's
// own SCARF_NO_ANALYTICS variable, and a Dapr-specific opt-out.
func analyticsDisabled() bool {
	for _, name := range []string{"DO_NOT_TRACK", "SCARF_NO_ANALYTICS", "DAPR_DISABLE_ANALYTICS"} {
		if isTruthy(os.Getenv(name)) {
			return true
		}
	}
	return false
}

func isTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func sendAnalyticsEvent() {
	// A panic in telemetry must never reach the application.
	defer func() {
		_ = recover()
	}()

	endpoint, err := url.Parse(analyticsEndpoint)
	if err != nil {
		return
	}

	q := endpoint.Query()
	q.Set("version", strings.TrimSpace(version.SDKVersion))
	q.Set("os", runtime.GOOS)
	q.Set("arch", runtime.GOARCH)
	q.Set("go_version", runtime.Version())
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "dapr-go-sdk/"+strings.TrimSpace(version.SDKVersion))

	httpClient := &http.Client{Timeout: analyticsTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
