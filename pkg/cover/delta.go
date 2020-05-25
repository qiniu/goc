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

type GroupChanges struct {
	Added     []Coverage
	Deleted   []Coverage
	Unchanged []Coverage
	Changed   []Incremental
	BaseGroup *CoverageList
	NewGroup  *CoverageList
}

type Incremental struct {
	base Coverage
	new  Coverage
}

func GenLocalCoverDiffReport(newList *CoverageList, baseList *CoverageList) [][]string {
	var rows [][]string
	basePMap := baseList.Map()

	for _, l := range newList.Groups {
		baseCov, ok := basePMap[l.Name()]
		if !ok {
			rows = append(rows, []string{l.FileName, "None", l.Percentage(), PercentStr(Delta(l, baseCov))})
			continue
		}
		if l.Percentage() == baseCov.Percentage() {
			continue
		}
		rows = append(rows, []string{l.FileName, baseCov.Percentage(), l.Percentage(), PercentStr(Delta(l, baseCov))})
	}

	return rows
}

func Delta(new Coverage, base Coverage) float32 {
	baseRatio, _ := base.Ratio()
	newRatio, _ := new.Ratio()
	return newRatio - baseRatio
}
