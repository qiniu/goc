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
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/julienschmidt/httprouter"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/sirupsen/logrus"
)

func MockQiniuServer(config *Config) (client *Client, router *httprouter.Router, serverURL string, teardown func()) {
	// router is the HTTP request multiplexer used with the test server.
	router = httprouter.New()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(router)

	config.Domain = server.URL
	client = NewClient(config)
	client.BucketManager.Cfg = &storage.Config{
		RsfHost: server.URL,
	}

	logrus.Infof("server url is: %s", server.URL)
	return client, router, server.URL, server.Close
}

func MockRouterAPI(router *httprouter.Router, profile string) {
	// mock rsf /v2/list
	router.HandlerFunc("POST", "/v2/list", func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("request url is: %s", r.URL.String())

		fmt.Fprint(w, `{
	"item": {
		"key": "logs/kodo-postsubmits-go-st-coverage/1181915661132107776/finished.json",
		"hash": "FkBhdo9odL2Xjvu-YdwtDIw79fIL",
		"fsize": 51523,
		"mimeType": "application/octet-stream",
		"putTime": 15909068578047958,
		"type": 0,
		"status": 0,
		"md5": "e0bd20e97ea1c6a5e2480192ee3ae884"
	},
	"marker": "",
	"dir": "logs/kodo-postsubmits-go-st-coverage/1181915661132107776/"
}`)
	})

	// mock io get statusJSON file
	router.HandlerFunc("GET", "/logs/kodo-postsubmits-go-st-coverage/1181915661132107776/finished.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"timestamp":1590750306,"passed":true,"result":"SUCCESS","repo-version":"76433418ea48aae57af028f9cb2fa3735ce08c7d"}`)
	})

	// mock io get remote coverage profile
	router.HandlerFunc("GET", "/logs/kodo-postsubmits-go-st-coverage/1181915661132107776/artifacts/filterd.cov", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, profile)
	})

}
