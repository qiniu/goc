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
	"net/http/httptest"
	"os"
	"testing"

	"net/http"

	"github.com/stretchr/testify/assert"
)

func TestClientAction(t *testing.T) {
	// mock goc server
	ts := httptest.NewServer(GocServer(os.Stdout))
	defer ts.Close()
	var client = NewWorker(ts.URL)

	// mock profile server
	profileMockResponse := "mode: count\nmockService/main.go:30.13,48.33 13 1"
	profileSuccessMockSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(profileMockResponse))
	}))
	defer profileSuccessMockSvr.Close()

	profileErrMockSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("error"))
	}))
	defer profileErrMockSvr.Close()

	// regsiter service into goc server
	var src Service
	src.Name = "serviceSuccess"
	src.Address = profileSuccessMockSvr.URL
	res, err := client.RegisterService(src)
	assert.NoError(t, err)
	assert.Contains(t, string(res), "success")

	// do list and check service
	res, err = client.ListServices()
	assert.NoError(t, err)
	assert.Contains(t, string(res), src.Address)
	assert.Contains(t, string(res), src.Name)

	// get porfile from goc server
	profileItems := []struct {
		service Service
		param   ProfileParam
		res     string
	}{
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{Force: false, Service: []string{"serviceOK"}, Address: []string{profileSuccessMockSvr.URL}},
			res:     "use 'service' flag and 'address' flag at the same time may cause ambiguity, please use them separately",
		},
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{},
			res:     profileMockResponse,
		},
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{Service: []string{"serviceOK"}},
			res:     profileMockResponse,
		},
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{Address: []string{profileSuccessMockSvr.URL}},
			res:     profileMockResponse,
		},
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{Service: []string{"unknown"}},
			res:     "service [unknown] not found",
		},
		{
			service: Service{Name: "serviceErr", Address: profileErrMockSvr.URL},
			res:     "bad mode line: error",
		},
		{
			service: Service{Name: "serviceNotExist", Address: "http://172.0.0.2:7777"},
			res:     "connection refused",
		},
		{
			service: Service{Name: "serviceNotExist", Address: "http://172.0.0.2:7777"},
			param:   ProfileParam{Force: true},
			res:     "no profiles",
		},
	}
	for _, item := range profileItems {
		// init server
		_, err = client.InitSystem()
		assert.NoError(t, err)
		// register server
		res, err = client.RegisterService(item.service)
		assert.NoError(t, err)
		assert.Contains(t, string(res), "success")
		res, err = client.Profile(item.param)
		if err != nil {
			assert.Equal(t, err.Error(), item.res)
		} else {
			assert.Contains(t, string(res), item.res)
		}
	}

	// init system and check service again
	_, err = client.InitSystem()
	assert.NoError(t, err)
	res, err = client.ListServices()
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(res))
}

func TestClientRegisterService(t *testing.T) {
	c := &client{}

	// client register with empty address
	testService1 := Service{
		Address: "",
		Name:    "abc",
	}
	_, err := c.RegisterService(testService1)
	assert.Contains(t, err.Error(), "empty url")

	// client register with empty name
	testService2 := Service{
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
