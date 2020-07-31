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

package qiniu

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/client"
	"github.com/stretchr/testify/assert"
)

// test NewRangeReader logic
func TestNewRangeReader(t *testing.T) {
	cfg := &Config{
		Bucket:    "artifacts",
		AccessKey: "ak",
		SecretKey: "sk",
	}
	_, router, serverUrl, teardown := MockQiniuServer(cfg)
	defer teardown()
	cfg.Domain = serverUrl

	MockPrivateDomainUrl(router, 0)

	oh := &QnObjectHandle{
		key:    "key",
		cfg:    cfg,
		bm:     nil,
		mac:    qbox.NewMac(cfg.AccessKey, cfg.SecretKey),
		client: &client.Client{Client: http.DefaultClient},
	}

	// test read unlimited
	body, err := oh.NewRangeReader(context.Background(), 0, -1)
	assert.Equal(t, err, nil)

	bodyBytes, err := ioutil.ReadAll(body)
	assert.NoError(t, err)
	assert.Equal(t, string(bodyBytes), "mock server ok")

	// test with HEAD method
	body, err = oh.NewRangeReader(context.Background(), 0, 0)
	assert.Equal(t, err, nil)

	bodyBytes, err = ioutil.ReadAll(body)
	assert.NoError(t, err)
	assert.Equal(t, string(bodyBytes), "")

}

// test retry logic
func TestNewRangeReaderWithTimeoutAndRecover(t *testing.T) {
	cfg := &Config{
		Bucket:    "artifacts",
		AccessKey: "ak",
		SecretKey: "sk",
	}
	_, router, serverUrl, teardown := MockQiniuServer(cfg)
	defer teardown()
	cfg.Domain = serverUrl

	MockPrivateDomainUrl(router, 2)

	oh := &QnObjectHandle{
		key:    "key",
		cfg:    cfg,
		bm:     nil,
		mac:    qbox.NewMac(cfg.AccessKey, cfg.SecretKey),
		client: &client.Client{Client: http.DefaultClient},
	}

	// test with timeout
	oh.key = "timeout"
	body, err := oh.NewRangeReader(context.Background(), 0, 10)
	assert.Equal(t, err, nil)

	bodyBytes, err := ioutil.ReadAll(body)
	assert.NoError(t, err)
	assert.Equal(t, string(bodyBytes), "mock server ok")

	// test with retry with statuscode=571, 573
	oh.key = "retry"
	body, err = oh.NewRangeReader(context.Background(), 0, 10)
	assert.Equal(t, err, nil)

	bodyBytes, err = ioutil.ReadAll(body)
	assert.NoError(t, err)
	assert.Equal(t, string(bodyBytes), "mock server ok")
}

// test retry logic
func TestNewRangeReaderWithTimeoutNoRecover(t *testing.T) {
	cfg := &Config{
		Bucket:    "artifacts",
		AccessKey: "ak",
		SecretKey: "sk",
	}
	_, router, serverUrl, teardown := MockQiniuServer(cfg)
	defer teardown()
	cfg.Domain = serverUrl

	MockPrivateDomainUrl(router, 12)

	oh := &QnObjectHandle{
		key:    "key",
		cfg:    cfg,
		bm:     nil,
		mac:    qbox.NewMac(cfg.AccessKey, cfg.SecretKey),
		client: &client.Client{Client: http.DefaultClient},
	}

	// test with timeout
	oh.key = "timeout"
	_, err := oh.NewRangeReader(context.Background(), 0, -1)
	assert.Equal(t, err, fmt.Errorf("qiniu storage: object not exists"))

	// bodyBytes, err := ioutil.ReadAll(body)
	// assert.Equal(t, string(bodyBytes), "mock server ok")
}
