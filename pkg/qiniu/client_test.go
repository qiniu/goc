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

import "testing"

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
