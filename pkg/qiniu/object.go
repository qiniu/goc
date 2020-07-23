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
	"io"
	"net/http"
	"time"

	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/client"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/sirupsen/logrus"
)

// ObjectHandle provides operations on an object in a qiniu cloud bucket
type ObjectHandle struct {
	key    string
	cfg    *Config
	bm     *storage.BucketManager
	mac    *qbox.Mac
	client *client.Client
}

func (o *ObjectHandle) Attrs(ctx context.Context) (storage.FileInfo, error) {
	//TODO(CarlJi): need retry when errors
	return o.bm.Stat(o.cfg.Bucket, o.key)
}

// NewReader creates a reader to read the contents of the object.
// ErrObjectNotExist will be returned if the object is not found.
// The caller must call Close on the returned Reader when done reading.
func (o *ObjectHandle) NewReader(ctx context.Context) (io.ReadCloser, error) {
	return o.NewRangeReader(ctx, 0, -1)
}

// NewRangeReader reads parts of an object, reading at most length bytes starting
// from the given offset. If length is negative, the object is read until the end.
func (o *ObjectHandle) NewRangeReader(ctx context.Context, offset, length int64) (io.ReadCloser, error) {
	verb := "GET"
	if length == 0 {
		verb = "HEAD"
	}

	var res *http.Response
	var err error

	err = runWithRetry(3, func() (bool, error) {
		headers := http.Header{}
		start := offset
		if length < 0 && start >= 0 {
			headers.Set("Range", fmt.Sprintf("bytes=%d-", start))
		} else if length > 0 {
			// The end character isn't affected by how many bytes we have seen.
			headers.Set("Range", fmt.Sprintf("bytes=%d-%d", start, offset+length-1))
		}

		deadline := time.Now().Add(time.Second * 60 * 10).Unix()
		accessURL := storage.MakePrivateURL(o.mac, o.cfg.Domain, o.key, deadline)
		res, err = o.client.DoRequest(ctx, verb, accessURL, headers)
		if err != nil {
			time.Sleep(time.Second) //TODO enhance
			return true, err
		}

		if res.StatusCode == http.StatusNotFound {
			res.Body.Close()
			return true, fmt.Errorf("qiniu storage: object not exists")
		}

		return shouldRetry(res), nil
	})

	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func runWithRetry(maxTry int, f func() (bool, error)) error {
	var err error
	for maxTry > 0 {
		var needRetry bool
		needRetry, err = f() // fix -  needRetry, err := f(), err hides the outside error
		if err != nil {
			logrus.Warnf("err occurred: %v. try again", err)
		} else if needRetry {
			logrus.Warn("results do not meet the expectation. try again")
		} else {
			break
		}
		time.Sleep(time.Millisecond * 100)
		maxTry = maxTry - 1
	}

	return err
}

func shouldRetry(res *http.Response) bool {

	// 571 and 573 mean the request was limited by cloud storage because of concurrency count exceed
	// so it's better to retry after a while
	if res.StatusCode == 571 || res.StatusCode == 573 {
		return true
	}

	return false
}
