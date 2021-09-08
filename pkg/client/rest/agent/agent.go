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

package agent

import (
	"encoding/json"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Agent struct {
	Id       string `json:"id"`
	RemoteIP string `json:"rpc_remoteip"`
	Hostname string `json:"hostname"`
	CmdLine  string `json:"cmdline"`
	Pid      string `json:"pid"`
	Status   int    `json:"status"`
	Extra    string `json:"extra"`
}

const (
	agentsAPI = "/v2/agents"
)

type AgentInterface interface {
	Get(ids []string) ([]Agent, error)
	Delete(ids []string) error
}

type agentsClient struct {
	c *resty.Client
}

func NewAgentsClient(c *resty.Client) *agentsClient {
	return &agentsClient{
		c: c,
	}
}

type agentOption func(*agentsClient)

func (a *agentsClient) Get(ids []string) ([]Agent, error) {

	req := a.c.R()

	idQuery := strings.Join(ids, ",")

	req.QueryParam.Add("id", idQuery)

	res := struct {
		Items []Agent `json:"items"`
	}{}

	resp, err := req.
		Get(agentsAPI)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		return nil, err
	}

	return res.Items, nil
}

func (a *agentsClient) Delete(ids []string) error {

	req := a.c.R()

	idQuery := strings.Join(ids, ",")

	req.QueryParam.Add("id", idQuery)

	_, err := req.
		Delete(agentsAPI)
	if err != nil {
		return err
	}

	return nil
}
