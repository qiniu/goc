package client

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func captureStdout(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		logrus.WithError(err).Fatal("os pipe fail")
	}
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
	mockAgents := `{"items": [{"id": "testID", "remoteip": "1.1.1.1", "hostname": "testHost", "cmdline": "testCmd", "pid": "0"}]}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockAgents))
	}))
	defer mockServer.Close()

	c := NewWorker(mockServer.URL)
	output := captureStdout(c.ListAgents)
	expected := `+--------+----------+----------+-----+---------+
|   ID   | RemoteIP | Hostname | Pid |   CMD   |
+--------+----------+----------+-----+---------+
| testID | 1.1.1.1  | testHost |  0  | testCmd |
+--------+----------+----------+-----+---------+
`
	assert.Equal(t, output, expected)
}

func TestClientDo(t *testing.T) {
	c := &client{
		client: http.DefaultClient,
	}
	_, _, err := c.do(" ", "http://127.0.0.1:7777", "", nil) // a invalid method
	assert.Contains(t, err.Error(), "invalid method")
}
