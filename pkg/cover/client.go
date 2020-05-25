/*
 Copyright 2020 Qiniu Cloud (七牛云)

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

package cover

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
)

// Action provides methods to contact with the coverd service under test
type Action interface {
	Profile(host string) ([]byte, error)
	Clear(host string) ([]byte, error)
	InitSystem(host string) ([]byte, error)
}

// CoverProfileAPI is provided by the covered service to get profiles
const CoverProfileAPI = "/v1/cover/profile"

// CoverProfileClearAPI is provided by the covered service to clear profiles
const CoverProfileClearAPI = "/v1/cover/clear"

// CoverInitSystemAPI prepare a new round of testing
const CoverInitSystemAPI = "/v1/cover/init"

type client struct {
	client *http.Client
}

// NewWorker creates a worker to contact with service
func NewWorker() Action {
	return &client{
		client: http.DefaultClient,
	}
}

func (w *client) Profile(host string) ([]byte, error) {
	u := fmt.Sprintf("%s%s", host, CoverProfileAPI)
	profile, err := w.do("GET", u, nil)
	if err != nil && isNetworkError(err) {
		profile, err = w.do("GET", u, nil)
	}

	return profile, err
}

func (w *client) Clear(host string) ([]byte, error) {
	u := fmt.Sprintf("%s%s", host, CoverProfileClearAPI)
	resp, err := w.do("POST", u, nil)
	if err != nil && isNetworkError(err) {
		resp, err = w.do("POST", u, nil)
	}
	return resp, err
}

func (w *client) InitSystem(host string) ([]byte, error) {
	u := fmt.Sprintf("%s%s", host, CoverInitSystemAPI)
	return w.do("POST", u, nil)
}

func (w *client) do(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	res, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func isNetworkError(err error) bool {
	if err == io.EOF {
		return true
	}
	_, ok := err.(net.Error)
	return ok
}
