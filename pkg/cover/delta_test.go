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

	"github.com/stretchr/testify/assert"
)

func TestGetDeltaCov(t *testing.T) {
	items := []struct {
		newList     CoverageList
		baseList    CoverageList
		expectDelta DeltaCovList
		rows        int
	}{
		//coverage increase
		{
			newList:     CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 20}},
			baseList:    CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 10, NAllStmts: 20}},
			expectDelta: DeltaCovList{{FileName: "fake-coverage", BasePer: "50.0%", NewPer: "75.0%", DeltaPer: "25.0%"}},
			rows:        1,
		},
		//coverage decrease
		{
			newList:     CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 20}},
			baseList:    CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 20, NAllStmts: 20}},
			expectDelta: DeltaCovList{{FileName: "fake-coverage", BasePer: "100.0%", NewPer: "75.0%", DeltaPer: "-25.0%"}},
			rows:        1,
		},
		//diff file
		{
			newList:  CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 20}},
			baseList: CoverageList{Coverage{FileName: "fake-coverage-v1", NCoveredStmts: 10, NAllStmts: 20}},
			expectDelta: DeltaCovList{{FileName: "fake-coverage", BasePer: "None", NewPer: "75.0%", DeltaPer: "75.0%"},
				{FileName: "fake-coverage-v1", BasePer: "50.0%", NewPer: "None", DeltaPer: "-50.0%"}},
			rows: 2,
		},
		//one file has same coverage rate
		{
			newList: CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 20}},
			baseList: CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 20},
				Coverage{FileName: "fake-coverage-v1", NCoveredStmts: 10, NAllStmts: 20}},
			expectDelta: DeltaCovList{{FileName: "fake-coverage-v1", BasePer: "50.0%", NewPer: "None", DeltaPer: "-50.0%"}},
			rows:        1,
		},
	}

	for _, tc := range items {
		d := GetDeltaCov(tc.newList, tc.baseList)
		assert.Equal(t, tc.rows, len(d))
		assert.Equal(t, tc.expectDelta, d)
	}
}

func TestGetChFileDeltaCov(t *testing.T) {
	items := []struct {
		newList      CoverageList
		baseList     CoverageList
		changedFiles []string
		expectDelta  DeltaCovList
	}{
		{
			newList:      CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 20}},
			baseList:     CoverageList{Coverage{FileName: "fake-coverage-v1", NCoveredStmts: 10, NAllStmts: 20}},
			changedFiles: []string{"fake-coverage"},
			expectDelta:  DeltaCovList{{FileName: "fake-coverage", BasePer: "None", NewPer: "75.0%", DeltaPer: "75.0%"}},
		},
	}
	for _, tc := range items {
		d := GetChFileDeltaCov(tc.newList, tc.baseList, tc.changedFiles)
		assert.Equal(t, tc.expectDelta, d)
	}
}

func TestMapAndSort(t *testing.T) {
	items := []struct {
		dList      DeltaCovList
		expectMap  map[string]DeltaCov
		expectSort DeltaCovList
	}{
		{
			dList: DeltaCovList{DeltaCov{FileName: "b", BasePer: "10.0%", NewPer: "20.0%", DeltaPer: "10.0%"},
				DeltaCov{FileName: "a", BasePer: "10.0%", NewPer: "30.0%", DeltaPer: "20.0%"},
			},
			expectMap: map[string]DeltaCov{
				"a": {FileName: "a", BasePer: "10.0%", NewPer: "30.0%", DeltaPer: "20.0%"},
				"b": {FileName: "b", BasePer: "10.0%", NewPer: "20.0%", DeltaPer: "10.0%"},
			},
			expectSort: DeltaCovList{DeltaCov{FileName: "a", BasePer: "10.0%", NewPer: "30.0%", DeltaPer: "20.0%"},
				DeltaCov{FileName: "b", BasePer: "10.0%", NewPer: "20.0%", DeltaPer: "10.0%"},
			},
		},
		{
			dList: DeltaCovList{DeltaCov{FileName: "b", BasePer: "10.0%", NewPer: "20.0%", DeltaPer: "10.0%"},
				DeltaCov{FileName: "b-1", BasePer: "10.0%", NewPer: "30.0%", DeltaPer: "20.0%"},
				DeltaCov{FileName: "1-b", BasePer: "10.0%", NewPer: "40.0%", DeltaPer: "30.0%"},
			},
			expectMap: map[string]DeltaCov{
				"1-b": {FileName: "1-b", BasePer: "10.0%", NewPer: "40.0%", DeltaPer: "30.0%"},
				"b":   {FileName: "b", BasePer: "10.0%", NewPer: "20.0%", DeltaPer: "10.0%"},
				"b-1": {FileName: "b-1", BasePer: "10.0%", NewPer: "30.0%", DeltaPer: "20.0%"},
			},
			expectSort: DeltaCovList{DeltaCov{FileName: "1-b", BasePer: "10.0%", NewPer: "40.0%", DeltaPer: "30.0%"},
				DeltaCov{FileName: "b", BasePer: "10.0%", NewPer: "20.0%", DeltaPer: "10.0%"},
				DeltaCov{FileName: "b-1", BasePer: "10.0%", NewPer: "30.0%", DeltaPer: "20.0%"},
			},
		},
	}

	for _, tc := range items {
		assert.Equal(t, tc.expectMap, tc.dList.Map())
		tc.dList.Sort()
		assert.Equal(t, tc.expectSort, tc.dList)

	}

}
