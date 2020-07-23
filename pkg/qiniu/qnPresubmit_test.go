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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindBaseProfileFromQiniu(t *testing.T) {
	conf := Config{
		Bucket: "artifacts",
	}
	qc, router, _, teardown := MockQiniuServer(&conf)
	defer teardown()
	prowJobName := "kodo-postsubmits-go-st-coverage"
	covProfileName := "filterd.cov"
	mockProfileContent := `mode: atomic
"qiniu.com/kodo/apiserver/server/main.go:32.49,33.13 1 30
"qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 0`

	MockRouterAPI(router, mockProfileContent, 0)
	getProfile, err := FindBaseProfileFromQiniu(qc, prowJobName, covProfileName)
	assert.Equal(t, err, nil)
	assert.Equal(t, string(getProfile), mockProfileContent)
}
