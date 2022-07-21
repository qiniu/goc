package cover

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/tools/cover"
)

// MockStore is mock store mainly for unittest
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Add(s ServiceUnderTest) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockStore) Remove(a string) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockStore) Get(name string) []string {
	args := m.Called(name)
	return args.Get(0).([]string)
}

func (m *MockStore) GetAll() map[string][]string {
	args := m.Called()
	return args.Get(0).(map[string][]string)
}

func (m *MockStore) Init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStore) Set(services map[string][]string) error {
	args := m.Called()
	return args.Error(0)
}

func TestContains(t *testing.T) {
	assert.Equal(t, contains([]string{"a", "b"}, "a"), true)
	assert.Equal(t, contains([]string{"a", "b"}, "c"), false)
}

func TestFilterAddrs(t *testing.T) {
	svrAll := map[string][]string{
		"service1": {"http://127.0.0.1:7777", "http://127.0.0.1:8888"},
		"service2": {"http://127.0.0.1:9999"},
	}
	addrAll := []string{}
	for _, addr := range svrAll {
		addrAll = append(addrAll, addr...)
	}
	items := []struct {
		svrList  []string
		addrList []string
		force    bool
		err      string
		addrRes  []string
	}{
		{
			svrList:  []string{"service1"},
			addrList: []string{"http://127.0.0.1:7777"},
			err:      "use 'service' flag and 'address' flag at the same time may cause ambiguity, please use them separately",
		},
		{
			addrRes: addrAll,
		},
		{
			svrList: []string{"service1", "unknown"},
			err:     "service [unknown] not found",
		},
		{
			svrList: []string{"service1", "service2", "unknown"},
			force:   true,
			addrRes: addrAll,
		},
		{
			svrList: []string{"unknown"},
			force:   true,
		},
		{
			addrList: []string{"http://127.0.0.1:7777", "http://127.0.0.2:7777"},
			err:      "address [http://127.0.0.2:7777] not found",
		},
		{
			addrList: []string{"http://127.0.0.1:7777", "http://127.0.0.1:9999", "http://127.0.0.2:7777"},
			force:    true,
			addrRes:  []string{"http://127.0.0.1:7777", "http://127.0.0.1:9999"},
		},
	}
	for _, item := range items {
		res, err := filterAddrInfo(item.svrList, item.addrList, item.force, svrAll)
		if err != nil {
			assert.Equal(t, err.Error(), item.err)
		} else {
			addrs := []string{}
			if len(res) == 0 {
				addrs = nil
			} else {
				for _, addr := range res {
					addrs = append(addrs, addr.Address)
				}
			}
			if len(addrs) == 0 {
				assert.Equal(t, addrs, item.addrRes)
			}
			for _, a := range addrs {
				assert.Contains(t, item.addrRes, a)
			}
		}
	}
}

func TestRegisterService(t *testing.T) {
	server, err := NewFileBasedServer("_svrs_address.txt")
	assert.NoError(t, err)
	router := server.Route(os.Stdout)

	// register with empty service struct
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/cover/register", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// register with invalid service.Address(bad uri)
	data := url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "&%%")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid URL escape")

	// regist with invalid service.Address(schema)
	data = url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "fpt://127.0.0.1:21")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unsupport schema")

	// regist with unempty path
	data = url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "http://127.0.0.1:21/") // valid scenario, the final stored address will be http://127.0.0.0.1:21
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// regist with empty host(in direct mode,empty must fail,default is success)
	// in default,use request realhost as host,use 80 as port
	server.IPRevise = false
	data = url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "http://")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "empty host")

	data = url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "http://:8080") //valid scenario, be consistent with curl
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	data = url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "http://127.0.0.1")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	server.IPRevise = true //use clientip and default port(80)
	data = url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "http://")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	data = url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "http://:8080")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	data = url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "http://127.0.0.1")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// register with store failure
	expectedS := ServiceUnderTest{
		Name:     "foo",
		Address:  "http://:8080",
		IPRevise: "true",
	}
	testObj := new(MockStore)
	testObj.On("Get", "foo").Return([]string{"http://127.0.0.1:66666"})
	testObj.On("Add", expectedS).Return(fmt.Errorf("lala error"))

	server.Store = testObj

	w = httptest.NewRecorder()
	data.Set("name", expectedS.Name)
	data.Set("address", expectedS.Address)
	data.Set("ip_revise", expectedS.IPRevise)
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "lala error")
}

