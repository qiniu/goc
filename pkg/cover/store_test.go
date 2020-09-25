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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalStore(t *testing.T) {
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
