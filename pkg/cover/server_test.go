package cover

import (
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

func (m *MockStore) Add(s Service) error {
	args := m.Called(s)
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

func (m *MockStore) Set(services map[string][]string) {
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
		addrs, err := filterAddrs(item.svrList, item.addrList, item.force, svrAll)
		if err != nil {
			assert.Equal(t, err.Error(), item.err)
		} else {
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
	router := GocServer(os.Stdout)

	// register with empty service struct
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/cover/register", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// register with invalid service.Address
	data := url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "&%%")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid URL escape")

	// register with host but no port
	data = url.Values{}
	data.Set("name", "aaa")
	data.Set("address", "http://127.0.0.1")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "missing port in address")

	// register with store failure
	expectedS := Service{
		Name:    "foo",
		Address: "http://:64444", // the real IP is empty in unittest, so server will get a empty one
	}
	testObj := new(MockStore)
	testObj.On("Get", "foo").Return([]string{"http://127.0.0.1:66666"})
	testObj.On("Add", expectedS).Return(fmt.Errorf("lala error"))

	DefaultStore = testObj

	w = httptest.NewRecorder()
	data.Set("name", expectedS.Name)
	data.Set("address", expectedS.Address)
	req, _ = http.NewRequest("POST", "/v1/cover/register", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "lala error")
}

func TestProfileService(t *testing.T) {
	router := GocServer(os.Stdout)

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

	DefaultStore = testObj

	router := GocServer(os.Stdout)

	// get profile with invalid force parameter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/cover/clear", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusExpectationFailed, w.Code)
	assert.Contains(t, w.Body.String(), "invalid port")
}

func TestInitService(t *testing.T) {
	testObj := new(MockStore)
	testObj.On("Init").Return(fmt.Errorf("lala error"))

	DefaultStore = testObj

	router := GocServer(os.Stdout)

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

func stringifyCoverProfile(profiles []*cover.Profile) string {
	res := make([]cover.Profile, 0, len(profiles))
	for _, p := range profiles {
		res = append(res, *p)
	}

	return fmt.Sprintf("%#v", res)
}
