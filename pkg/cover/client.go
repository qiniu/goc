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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Action provides methods to contact with the covered service under test
type Action interface {
	Profile(param ProfileParam) ([]byte, error)
	Clear() ([]byte, error)
	InitSystem() ([]byte, error)
	ListServices() ([]byte, error)
	RegisterService(svr Service) ([]byte, error)
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
	//CoverRegisterServiceAPI register a service into service center
	CoverRegisterServiceAPI = "/v1/cover/register"
)

type client struct {
	Host   string
	client *http.Client
}

// NewWorker creates a worker to contact with service
func NewWorker(host string) Action {
	_, err := url.ParseRequestURI(host)
	if err != nil {
		log.Fatalf("Parse url %s failed, err: %v", host, err)
	}
	return &client{
		Host:   host,
		client: http.DefaultClient,
	}
}

func (c *client) RegisterService(srv Service) ([]byte, error) {
	if _, err := url.ParseRequestURI(srv.Address); err != nil {
		return nil, err
	}
	if strings.TrimSpace(srv.Name) == "" {
		return nil, fmt.Errorf("invalid service name")
	}
	u := fmt.Sprintf("%s%s?name=%s&address=%s", c.Host, CoverRegisterServiceAPI, srv.Name, srv.Address)
	_, res, err := c.do("POST", u, "", nil)
	return res, err
}

func (c *client) ListServices() ([]byte, error) {
	u := fmt.Sprintf("%s%s", c.Host, CoverServicesListAPI)
	_, services, err := c.do("GET", u, "", nil)
	if err != nil && isNetworkError(err) {
		_, services, err = c.do("GET", u, "", nil)
	}

	return services, err
}

func (c *client) Profile(param ProfileParam) ([]byte, error) {
	u := fmt.Sprintf("%s%s", c.Host, CoverProfileAPI)
	if len(param.Service) != 0 && len(param.Address) != 0 {
		return nil, fmt.Errorf("use 'service' flag and 'address' flag at the same time may cause ambiguity, please use them separately")
	}

	body, err := json.Marshal(param)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed, param: %v, err:%v", param, err)
	}

	res, profile, err := c.do("POST", u, "application/json", bytes.NewReader(body))
	if err != nil && isNetworkError(err) {
		res, profile, err = c.do("POST", u, "application/json", bytes.NewReader(body))
	}

	if err == nil && res.StatusCode != 200 {
		err = fmt.Errorf(string(profile))
	}
	return profile, err
}

func (c *client) Clear() ([]byte, error) {
	u := fmt.Sprintf("%s%s", c.Host, CoverProfileClearAPI)
	_, resp, err := c.do("POST", u, "", nil)
	if err != nil && isNetworkError(err) {
		_, resp, err = c.do("POST", u, "", nil)
	}
	return resp, err
}

func (c *client) InitSystem() ([]byte, error) {
	u := fmt.Sprintf("%s%s", c.Host, CoverInitSystemAPI)
	_, body, err := c.do("POST", u, "", nil)
	return body, err
}

func (c *client) do(method, url, contentType string, body io.Reader) (*http.Response, []byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return res, nil, err
	}
	return res, responseBody, nil
}

func isNetworkError(err error) bool {
	if err == io.EOF {
		return true
	}
	_, ok := err.(net.Error)
	return ok
}
