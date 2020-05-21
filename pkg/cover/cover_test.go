/*
 Copyright 2020 Qiniu Cloud (七牛云)

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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testCoverage() (c *Coverage) {
	return &Coverage{FileName: "fake-coverage", NCoveredStmts: 200, NAllStmts: 300}
}

func TestCoverageRatio(t *testing.T) {
	c := testCoverage()
	actualRatio, _ := c.Ratio()
	assert.Equal(t, float32(c.NCoveredStmts)/float32(c.NAllStmts), actualRatio)
}

func TestRatioErr(t *testing.T) {
	c := &Coverage{FileName: "fake-coverage", NCoveredStmts: 200, NAllStmts: 0}
	_, err := c.Ratio()
	assert.NotNil(t, err)
}

func TestPercentageNA(t *testing.T) {
	c := &Coverage{FileName: "fake-coverage", NCoveredStmts: 200, NAllStmts: 0}
	assert.Equal(t, "N/A", c.Percentage())
}

func TestGenLocalCoverDiffReport(t *testing.T) {
	//coverage increase
	newList := &CoverageList{Groups: []Coverage{{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 20}}}
	baseList := &CoverageList{Groups: []Coverage{{FileName: "fake-coverage", NCoveredStmts: 10, NAllStmts: 20}}}
	rows := GenLocalCoverDiffReport(newList, baseList)
	assert.Equal(t, 1, len(rows))
	assert.Equal(t, []string{"fake-coverage", "50.0%", "75.0%", "25.0%"}, rows[0])

	//coverage decrease
	baseList = &CoverageList{Groups: []Coverage{{FileName: "fake-coverage", NCoveredStmts: 20, NAllStmts: 20}}}
	rows = GenLocalCoverDiffReport(newList, baseList)
	assert.Equal(t, []string{"fake-coverage", "100.0%", "75.0%", "-25.0%"}, rows[0])

	//diff file
	baseList = &CoverageList{Groups: []Coverage{{FileName: "fake-coverage-v1", NCoveredStmts: 10, NAllStmts: 20}}}
	rows = GenLocalCoverDiffReport(newList, baseList)
	assert.Equal(t, []string{"fake-coverage", "None", "75.0%", "75.0%"}, rows[0])
}

func TestCovList(t *testing.T) {
	fileName := "qiniu.com/kodo/apiserver/server/main.go"

	// percentage is 100%
	p := strings.NewReader("mode: atomic\n" +
		fileName + ":32.49,33.13 1 30\n")
	covL, err := CovList(p)
	covF := covL.Map()[fileName]
	assert.Nil(t, err)
	assert.Equal(t, "100.0%", covF.Percentage())

	// percentage is 50%
	p = strings.NewReader("mode: atomic\n" +
		fileName + ":32.49,33.13 1 30\n" +
		fileName + ":42.49,43.13 1 0\n")
	covL, err = CovList(p)
	covF = covL.Map()[fileName]
	assert.Nil(t, err)
	assert.Equal(t, "50.0%", covF.Percentage())

	// two files
	fileName1 := "qiniu.com/kodo/apiserver/server/svr.go"
	p = strings.NewReader("mode: atomic\n" +
		fileName + ":32.49,33.13 1 30\n" +
		fileName1 + ":42.49,43.13 1 0\n")
	covL, err = CovList(p)
	covF = covL.Map()[fileName]
	covF1 := covL.Map()[fileName1]
	assert.Nil(t, err)
	assert.Equal(t, "100.0%", covF.Percentage())
	assert.Equal(t, "0.0%", covF1.Percentage())
}
