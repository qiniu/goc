/*
 Copyright 2020 Qiniu Cloud (qiniu.com)
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

package websocket

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// Dialer contains all options for connecting to a specified websocket server
type Dialer struct {
	Proxy            func(*http.Request) (*url.URL, error)
	HandshakeTimeout time.Duration
	Subprotocols     []string
}

// DefaultDialer is dialer with all necessary fields set to default value
var DefaultDialer = &Dialer{
	Proxy:            http.ProxyFromEnvironment,
	HandshakeTimeout: 45 * time.Second,
}

func (d *Dialer) Dial() {

}

func (d *Dialer) DialContext(ctx context.Context) {

}
