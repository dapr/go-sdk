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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		expTarget string
		expTLS    bool
		expError  bool
	}{
		"": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		":5000": {
			expTarget: "dns:localhost:5000",
			expTLS:    false,
			expError:  false,
		},
		":5000?tls=false": {
			expTarget: "dns:localhost:5000",
			expTLS:    false,
			expError:  false,
		},
		":5000?tls=true": {
			expTarget: "dns:localhost:5000",
			expTLS:    true,
			expError:  false,
		},
		"myhost": {
			expTarget: "dns:myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"myhost?tls=false": {
			expTarget: "dns:myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"myhost?tls=true": {
			expTarget: "dns:myhost:443",
			expTLS:    true,
			expError:  false,
		},
		"myhost:443": {
			expTarget: "dns:myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"myhost:443?tls=false": {
			expTarget: "dns:myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"myhost:443?tls=true": {
			expTarget: "dns:myhost:443",
			expTLS:    true,
			expError:  false,
		},
		"http://myhost": {
			expTarget: "dns:myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"http://myhost?tls=false": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"http://myhost?tls=true": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"http://myhost:443": {
			expTarget: "dns:myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"http://myhost:443?tls=false": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"http://myhost:443?tls=true": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"http://myhost:5000": {
			expTarget: "dns:myhost:5000",
			expTLS:    false,
			expError:  false,
		},
		"http://myhost:5000?tls=false": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"http://myhost:5000?tls=true": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"https://myhost:443": {
			expTarget: "dns:myhost:443",
			expTLS:    true,
			expError:  false,
		},
		"https://myhost:443/tls=false": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"https://myhost:443?tls=true": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"dns:myhost": {
			expTarget: "dns:myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"dns:myhost?tls=false": {
			expTarget: "dns:myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"dns:myhost?tls=true": {
			expTarget: "dns:myhost:443",
			expTLS:    true,
			expError:  false,
		},
		"dns://myauthority:53/myhost": {
			expTarget: "dns://myauthority:53/myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"dns://myauthority:53/myhost?tls=false": {
			expTarget: "dns://myauthority:53/myhost:443",
			expTLS:    false,
			expError:  false,
		},
		"dns://myauthority:53/myhost?tls=true": {
			expTarget: "dns://myauthority:53/myhost:443",
			expTLS:    true,
			expError:  false,
		},
		"dns://myhost": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"unix:my.sock": {
			expTarget: "unix:my.sock",
			expTLS:    false,
			expError:  false,
		},
		"unix:my.sock?tls=true": {
			expTarget: "unix:my.sock",
			expTLS:    true,
			expError:  false,
		},
		"unix://my.sock": {
			expTarget: "unix://my.sock",
			expTLS:    false,
			expError:  false,
		},
		"unix:///my.sock": {
			expTarget: "unix:///my.sock",
			expTLS:    false,
			expError:  false,
		},
		"unix://my.sock?tls=true": {
			expTarget: "unix://my.sock",
			expTLS:    true,
			expError:  false,
		},
		"unix-abstract:my.sock": {
			expTarget: "unix-abstract:my.sock",
			expTLS:    false,
			expError:  false,
		},
		"unix-abstract:my.sock?tls=false": {
			expTarget: "unix-abstract:my.sock",
			expTLS:    false,
			expError:  false,
		},
		"unix-abstract:my.sock?tls=true": {
			expTarget: "unix-abstract:my.sock",
			expTLS:    true,
			expError:  false,
		},
		"vsock:mycid:5000": {
			expTarget: "vsock:mycid:5000",
			expTLS:    false,
			expError:  false,
		},
		"vsock:mycid:5000?tls=false": {
			expTarget: "vsock:mycid:5000",
			expTLS:    false,
			expError:  false,
		},
		"vsock:mycid:5000?tls=true": {
			expTarget: "vsock:mycid:5000",
			expTLS:    true,
			expError:  false,
		},
		"dns:1.2.3.4:443": {
			expTarget: "dns:1.2.3.4:443",
			expTLS:    false,
			expError:  false,
		},
		"dns:[2001:db8:1f70::999:de8:7648:6e8]:443": {
			expTarget: "dns:[2001:db8:1f70::999:de8:7648:6e8]:443",
			expTLS:    false,
			expError:  false,
		},
		"dns:[2001:db8:1f70::999:de8:7648:6e8]:5000": {
			expTarget: "dns:[2001:db8:1f70::999:de8:7648:6e8]:5000",
			expTLS:    false,
			expError:  false,
		},
		"dns:[2001:db8:1f70::999:de8:7648:6e8]:5000?abc=[]": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"dns://myauthority:53/[2001:db8:1f70::999:de8:7648:6e8]": {
			expTarget: "dns://myauthority:53/[2001:db8:1f70::999:de8:7648:6e8]:443",
			expTLS:    false,
			expError:  false,
		},
		"https://[2001:db8:1f70::999:de8:7648:6e8]": {
			expTarget: "dns:[2001:db8:1f70::999:de8:7648:6e8]:443",
			expTLS:    true,
			expError:  false,
		},
		"https://[2001:db8:1f70::999:de8:7648:6e8]:5000": {
			expTarget: "dns:[2001:db8:1f70::999:de8:7648:6e8]:5000",
			expTLS:    true,
			expError:  false,
		},
		"host:5000/v1/dapr": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"host:5000/?a=1": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"inv-scheme://myhost": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
		"inv-scheme:myhost:5000": {
			expTarget: "",
			expTLS:    false,
			expError:  true,
		},
	}

	for url, tc := range tests {
		t.Run(url, func(t *testing.T) {
			parsed, err := ParseGRPCEndpoint(url)
			assert.Equalf(t, tc.expError, err != nil, "%v", err)
			assert.Equal(t, tc.expTarget, parsed.Target)
			assert.Equal(t, tc.expTLS, parsed.TLS)
		})
	}
}
