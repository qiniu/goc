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

package rest

import (
	"github.com/go-resty/resty/v2"
	"github.com/qiniu/goc/v2/pkg/client/rest/agent"
	"github.com/qiniu/goc/v2/pkg/client/rest/profile"
)

// V2Client provides methods contact with the covered agent under test
type V2Client struct {
	rest *resty.Client
}

func NewV2Client(host string) *V2Client {
	return &V2Client{
		rest: resty.New().SetHostURL("http://" + host),
	}
}

func (c *V2Client) Agent() agent.AgentInterface {
	return agent.NewAgentsClient(c.rest)
}

func (c *V2Client) Profile() profile.ProfileInterface {
	return profile.NewProfileClient(c.rest)
}
