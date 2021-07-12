package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
			fmt.Println(output)
			assert.Equal(t, output, tt.expected)
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
