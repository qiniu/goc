package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/qiniu/goc/v2/pkg/log"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
)

// Action provides methods to contact with the covered agent under test
type Action interface {
	ListAgents()
}

const (
	// CoverAgentsListAPI list all the registered agents
	CoverAgentsListAPI = "/v2/rpcagents"
)

var (
	ERROR_GOC_LIST = errors.New("goc list failed")
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

// NewWorker creates a worker to contact with host
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

func (c *client) ListAgents() {
	u := fmt.Sprintf("%s%s", c.Host, CoverAgentsListAPI)
	_, body, err := c.do("GET", u, "", nil)
	if err != nil && isNetworkError(err) {
		_, body, err = c.do("GET", u, "", nil)
	}
	if err != nil {
		err = fmt.Errorf("%w: %v", ERROR_GOC_LIST, err)
		log.Fatalf(err.Error())
	}
	agents := gocListAgents{}
	err = json.Unmarshal(body, &agents)
	if err != nil {
		err = fmt.Errorf("%w: json unmarshal failed: %v", ERROR_GOC_LIST, err)
		log.Fatalf(err.Error())
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "RemoteIP", "Hostname", "Pid", "CMD"})
	table.SetAutoFormatHeaders(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})
	for _, agent := range agents.Items {
		table.Append([]string{agent.Id, agent.RemoteIP, agent.Hostname, agent.Pid, agent.CmdLine})
	}
	table.Render()
	return
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
