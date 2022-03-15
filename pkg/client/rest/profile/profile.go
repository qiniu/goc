/*
 Copyright 2021 Qiniu Cloud (qiniu.com)
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

package profile

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Profile struct {
	Profile string `json:"profile"`
}

const (
	profileAPI = "/v2/cover/profile"
)

type ProfileInterface interface {
	Get(ids []string, opts ...profileOption) (string, error)
	Delete(ids []string, opts ...profileOption) error
}

type profileClient struct {
	c            *resty.Client
	skipPatterns []string
	extraPattern string
}

func NewProfileClient(c *resty.Client) *profileClient {
	return &profileClient{
		c: c,
	}
}

type profileOption func(*profileClient)

func WithPackagePattern(skips []string) profileOption {
	return func(pc *profileClient) {
		pc.skipPatterns = skips
	}
}

func WithExtraPattern(pattern string) profileOption {
	return func(pc *profileClient) {
		pc.extraPattern = pattern
	}
}

func (p *profileClient) Get(ids []string, opts ...profileOption) (string, error) {
	for _, opt := range opts {
		opt(p)
	}

	req := p.c.R()

	idQuery := strings.Join(ids, ",")
	skipQuery := strings.Join(p.skipPatterns, ",")

	req.QueryParam.Add("id", idQuery)
	req.QueryParam.Add("skippattern", skipQuery)
	req.QueryParam.Add("extra", p.extraPattern)

	res := struct {
		Data string `json:"profile,omitempty"`
		Msg  string `jaon:"msg,omitempty"`
	}{}

	resp, err := req.
		Get(profileAPI)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return res.Msg, fmt.Errorf("status code not 200")
	}

	return res.Data, nil
}

func (p *profileClient) Delete(ids []string, opts ...profileOption) error {
	for _, opt := range opts {
		opt(p)
	}

	req := p.c.R()

	idQuery := strings.Join(ids, ",")
	skipQuery := strings.Join(p.skipPatterns, ",")

	req.QueryParam.Add("id", idQuery)
	req.QueryParam.Add("skippattern", skipQuery)
	req.QueryParam.Add("extra", p.extraPattern)

	_, err := req.
		Delete(profileAPI)
	if err != nil {
		return err
	}

	return nil
}
