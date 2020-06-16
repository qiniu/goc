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
	"testing"
)

func TestLocalStore(t *testing.T) {
	localStore := NewFileStore()
	var tc1 = Service{
		Name:    "a",
		Address: "http://127.0.0.1",
	}
	var tc2 = Service{
		Name:    "b",
		Address: "http://127.0.0.2",
	}
	var tc3 = Service{
		Name:    "c",
		Address: "http://127.0.0.3",
	}
	var tc4 = Service{
		Name:    "a",
		Address: "http://127.0.0.4",
	}
	localStore.Add(tc1)
	localStore.Add(tc2)
	localStore.Add(tc3)
	localStore.Add(tc4)
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

	localStore.Init()
	if len(localStore.GetAll()) != 0 {
		t.Error("local store init failed")
	}
}
