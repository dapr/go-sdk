/*
Copyright 2023 The Dapr Authors
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

package internal

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

const (
	dnsScheme          = "dns"
	unixScheme         = "unix"
	unixAbstractScheme = "unix-abstract"
	vsockScheme        = "vsock"
	httpScheme         = "http"
	httpsScheme        = "https"
)

// Parsed represents a parsed gRPC endpoint.
type Parsed struct {
	Target string
	TLS    bool
}

func ParseGRPCEndpoint(endpoint string) (Parsed, error) {
	target := endpoint
	if len(target) == 0 {
		return Parsed{}, errors.New("target is required")
	}

	var dnsAuthority string
	var hostname string
	var tls bool

	noSchemeInEndpoint := false
	urlSplit := strings.Split(target, ":")
	if len(urlSplit) == 3 && !strings.Contains(target, "://") {
		target = strings.Replace(target, ":", "://", 1)
	} else if len(urlSplit) >= 2 && !strings.Contains(target, "://") && schemeKnown(urlSplit[0]) {
		target = strings.Replace(target, ":", "://", 1)
	} else {
		urlSplit = strings.Split(target, "://")
		if len(urlSplit) == 1 {
			target = "dns://" + target
			noSchemeInEndpoint = true
		} else {
			scheme := urlSplit[0]
			if !schemeKnown(scheme) {
				return Parsed{}, fmt.Errorf(("unknown scheme: %q"), scheme)
			}

			if scheme == dnsScheme {
				urlSplit = strings.Split(target, "/")
				if len(urlSplit) < 4 {
					return Parsed{}, fmt.Errorf("invalid dns scheme: %q", target)
				}
				dnsAuthority = urlSplit[2]
				target = "dns://" + urlSplit[3]
			}
		}
	}

	ptarget, err := url.Parse(target)
	if err != nil {
		return Parsed{}, err
	}

	var errs []string
	for k := range ptarget.Query() {
		if k != "tls" {
			errs = append(errs, fmt.Sprintf("unrecognized query parameter: %q", k))
		}
	}
	if len(errs) > 0 {
		return Parsed{}, fmt.Errorf("failed to parse target %q: %s", target, strings.Join(errs, "; "))
	}

	if ptarget.Query().Has("tls") {
		if ptarget.Scheme == httpScheme || ptarget.Scheme == httpsScheme {
			return Parsed{}, errors.New("cannot use tls query parameter with http(s) scheme")
		}

		qtls := ptarget.Query().Get("tls")
		if qtls != "true" && qtls != "false" {
			return Parsed{}, fmt.Errorf("invalid value for tls query parameter: %q", qtls)
		}

		tls = qtls == "true"
	}

	scheme := ptarget.Scheme
	if scheme == httpsScheme {
		tls = true
	}
	if scheme == httpScheme || scheme == httpsScheme {
		scheme = dnsScheme
	}

	hostname = ptarget.Host

	host, port, err := net.SplitHostPort(hostname)
	aerr, ok := err.(*net.AddrError)
	if ok && aerr.Err == "missing port in address" {
		port = "443"
	} else if err != nil {
		return Parsed{}, err
	} else {
		hostname = host
	}

	if len(hostname) == 0 {
		if scheme == dnsScheme {
			hostname = "localhost"
		} else {
			hostname = ptarget.Path
		}
	}

	switch scheme {
	case unixScheme:
		separator := ":"
		if strings.HasPrefix(endpoint, "unix://") {
			separator = "://"
		}
		target = scheme + separator + hostname

	case vsockScheme:
		target = scheme + ":" + hostname + ":" + port

	case unixAbstractScheme:
		target = scheme + ":" + hostname

	case dnsScheme:
		if len(ptarget.Path) > 0 {
			return Parsed{}, fmt.Errorf("path is not allowed: %q", ptarget.Path)
		}

		if strings.Count(hostname, ":") == 7 && !strings.HasPrefix(hostname, "[") && !strings.HasSuffix(hostname, "]") {
			hostname = "[" + hostname + "]"
		}
		if len(dnsAuthority) > 0 {
			target = scheme + "://" + dnsAuthority + "/" + hostname + ":" + port
		} else if ptarget.Scheme == dnsScheme && !noSchemeInEndpoint {
			// when real scheme is dns, keep the previous format (dns:hostname:port)
			target = scheme + ":" + hostname + ":" + port
		} else {
			// Use passthrough resolver (just host:port) instead of dns: scheme.
			// gRPC v1.78's DNS resolver is broken with grpc.DialContext() — it
			// never fires regardless of URI format. The passthrough resolver
			// works correctly and grpc.DialContext() defaults to it for bare
			// host:port targets.
			target = hostname + ":" + port
		}

	default:
		return Parsed{}, fmt.Errorf("unsupported scheme: %q", scheme)
	}

	return Parsed{
		Target: target,
		TLS:    tls,
	}, nil
}

func schemeKnown(scheme string) bool {
	for _, s := range []string{
		dnsScheme,
		unixScheme,
		unixAbstractScheme,
		vsockScheme,
		httpScheme,
		httpsScheme,
	} {
		if scheme == s {
			return true
		}
	}

	return false
}
