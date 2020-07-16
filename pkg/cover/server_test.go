package cover

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContains(t *testing.T) {
	assert.Equal(t, contains([]string{"a", "b"}, "a"), true)
	assert.Equal(t, contains([]string{"a", "b"}, "c"), false)
}

func TestGetSvrUnderTest(t *testing.T) {
	svrAll := map[string][]string{
		"service1": {"http://127.0.0.1:7777", "http://127.0.0.1:8888"},
		"service2": {"http://127.0.0.1:9999"},
	}
	items := []struct {
		svrList  []string
		addrList []string
		force    bool
		err      string
		svrRes   map[string][]string
	}{
		{
			svrList:  []string{"service1"},
			addrList: []string{"http://127.0.0.1:7777"},
			err:      "use this flag and 'address' flag at the same time is illegal",
		},
		{
			svrRes: svrAll,
		},
		{
			svrList: []string{"service1", "unknown"},
			err:     "service [unknown] not found",
		},
		{
			svrList: []string{"service1", "service1", "service2", "unknown"},
			force:   true,
			svrRes:  svrAll,
		},
		{
			addrList: []string{"http://127.0.0.1:7777", "http://127.0.0.2:7777"},
			err:      "address [http://127.0.0.2:7777] not found",
		},
		{
			addrList: []string{"http://127.0.0.1:7777", "http://127.0.0.1:7777", "http://127.0.0.1:9999", "http://127.0.0.2:7777"},
			force:    true,
			svrRes:   map[string][]string{"service1": {"http://127.0.0.1:7777"}, "service2": {"http://127.0.0.1:9999"}},
		},
	}
	for _, item := range items {
		svrs, err := getSvrUnderTest(item.svrList, item.addrList, item.force, svrAll)
		if err != nil {
			assert.Equal(t, err.Error(), item.err)
		} else {
			assert.Equal(t, svrs, item.svrRes)
		}
	}
}