func TestProfileService(t *testing.T) {
	server, err := NewFileBasedServer("_svrs_address.txt")
	assert.NoError(t, err)
	router := server.Route(os.Stdout)

	// get profile with invalid force parameter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/cover/profile?force=11", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Contains(t, w.Body.String(), "invalid syntax")
}

func TestClearService(t *testing.T) {
	testObj := new(MockStore)
	testObj.On("GetAll").Return(map[string][]string{"foo": {"http://127.0.0.1:66666"}})

	server := &server{
		Store: testObj,
	}
	router := server.Route(os.Stdout)

	// clear profile with non-exist port
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/cover/clear", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Contains(t, w.Body.String(), "invalid port")

	// clear profile with invalid service
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/clear", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")

	// clear profile with service and address set at at the same time
	p := ProfileParam{
		Service: []string{"goc"},
		Address: []string{"http://127.0.0.1:3333"},
	}
	encoded, err := json.Marshal(p)
	assert.NoError(t, err)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/clear", bytes.NewBuffer(encoded))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Contains(t, w.Body.String(), "use 'service' flag and 'address' flag at the same time may cause ambiguity, please use them separately")
}

func TestRemoveServices(t *testing.T) {
	testObj := new(MockStore)
	testObj.On("GetAll").Return(map[string][]string{"foo": {"test1", "test2"}})
	testObj.On("Remove", "test1").Return(nil)

	server := &server{
		Store: testObj,
	}
	router := server.Route(os.Stdout)

	// remove with invalid request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/cover/remove", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")

	// remove service
	p := ProfileParam{
		Address: []string{"test1"},
	}
	encoded, err := json.Marshal(p)
	assert.NoError(t, err)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/remove", bytes.NewBuffer(encoded))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Register service test1 removed from the center.")

	// remove service with non-exist address
	testObj.On("Remove", "test2").Return(fmt.Errorf("no service found"))
	p = ProfileParam{
		Address: []string{"test2"},
	}
	encoded, err = json.Marshal(p)
	assert.NoError(t, err)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/remove", bytes.NewBuffer(encoded))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Contains(t, w.Body.String(), "no service found")

	// clear profile with service and address set at at the same time
	p = ProfileParam{
		Service: []string{"goc"},
		Address: []string{"http://127.0.0.1:3333"},
	}
	encoded, err = json.Marshal(p)
	assert.NoError(t, err)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/remove", bytes.NewBuffer(encoded))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Contains(t, w.Body.String(), "use 'service' flag and 'address' flag at the same time may cause ambiguity, please use them separately")
}

func TestInitService(t *testing.T) {
	testObj := new(MockStore)
	testObj.On("Init").Return(fmt.Errorf("lala error"))

	server := &server{
		Store: testObj,
	}
	router := server.Route(os.Stdout)

	// get profile with invalid force parameter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/cover/init", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "lala error")
}

