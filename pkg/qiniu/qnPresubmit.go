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
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	//statusJSON is the JSON file that stores build success info
	statusJSON = "finished.json"

	// ArtifactsDirName is the name of directory defined in prow to store test artifacts
	ArtifactsDirName = "artifacts"

	//PostSubmitCoverProfile represents the default output coverage file generated in prow environment
	PostSubmitCoverProfile = "filtered.cov"

	//ChangedProfileName represents the default changed coverage profile based on files changed in Pull Request
	ChangedProfileName = "changed-file-profile.cov"
)

// sortBuilds converts all build from str to int and sorts all builds in descending order and
// returns the sorted slice
func sortBuilds(strBuilds []string) []int {
	var res []int
	for _, buildStr := range strBuilds {
		num, err := strconv.Atoi(buildStr)
		if err != nil {
			log.Printf("Non-int build number found: '%s'", buildStr)
		} else {
			res = append(res, num)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(res)))
	return res
}

type finishedStatus struct {
	Timestamp int
	Passed    bool
}

func isBuildSucceeded(jsonText []byte) bool {
	var status finishedStatus
	err := json.Unmarshal(jsonText, &status)
	return err == nil && status.Passed
}

// FindBaseProfileFromQiniu finds the coverage profile file from the latest healthy build
// stored in given gcs directory
func FindBaseProfileFromQiniu(qc Client, prowJobName, covProfileName string) ([]byte, error) {
	dirOfJob := path.Join("logs", prowJobName)
	prefix := dirOfJob + "/"
	strBuilds, err := qc.ListSubDirs(prefix)
	if err != nil {
		return nil, fmt.Errorf("error listing qiniu objects, prowjob:%v, err:%v", prowJobName, err)
	}
	if len(strBuilds) == 0 {
		log.Printf("no cover profiles found from remote, do nothing")
		return nil, nil
	}
	log.Printf("total sub dirs: %d", len(strBuilds))

	builds := sortBuilds(strBuilds)
	profilePath := ""
	for _, build := range builds {
		buildDirPath := path.Join(dirOfJob, strconv.Itoa(build))
		dirOfStatusJSON := path.Join(buildDirPath, statusJSON)

		statusText, err := qc.ReadObject(dirOfStatusJSON)
		if err != nil {
			log.Printf("Cannot read finished.json (%s) ", dirOfStatusJSON)
		} else if isBuildSucceeded(statusText) {
			artifactsDirPath := path.Join(buildDirPath, ArtifactsDirName)
			profilePath = path.Join(artifactsDirPath, covProfileName)
			break
		}
	}
	if profilePath == "" {
		log.Printf("no cover profiles found from remote job %s, do nothing", prowJobName)
		return nil, nil
	}

	log.Printf("base cover profile path: %s", profilePath)
	return qc.ReadObject(profilePath)
}

type Artifacts interface {
	ProfilePath() string
	CreateChangedProfile() *os.File
	GetChangedProfileName() string
}

// ProfileArtifacts prepresents the rule to store test artifacts in prow
type ProfileArtifacts struct {
	Directory          string
	ProfileName        string
	ChangedProfileName string // create temporary to save changed file related coverage profile
}

// ProfilePath returns a full path for profile
func (a *ProfileArtifacts) ProfilePath() string {
	return path.Join(a.Directory, a.ProfileName)
}

// CreateChangedProfile creates a profile in order to store the most related files based on Github Pull Request
func (a *ProfileArtifacts) CreateChangedProfile() *os.File {
	if a.ChangedProfileName == "" {
		log.Fatalf("param Artifacts.ChangedProfileName should not be empty")
	}
	p, err := os.Create(a.ChangedProfileName)
	log.Printf("os create: %s", a.ChangedProfileName)
	if err != nil {
		log.Fatalf("file(%s) create failed: %v", a.ChangedProfileName, err)
	}

	return p
}

func (a *ProfileArtifacts) GetChangedProfileName() string {
	return a.ChangedProfileName
}
