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

package cover

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
)

// Action provides methods to contact with the covered service under test
type Action interface {
	Profile(host string) ([]byte, error)
	Clear(host string) ([]byte, error)
	InitSystem(host string) ([]byte, error)
	ListServices(host string) ([]byte, error)
}

const (
	//CoverInitSystemAPI prepare a new round of testing
	CoverInitSystemAPI = "/v1/cover/init"
	//CoverProfileAPI is provided by the covered service to get profiles
	CoverProfileAPI = "/v1/cover/profile"
	//CoverProfileClearAPI is provided by the covered service to clear profiles
	CoverProfileClearAPI = "/v1/cover/clear"
	//CoverServicesListAPI list all the registered services
	CoverServicesListAPI = "/v1/cover/list"
)

type client struct {
	client *http.Client
}

// NewWorker creates a worker to contact with service
func NewWorker() Action {
	return &client{
		client: http.DefaultClient,
	}
}

func (c *client) ListServices(host string) ([]byte, error) {
	u := fmt.Sprintf("%s%s", host, CoverServicesListAPI)
	services, err := c.do("GET", u, nil)
	if err != nil && isNetworkError(err) {
		services, err = c.do("GET", u, nil)
	}

	return services, err
}

func (c *client) Profile(host string) ([]byte, error) {
	u := fmt.Sprintf("%s%s", host, CoverProfileAPI)
	profile, err := c.do("GET", u, nil)
	if err != nil && isNetworkError(err) {
		profile, err = c.do("GET", u, nil)
	}

	return profile, err
}

func (c *client) Clear(host string) ([]byte, error) {
	u := fmt.Sprintf("%s%s", host, CoverProfileClearAPI)
	resp, err := c.do("POST", u, nil)
	if err != nil && isNetworkError(err) {
		resp, err = c.do("POST", u, nil)
	}
	return resp, err
}

func (c *client) InitSystem(host string) ([]byte, error) {
	u := fmt.Sprintf("%s%s", host, CoverInitSystemAPI)
	return c.do("POST", u, nil)
}

func (c *client) do(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
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