func TestFilterProfile(t *testing.T) {
	var tcs = []struct {
		name      string
		pattern   []string
		input     []*cover.Profile
		output    []*cover.Profile
		expectErr bool
	}{
		{
			name:    "normal path",
			pattern: []string{"some/fancy/gopath", "a/fancy/gopath"},
			input: []*cover.Profile{
				{
					FileName: "some/fancy/gopath/a.go",
				},
				{
					FileName: "some/fancy/gopath/b/a.go",
				},
				{
					FileName: "a/fancy/gopath/a.go",
				},
				{
					FileName: "b/fancy/gopath/a.go",
				},
				{
					FileName: "b/a/fancy/gopath/a.go",
				},
			},
			output: []*cover.Profile{
				{
					FileName: "some/fancy/gopath/a.go",
				},
				{
					FileName: "some/fancy/gopath/b/a.go",
				},
				{
					FileName: "a/fancy/gopath/a.go",
				},
				{
					FileName: "b/a/fancy/gopath/a.go",
				},
			},
		},
		{
			name:    "with regular expression",
			pattern: []string{"fancy/gopath/a.go$", "^b/a/"},
			input: []*cover.Profile{
				{
					FileName: "some/fancy/gopath/a.go",
				},
				{
					FileName: "some/fancy/gopath/b/a.go",
				},
				{
					FileName: "a/fancy/gopath/a.go",
				},
				{
					FileName: "b/fancy/gopath/c/a.go",
				},
				{
					FileName: "b/a/fancy/gopath/a.go",
				},
			},
			output: []*cover.Profile{
				{
					FileName: "some/fancy/gopath/a.go",
				},
				{
					FileName: "a/fancy/gopath/a.go",
				},
				{
					FileName: "b/a/fancy/gopath/a.go",
				},
			},
		},
		{
			name:    "with invalid regular expression",
			pattern: []string{"(?!a)"},
			input: []*cover.Profile{
				{
					FileName: "some/fancy/gopath/a.go",
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out, err := filterProfile(tc.pattern, tc.input)
			if err != nil {
				if !tc.expectErr {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			if tc.expectErr {
				t.Errorf("Expected an error, but got value %s", stringifyCoverProfile(out))
			}

			if !reflect.DeepEqual(out, tc.output) {
				t.Errorf("Mismatched results. \nExpected: %s\nActual:%s", stringifyCoverProfile(tc.output), stringifyCoverProfile(out))
			}
		})
	}
}

func TestSkipProfile(t *testing.T) {
	var tcs = []struct {
		name      string
		pattern   []string
		input     []*cover.Profile
		output    []*cover.Profile
		expectErr bool
	}{
		{
			name:    "normal path",
			pattern: []string{"some/fancy/gopath", "a/fancy/gopath"},
			input: []*cover.Profile{
				{
					FileName: "some/fancy/gopath/a.go",
				},
				{
					FileName: "some/fancy/gopath/b/a.go",
				},
				{
					FileName: "a/fancy/gopath/a.go",
				},
				{
					FileName: "b/fancy/gopath/a.go",
				},
				{
					FileName: "b/a/fancy/gopath/a.go",
				},
			},
			output: []*cover.Profile{
				{
					FileName: "b/fancy/gopath/a.go",
				},
			},
		},
		{
			name:    "with regular expression",
			pattern: []string{"fancy/gopath/a.go$", "^b/a/"},
			input: []*cover.Profile{
				{
					FileName: "some/fancy/gopath/a.go",
				},
				{
					FileName: "some/fancy/gopath/b/a.go",
				},
				{
					FileName: "a/fancy/gopath/a.go",
				},
				{
					FileName: "b/fancy/gopath/c/a.go",
				},
				{
					FileName: "b/a/fancy/gopath/a.go",
				},
			},
			output: []*cover.Profile{
				{
					FileName: "some/fancy/gopath/b/a.go",
				},
				{
					FileName: "b/fancy/gopath/c/a.go",
				},
			},
		},
		{
			name:    "with invalid regular expression",
			pattern: []string{"(?!a)"},
			input: []*cover.Profile{
				{
					FileName: "some/fancy/gopath/a.go",
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out, err := skipProfile(tc.pattern, tc.input)
			if err != nil {
				if !tc.expectErr {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			if tc.expectErr {
				t.Errorf("Expected an error, but got value %s", stringifyCoverProfile(out))
			}

			if !reflect.DeepEqual(out, tc.output) {
				t.Errorf("Mismatched results. \nExpected: %s\nActual:%s", stringifyCoverProfile(tc.output), stringifyCoverProfile(out))
			}
		})
	}
}

func stringifyCoverProfile(profiles []*cover.Profile) string {
	res := make([]cover.Profile, 0, len(profiles))
	for _, p := range profiles {
		res = append(res, *p)
	}

	return fmt.Sprintf("%#v", res)
}
