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

import "sort"

// DeltaCov contains the info of a delta coverage
type DeltaCov struct {
	FileName    string
	BasePer     string
	NewPer      string
	DeltaPer    string
	LineCovLink string
}

// DeltaCovList is the list of DeltaCov
type DeltaCovList []DeltaCov

// GetFullDeltaCov get full delta coverage between new and base profile
func GetFullDeltaCov(newList CoverageList, baseList CoverageList) (delta DeltaCovList) {
	newMap := newList.Map()
	baseMap := baseList.Map()

	for file, n := range newMap {
		b, ok := baseMap[file]
		//if the file not in base profile, set None
		if !ok {
			delta = append(delta, DeltaCov{
				FileName: file,
				BasePer:  "None",
				NewPer:   n.Percentage(),
				DeltaPer: PercentStr(Delta(n, b))})
			continue
		}
		delta = append(delta, DeltaCov{
			FileName: file,
			BasePer:  b.Percentage(),
			NewPer:   n.Percentage(),
			DeltaPer: PercentStr(Delta(n, b))})
	}

	for file, b := range baseMap {
		//if the file not in new profile, set None
		if n, ok := newMap[file]; !ok {
			delta = append(delta, DeltaCov{
				FileName: file,
				BasePer:  b.Percentage(),
				NewPer:   "None",
				DeltaPer: PercentStr(Delta(n, b))})
		}
	}
	return
}

// GetDeltaCov get two profile diff cov
func GetDeltaCov(newList CoverageList, baseList CoverageList) (delta DeltaCovList) {
	d := GetFullDeltaCov(newList, baseList)
	for _, v := range d {
		if v.DeltaPer == "0.0%" {
			continue
		}
		delta = append(delta, v)
	}
	return
}

// GetChFileDeltaCov get two profile diff cov of changed files
func GetChFileDeltaCov(newList CoverageList, baseList CoverageList, changedFiles []string) (list DeltaCovList) {
	d := GetFullDeltaCov(newList, baseList)
	dMap := d.Map()
	for _, file := range changedFiles {
		if _, ok := dMap[file]; ok {
			list = append(list, dMap[file])
		}
	}
	return
}

// Delta calculate two coverage delta
func Delta(new Coverage, base Coverage) float32 {
	baseRatio, _ := base.Ratio()
	newRatio, _ := new.Ratio()
	return newRatio - baseRatio
}

// TotalDelta calculate two coverage delta
func TotalDelta(new CoverageList, base CoverageList) float32 {
	baseRatio, _ := base.TotalRatio()
	newRatio, _ := new.TotalRatio()
	return newRatio - baseRatio
}

// Map returns maps the file name to its DeltaCov for faster retrieval & membership check
func (d DeltaCovList) Map() map[string]DeltaCov {
	m := make(map[string]DeltaCov)
	for _, c := range d {
		m[c.FileName] = c
	}
	return m
}

// Sort sort DeltaCovList with filenames
func (d DeltaCovList) Sort() {
	sort.SliceStable(d, func(i, j int) bool {
		return d[i].Name() < d[j].Name()
	})
}

// Name returns the file name
func (c *DeltaCov) Name() string {
	return c.FileName
}

// GetLineCovLink get the LineCovLink of the DeltaCov
func (c *DeltaCov) GetLineCovLink() string {
	return c.LineCovLink
}

// SetLineCovLink set LineCovLink of the DeltaCov
func (c *DeltaCov) SetLineCovLink(link string) {
	c.LineCovLink = link
}
