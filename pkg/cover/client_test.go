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

	"github.com/stretchr/testify/assert"
)

func TestClientAction(t *testing.T) {
	ts := httptest.NewServer(GocServer(os.Stdout))
	defer ts.Close()
	var client = NewWorker(ts.URL)

	// regsiter service into goc server
	var src Service
	src.Name = "goc"
	src.Address = "http://127.0.0.1:7777"
	res, err := client.RegisterService(src)
	assert.NoError(t, err)
	assert.Contains(t, string(res), "success")

	// do list and check service
	res, err = client.ListServices()
	assert.NoError(t, err)
	assert.Contains(t, string(res), src.Address)
	assert.Contains(t, string(res), src.Name)

	// init system and check service again
	res, err = client.InitSystem()
	assert.NoError(t, err)
	res, err = client.ListServices()
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(res))
}

func TestE2E(t *testing.T) {
	// FIXME: start goc server
	// FIXME: call goc build to cover goc server
	// FIXME: do some tests again goc server
	// FIXME: goc profile and checkout coverage
}
