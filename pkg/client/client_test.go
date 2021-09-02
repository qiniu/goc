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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = stdout
	}()

	f()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

func TestClientListAgents(t *testing.T) {
	mockAgents := `{"items": [{"id": "testID", "remoteip": "1.1.1.1", "hostname": "testHost", "cmdline": "./testCmd -f testArgs", "pid": "0"}]}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockAgents))
	}))
	defer mockServer.Close()

	c := NewWorker(mockServer.URL)
	testCases := map[string]struct {
		input    bool
		expected string
	}{
		"simple list": {
			false,
			`ID       REMOTEIP   CMD              
testID   1.1.1.1    ./testCmd -f tes   
`,
		},
		"wide list": {
			true,
			`ID       REMOTEIP   HOSTNAME   PID   CMD                   
testID   1.1.1.1    testHost   0     ./testCmd -f testArgs   
`,
		},
	}
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			f := func() { c.ListAgents(tt.input) }
			output := captureStdout(f)
			assert.Equal(t, output, tt.expected)
		})
	}
}

func TestClientProfile(t *testing.T) {
	mockAgents := `{"profile": "mode: count\nmockService/main.go:30.13,48.33 13 1\nb/b.go:30.13,48.33 13 1"}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockAgents))
	}))
	defer mockServer.Close()

	c := NewWorker(mockServer.URL)
	f := func() { c.Profile("") }
	output := captureStdout(f)
	assert.Regexp(t, "mockService/main.go:30.13,48.33 13 1", output)
}

func TestClientProfile_Output(t *testing.T) {
	mockAgents := `{"profile": "mode: count\nmockService/main.go:30.13,48.33 13 1\nb/b.go:30.13,48.33 13 1"}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockAgents))
	}))
	defer mockServer.Close()

	c := NewWorker(mockServer.URL)
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)
	testCases := map[string]struct {
		input  string
		output string
	}{
		"file":           {"test.cov", "test.cov"},
		"file with path": {filepath.Join(exPath, "test.txt"), filepath.Join(exPath, "test.txt")},
		"just path":      {fmt.Sprintf("%s%c", exPath, os.PathSeparator), filepath.Join(exPath, "coverage.cov")},
	}
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			c.Profile(tt.input)
			defer os.RemoveAll(tt.output)
			assert.FileExists(t, tt.output)
		})
	}
}

func TestClientDo(t *testing.T) {
	c := &client{
		client: http.DefaultClient,
	}
	_, _, err := c.do(" ", "http://127.0.0.1:7777", "", nil) // a invalid method
	assert.Contains(t, err.Error(), "invalid method")
}

func TestGetSimpleSvcName(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected string
	}{
		"short path": {"1234567890abc.go", "1234567890abc.go"},
		"long path":  {"1234567890abcdef.go", "1234567890abcdef"},
	}
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, getSimpleCmdLine(0, tt.input), tt.expected)
		})
	}
}
