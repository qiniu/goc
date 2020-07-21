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
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBuildId(t *testing.T) {
	type tc struct {
		dir      string
		expected string
	}

	tcs := []tc{
		{dir: "logs/kodo-periodics-integration-test/1181915661132107776/", expected: "1181915661132107776"},
		{dir: "logs/kodo-periodics-integration-test/1181915661132107776", expected: ""},
		{dir: "pr-logs/directory/WIP-qtest-pull-request-kodo-test/1181915661132107776/", expected: "1181915661132107776"},
		{dir: "pr-logs/directory/WIP-qtest-pull-request-kodo-test/1181915661132107776.txt", expected: ""},
	}

	for _, tc := range tcs {
		got := getBuildId(tc.dir)
		if tc.expected != got {
			t.Errorf("getBuildId error, dir: %s, expect: %s, but got: %s", tc.dir, tc.expected, got)
		}
	}
}

// test basic listEntries function
func TestListAllFiles(t *testing.T) {
	conf := Config{
		Bucket: "artifacts",
	}
	qc, router, _, teardown := MockQiniuServer(&conf)
	defer teardown()
	prowJobName := "kodo-postsubmits-go-st-coverage"
	dirOfJob := path.Join("logs", prowJobName)
	prefix := dirOfJob + "/"

	MockRouterListAllAPI(router, 0)
	listItems, err := qc.listEntries(prefix, "/")

	assert.Equal(t, err, nil)
	assert.Equal(t, len(listItems), 1)
	assert.Equal(t, listItems[0].Key, "logs/kodo-postsubmits-go-st-coverage/1181915661132107776/finished.json")
}

// test basic listEntries function, recover after 3 times
func TestListAllFilesWithServerTimeoutAndRecover(t *testing.T) {
	conf := Config{
		Bucket: "artifacts",
	}
	qc, router, _, teardown := MockQiniuServer(&conf)
	defer teardown()
	prowJobName := "kodo-postsubmits-go-st-coverage"
	dirOfJob := path.Join("logs", prowJobName)
	prefix := dirOfJob + "/"

	// recover after 3 times
	MockRouterListAllAPI(router, 3)
	listItems, err := qc.listEntries(prefix, "/")

	assert.Equal(t, err, nil)
	assert.Equal(t, len(listItems), 1)
	assert.Equal(t, listItems[0].Key, "logs/kodo-postsubmits-go-st-coverage/1181915661132107776/finished.json")
}

// test basic listEntries function, never recover
func TestListAllFilesWithServerTimeout(t *testing.T) {
	conf := Config{
		Bucket: "artifacts",
	}
	qc, router, _, teardown := MockQiniuServer(&conf)
	defer teardown()
	prowJobName := "kodo-postsubmits-go-st-coverage"
	dirOfJob := path.Join("logs", prowJobName)
	prefix := dirOfJob + "/"

	// never recover
	MockRouterListAllAPI(router, 13)
	_, err := qc.listEntries(prefix, "/")

	assert.Equal(t, strings.Contains(err.Error(), "timed out: error accessing QINIU artifact"), true)
}

// test ListAll function
func TestListAllFilesWithContext(t *testing.T) {
	conf := Config{
		Bucket: "artifacts",
	}
	qc, router, _, teardown := MockQiniuServer(&conf)
	defer teardown()
	prowJobName := "kodo-postsubmits-go-st-coverage"
	dirOfJob := path.Join("logs", prowJobName)
	prefix := dirOfJob + "/"

	MockRouterListAllAPI(router, 0)
	listItems, err := qc.ListAll(context.Background(), prefix, "/")

	assert.Equal(t, err, nil)
	assert.Equal(t, len(listItems), 1)
	assert.Equal(t, listItems[0], "logs/kodo-postsubmits-go-st-coverage/1181915661132107776/finished.json")
}

// test GetArtifactDetails function
func TestGetArtifactDetails(t *testing.T) {
	conf := Config{
		Bucket: "artifacts",
	}
	qc, router, _, teardown := MockQiniuServer(&conf)
	defer teardown()
	prowJobName := "kodo-postsubmits-go-st-coverage"
	dirOfJob := path.Join("logs", prowJobName)
	prefix := dirOfJob + "/"

	MockRouterListAllAPI(router, 0)
	tmpl, err := qc.GetArtifactDetails(prefix)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(tmpl.Items), 1)
	assert.Equal(t, tmpl.Items[0].Name, "1181915661132107776/finished.json")
	assert.Equal(t, strings.Contains(tmpl.Items[0].Url, prowJobName), true)
}

// test ListSubDirs function, recover after 3 times
func TestListSubDirsWithServerTimeoutAndRecover(t *testing.T) {
	conf := Config{
		Bucket: "artifacts",
	}
	qc, router, _, teardown := MockQiniuServer(&conf)
	defer teardown()
	prowJobName := "kodo-postsubmits-go-st-coverage"
	dirOfJob := path.Join("logs", prowJobName)
	prefix := dirOfJob + "/"
	localProfileContent := `mode: atomic
"qiniu.com/kodo/apiserver/server/main.go:32.49,33.13 1 30
"qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 0`
	// recover after 3 times
	MockRouterAPI(router, localProfileContent, 3)
	listItems, err := qc.ListSubDirs(prefix)

	assert.Equal(t, err, nil)
	assert.Equal(t, len(listItems), 1)
	assert.Equal(t, listItems[0], "1181915661132107776")
}

// test ListSubDirs function, never recover
func TestListSubDirsWithServerTimeout(t *testing.T) {
	conf := Config{
		Bucket: "artifacts",
	}
	qc, router, _, teardown := MockQiniuServer(&conf)
	defer teardown()
	prowJobName := "kodo-postsubmits-go-st-coverage"
	dirOfJob := path.Join("logs", prowJobName)
	prefix := dirOfJob + "/"
	localProfileContent := `mode: atomic
"qiniu.com/kodo/apiserver/server/main.go:32.49,33.13 1 30
"qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 0`
	// never recover
	MockRouterAPI(router, localProfileContent, 13)
	_, err := qc.ListSubDirs(prefix)

	assert.Equal(t, strings.Contains(err.Error(), "timed out: error accessing QINIU artifact"), true)
}
