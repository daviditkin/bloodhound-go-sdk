// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package oapiclient

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
)

// Return an http.Client that resolves subdomain.localhost[:port] to localhost[:port] RFC 6761
func GetLocalhostWithSubdomainHttpClient() (*http.Client, *http.Client) {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	customDialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		domainOnly := addr[strings.LastIndex(addr, ".")+1:]
		if strings.HasPrefix(domainOnly, "localhost") {
			addr = domainOnly
		}
		return dialer.DialContext(ctx, network, addr)
	}

	customTransport := &http.Transport{
		DialContext: customDialContext,
	}

	return &http.Client{
		Transport: customTransport,
	}, nil

}