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
	"net/http/httptest"
	"os"
	"testing"

	"net/http"

	"github.com/stretchr/testify/assert"
)

func TestClientAction(t *testing.T) {
	// mock goc server
	server, err := NewFileBasedServer("_svrs_address.txt")
	assert.NoError(t, err)
	ts := httptest.NewServer(server.Route(os.Stdout))
	defer ts.Close()
	var client = NewWorker(ts.URL)

	// mock profile server
	profileMockResponse := []byte("mode: count\nmockService/main.go:30.13,48.33 13 1\nb/b.go:30.13,48.33 13 1")
	profileSuccessMockSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(profileMockResponse)
	}))
	defer profileSuccessMockSvr.Close()

	profileErrMockSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("error"))
	}))
	defer profileErrMockSvr.Close()

	// register service into goc server
	var src ServiceUnderTest
	src.Name = "serviceSuccess"
	src.Address = profileSuccessMockSvr.URL
	res, err := client.RegisterService(src)
	assert.NoError(t, err)
	assert.Contains(t, string(res), "success")

	// do list and check server
	res, err = client.ListServices()
	assert.NoError(t, err)
	assert.Contains(t, string(res), src.Address)
	assert.Contains(t, string(res), src.Name)

	// get profile from goc server
	tcs := []struct {
		name        string
		service     ServiceUnderTest
		param       ProfileParam
		expected    string
		expectedErr bool
	}{
		{
			name:        "both server and address existed",
			service:     ServiceUnderTest{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:       ProfileParam{Force: false, Service: []string{"serviceOK"}, Address: []string{profileSuccessMockSvr.URL}},
			expectedErr: true,
		},
		{
			name:     "valid test with no server flag provide",
			service:  ServiceUnderTest{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:    ProfileParam{},
			expected: "mockService/main.go:30.13,48.33 13 1",
		},
		{
			name:     "valid test with server flag provide",
			service:  ServiceUnderTest{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:    ProfileParam{Service: []string{"serviceOK"}},
			expected: "mockService/main.go:30.13,48.33 13 1",
		},
		{
			name:     "valid test with address flag provide",
			service:  ServiceUnderTest{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:    ProfileParam{Address: []string{profileSuccessMockSvr.URL}},
			expected: "mockService/main.go:30.13,48.33 13 1",
		},
		{
			name:        "invalid test with invalid server flag provide",
			service:     ServiceUnderTest{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:       ProfileParam{Service: []string{"unknown"}},
			expected:    "server [unknown] not found",
			expectedErr: true,
		},
		{
			name:        "invalid test with invalid profile got by server",
			service:     ServiceUnderTest{Name: "serviceErr", Address: profileErrMockSvr.URL},
			expected:    "bad mode line: error",
			expectedErr: true,
		},
		{
			name:        "invalid test with disconnected server",
			service:     ServiceUnderTest{Name: "serviceNotExist", Address: "http://172.0.0.2:7777"},
			expected:    "connection refused",
			expectedErr: true,
		},
		{
			name:        "invalid test with empty profile",
			service:     ServiceUnderTest{Name: "serviceNotExist", Address: "http://172.0.0.2:7777"},
			param:       ProfileParam{Force: true},
			expectedErr: true,
			expected:    "no profiles",
		},
		{
			name:     "valid test with coverfile flag provide",
			service:  ServiceUnderTest{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:    ProfileParam{CoverFilePatterns: []string{"b.go$"}},
			expected: "b/b.go",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// init server
			_, err = client.InitSystem()
			assert.NoError(t, err)
			// register server
			res, err = client.RegisterService(tc.service)
			assert.NoError(t, err)
			assert.Contains(t, string(res), "success")
			res, err = client.Profile(tc.param)
			if err != nil {
				if !tc.expectedErr {
					t.Errorf("unexpected err got: %v", err)
				}
				return
			}

			if tc.expectedErr {
				t.Errorf("Expected an error, but got value %s", string(res))
			}

			assert.Regexp(t, tc.expected, string(res))
		})
	}

	// init system and check server again
	_, err = client.InitSystem()
	assert.NoError(t, err)
	res, err = client.ListServices()
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(res))
}

func TestClientRegisterService(t *testing.T) {
	c := &client{}

	// client register with empty address
	testService1 := ServiceUnderTest{
		Address: "",
		Name:    "abc",
	}
	_, err := c.RegisterService(testService1)
	assert.Contains(t, err.Error(), "empty url")

	// client register with empty name
	testService2 := ServiceUnderTest{
		Address: "http://127.0.0.1:444",
		Name:    "",
	}
	_, err = c.RegisterService(testService2)
	assert.EqualError(t, err, "invalid service name")
}

func TestClientListServices(t *testing.T) {
	c := &client{
		Host:   "http://127.0.0.1:64445", // a invalid host
		client: http.DefaultClient,
	}
	_, err := c.ListServices()
	assert.Contains(t, err.Error(), "connect: connection refused")
}

func TestClientDo(t *testing.T) {
	c := &client{
		client: http.DefaultClient,
	}
	_, _, err := c.do(" ", "http://127.0.0.1:7777", "", nil) // a invalid method
	assert.Contains(t, err.Error(), "invalid method")
}

func TestClientClearWithInvalidParam(t *testing.T) {
	p := ProfileParam{
		Service: []string{"goc"},
		Address: []string{"http://127.0.0.1:777"},
	}
	c := &client{
		client: http.DefaultClient,
	}
	_, err := c.Clear(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "use 'service' flag and 'address' flag at the same time may cause ambiguity, please use them separately")
}

func TestClientRemove(t *testing.T) {
	// remove by invalid param
	p := ProfileParam{
		Service: []string{"goc"},
		Address: []string{"http://127.0.0.1:777"},
	}
	c := &client{
		client: http.DefaultClient,
	}
	_, err := c.Remove(p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "use 'service' flag and 'address' flag at the same time may cause ambiguity, please use them separately")

	// remove by valid param
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	c.Host = ts.URL
	p = ProfileParam{
		Address: []string{"http://127.0.0.1:777"},
	}
	res, err := c.Remove(p)
	assert.NoError(t, err)
	assert.Equal(t, string(res), "Hello, client\n")

	// remove from a invalid center
	c.Host = "http://127.0.0.1:11111"
	_, err = c.Remove(p)
	assert.Error(t, err)
}
