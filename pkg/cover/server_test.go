package cover

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func TestRemoveDuplicateElement(t *testing.T) {
	strArr := []string{"a", "a", "b"}
	assert.Equal(t, removeDuplicateElement(strArr), []string{"a", "b"})
}
