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

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type diffFunc func(cmd *cobra.Command, args []string)

func captureStdout(f diffFunc, cmd *cobra.Command, args []string) string {
	r, w, err := os.Pipe()
	if err != nil {
		logrus.WithError(err).Fatal("os pipe fail")
	}
	stdout := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = stdout
	}()

	f(cmd, args)
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

func TestDoDiffForLocalProfiles(t *testing.T) {
	items := []struct {
		newCovFile   string
		newProfile   string
		baseCovFile  string
		baseProfile  string
		expectOutput string
	}{
		{
			newCovFile:  "new.cov",
			baseCovFile: "base.cov",
			newProfile: "mode: atomic\n" +
				"qiniu.com/kodo/apiserver/server/main.go:32.49,33.13 1 30\n" +
				"qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 1\n",
			baseProfile: "mode: atomic\n" +
				"qiniu.com/kodo/apiserver/server/main.go:32.49,33.13 1 30\n" +
				"qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 0\n",
			expectOutput: `+-----------------------------------------+---------------+--------------+-------+
|                  File                   | Base Coverage | New Coverage | Delta |
+-----------------------------------------+---------------+--------------+-------+
| qiniu.com/kodo/apiserver/server/main.go |     50.0%     |    100.0%    | 50.0% |
| Total                                   |     50.0%     |    100.0%    | 50.0% |
+-----------------------------------------+---------------+--------------+-------+
`,
		},
	}

	for _, tc := range items {
		err := ioutil.WriteFile(tc.newCovFile, []byte(tc.newProfile), 0644)
		if err != nil {
			logrus.WithError(err).Fatalf("write file %s failed", tc.newCovFile)
		}
		err = ioutil.WriteFile(tc.baseCovFile, []byte(tc.baseProfile), 0644)
		if err != nil {
			logrus.WithError(err).Fatalf("write file %s failed", tc.baseCovFile)
		}
		defer func() {
			os.Remove(tc.newCovFile)
			os.Remove(tc.baseCovFile)
		}()

		pwd, err := os.Getwd()
		if err != nil {
			logrus.WithError(err).Fatalf("get pwd failed")
		}

		diffCmd.Flags().Set("new-profile", fmt.Sprintf("%s/%s", pwd, tc.newCovFile))
		diffCmd.Flags().Set("base-profile", fmt.Sprintf("%s/%s", pwd, tc.baseCovFile))
		out := captureStdout(doDiffForLocalProfiles, diffCmd, nil)
		assert.Equal(t, out, tc.expectOutput)
	}

}
