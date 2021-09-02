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

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/term"

	"github.com/olekukonko/tablewriter"
	"github.com/qiniu/goc/v2/pkg/log"
)

// Action provides methods to contact with the covered agent under test
type Action interface {
	ListAgents(bool)
	Profile(string)
}

const (
	// CoverAgentsListAPI list all the registered agents
	CoverAgentsListAPI = "/v2/rpcagents"
	//CoverProfileAPI is provided by the covered service to get profiles
	CoverProfileAPI = "/v2/cover/profile"
)

type client struct {
	Host   string
	client *http.Client
}

// gocListAgents response of the list request
type gocListAgents struct {
	Items []gocCoveredAgent `json:"items"`
}

// gocCoveredAgent represents a covered client
type gocCoveredAgent struct {
	Id       string `json:"id"`
	RemoteIP string `json:"remoteip"`
	Hostname string `json:"hostname"`
	CmdLine  string `json:"cmdline"`
	Pid      string `json:"pid"`
}

type gocProfile struct {
	Profile string `json:"profile"`
}

// NewWorker creates a worker to contact with host
func NewWorker(host string) Action {
	_, err := url.ParseRequestURI(host)
	if err != nil {
		log.Fatalf("parse url %s failed, err: %v", host, err)
	}
	return &client{
		Host:   host,
		client: http.DefaultClient,
	}
}

func (c *client) ListAgents(wide bool) {
	u := fmt.Sprintf("%s%s", c.Host, CoverAgentsListAPI)
	_, body, err := c.do("GET", u, "", nil)
	if err != nil && isNetworkError(err) {
		_, body, err = c.do("GET", u, "", nil)
	}
	if err != nil {
		log.Fatalf("goc list failed: %v", err)
	}
	agents := gocListAgents{}
	err = json.Unmarshal(body, &agents)
	if err != nil {
		log.Fatalf("goc list failed: json unmarshal failed: %v", err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("   ") // pad with 3 blank spaces
	table.SetNoWhiteSpace(true)
	table.SetReflowDuringAutoWrap(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	if wide {
		table.SetHeader([]string{"ID", "REMOTEIP", "HOSTNAME", "PID", "CMD"})
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	} else {
		table.SetHeader([]string{"ID", "REMOTEIP", "CMD"})
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	}
	for _, agent := range agents.Items {
		if wide {
			table.Append([]string{agent.Id, agent.RemoteIP, agent.Hostname, agent.Pid, agent.CmdLine})
		} else {
			preLen := len(agent.Id) + len(agent.RemoteIP) + 9
			table.Append([]string{agent.Id, agent.RemoteIP, getSimpleCmdLine(preLen, agent.CmdLine)})
		}
	}
	table.Render()
	return
}

func (c *client) Profile(output string) {
	u := fmt.Sprintf("%s%s", c.Host, CoverProfileAPI)

	res, profile, err := c.do("GET", u, "application/json", nil)
	if err != nil && isNetworkError(err) {
		res, profile, err = c.do("GET", u, "application/json", nil)
	}

	if err == nil && res.StatusCode != 200 {
		log.Fatalf(string(profile))
	}
	var profileText gocProfile
	err = json.Unmarshal(profile, &profileText)
	if err != nil {
		log.Fatalf("profile unmarshal failed: %v", err)
	}
	if output == "" {
		fmt.Fprint(os.Stdout, profileText.Profile)
	} else {
		var dir, filename string = filepath.Split(output)
		if dir != "" {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				log.Fatalf("failed to create directory %s, err:%v", dir, err)
			}
		}
		if filename == "" {
			output += "coverage.cov"
		}

		f, err := os.Create(output)
		if err != nil {
			log.Fatalf("failed to create file %s, err:%v", output, err)
		}
		defer f.Close()
		_, err = io.Copy(f, bytes.NewReader([]byte(profileText.Profile)))
		if err != nil {
			log.Fatalf("failed to write file: %v, err: %v", output, err)
		}
	}
}

// getSimpleCmdLine
func getSimpleCmdLine(preLen int, cmdLine string) string {
	pathLen := len(cmdLine)
	width, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil || width <= preLen+16 {
		width = 16 + preLen // show at least 16 words of the command
	}
	if pathLen > width-preLen {
		return cmdLine[:width-preLen]
	}
	return cmdLine
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
