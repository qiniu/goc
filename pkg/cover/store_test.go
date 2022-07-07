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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalStore(t *testing.T) {
	os.Remove("_svrs_address.txt") //remove _svrs_address.txt file，make sure no data effect this unit test
	localStore, err := NewFileStore("_svrs_address.txt")
	assert.NoError(t, err)
	var tc1 = ServiceUnderTest{
		Name:    "a",
		Address: "http://127.0.0.1",
	}
	var tc2 = ServiceUnderTest{
		Name:    "b",
		Address: "http://127.0.0.2",
	}
	var tc3 = ServiceUnderTest{
		Name:    "c",
		Address: "http://127.0.0.3",
	}
	var tc4 = ServiceUnderTest{
		Name:    "a",
		Address: "http://127.0.0.4",
	}
	assert.NoError(t, localStore.Add(tc1))
	assert.Equal(t, localStore.Add(tc1), ErrServiceAlreadyRegistered)
	assert.NoError(t, localStore.Add(tc2))
	assert.NoError(t, localStore.Add(tc3))
	assert.NoError(t, localStore.Add(tc4))
	addrs := localStore.Get(tc1.Name)
	if len(addrs) != 2 {
		t.Error("unexpected result")
	}

	for _, addr := range addrs {
		if addr != tc1.Address && addr != tc4.Address {
			t.Error("get address failed")
		}
	}

	if len(localStore.GetAll()) != 3 {
		t.Error("local store check failed")
	}

	localStoreNew, err := NewFileStore("_svrs_address.txt")
	assert.NoError(t, err)
	assert.Equal(t, localStore.GetAll(), localStoreNew.GetAll())

	localStore.Init()
	if len(localStore.GetAll()) != 0 {
		t.Error("local store init failed")
	}
}

func TestMemoryStoreRemove(t *testing.T) {
	store := NewMemoryStore()
	s1 := ServiceUnderTest{Name: "test", Address: "http://127.0.0.1:8900"}
	s2 := ServiceUnderTest{Name: "test2", Address: "http://127.0.0.1:8901"}
	s3 := ServiceUnderTest{Name: "test2", Address: "http://127.0.0.1:8902"}

	_ = store.Add(s1)
	_ = store.Add(s2)
	_ = store.Add(s3)

	ss1 := store.Get("test")
	assert.Equal(t, 1, len(ss1))
	err := store.Remove("http://127.0.0.1:8900")
	assert.NoError(t, err)
	ss1 = store.Get("test")
	assert.Nil(t, ss1)

	ss2 := store.Get("test2")
	assert.Equal(t, 2, len(ss2))
	err = store.Remove("http://127.0.0.1:8901")
	assert.NoError(t, err)
	ss2 = store.Get("test2")
	assert.Equal(t, 1, len(ss2))

	err = store.Remove("http")
	assert.Error(t, err, fmt.Errorf("no service found"))
}

func TestFileStoreRemove(t *testing.T) {
	store, _ := NewFileStore("_svrs_address.txt")
	_ = store.Init()
	s1 := ServiceUnderTest{Name: "test", Address: "http://127.0.0.1:8900"}
	s2 := ServiceUnderTest{Name: "test2", Address: "http://127.0.0.1:8901"}
	s3 := ServiceUnderTest{Name: "test2", Address: "http://127.0.0.1:8902"}

	_ = store.Add(s1)
	_ = store.Add(s2)
	_ = store.Add(s3)

	ss1 := store.Get("test")
	assert.Equal(t, 1, len(ss1))
	err := store.Remove("http://127.0.0.1:8900")
	assert.NoError(t, err)
	ss1 = store.Get("test")
	assert.Nil(t, ss1)

	ss2 := store.Get("test2")
	assert.Equal(t, 2, len(ss2))
	err = store.Remove("http://127.0.0.1:8901")
	assert.NoError(t, err)
	ss2 = store.Get("test2")
	assert.Equal(t, 1, len(ss2))

	err = store.Remove("http")
	assert.Error(t, err, fmt.Errorf("no service found"))
}
