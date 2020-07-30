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

package prow

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/qiniu/goc/pkg/github"
	"github.com/qiniu/goc/pkg/qiniu"
)

// IProwAction defines the normal action in prow system
type IProwAction interface {
	Fetch(BuildID, name string) []byte
	RunPresubmit() error
	RunPostsubmit() error
	RunPeriodic() error
}

// Job is a prowjob in prow
type Job struct {
	JobName                string
	Org                    string
	RepoName               string
	PRNumStr               string
	BuildId                string //prow job build number
	PostSubmitJob          string
	PostSubmitCoverProfile string
	CovThreshold           int
	LocalProfilePath       string
	QiniuClient            qiniu.Client
	LocalArtifacts         qiniu.Artifacts
	GithubComment          github.PrComment
	FullDiff               bool
}

// Fetch the file from cloud
func (j *Job) Fetch(BuildID, name string) []byte {
	return []byte{}
}

// RunPresubmit run a presubmit job
func (j *Job) RunPresubmit() error {
	// step1: get local profile cov
	localP, err := cover.ReadFileToCoverList(j.LocalProfilePath)
	if err != nil {
		return fmt.Errorf("failed to get remote cover profile: %s", err.Error())
	}

	//step2: find the remote healthy cover profile from qiniu bucket
	remoteProfile, err := qiniu.FindBaseProfileFromQiniu(j.QiniuClient, j.PostSubmitJob, j.PostSubmitCoverProfile)
	if err != nil {
		return fmt.Errorf("failed to get remote cover profile: %s", err.Error())
	}
	if remoteProfile == nil {
		logrus.Infof("get non healthy remoteProfile, do nothing")
		return nil
	}
	baseP, err := cover.CovList(bytes.NewReader(remoteProfile))
	if err != nil {
		return fmt.Errorf("failed to get remote cover profile: %s", err.Error())
	}

	// step3: get github pull request changed files' name and calculate diff cov between local and remote profile
	changedFiles, deltaCovList, err := getFilesAndCovList(j.FullDiff, j.GithubComment, localP, baseP)
	if err != nil {
		return fmt.Errorf("Get files and covlist failed: %s", err.Error())
	}

	// step4: generate changed file html coverage
	err = j.WriteChangedCov(changedFiles)
	if err != nil {
		return fmt.Errorf("filter local profile to %s with changed files failed: %s", j.LocalArtifacts.GetChangedProfileName(), err.Error())
	}
	err = j.CreateChangedCovHtml()
	if err != nil {
		return fmt.Errorf("create changed file related coverage html failed: %s", err.Error())
	}
	j.SetDeltaCovLinks(deltaCovList)

	// step5: post comment to github
	commentPrefix := github.CommentsPrefix
	if j.GithubComment.GetCommentFlag() != "" {
		commentPrefix = fmt.Sprintf("**%s** ", j.GithubComment.GetCommentFlag()) + commentPrefix
	}
	if len(deltaCovList) > 0 {
		totalDelta := cover.PercentStr(cover.TotalDelta(localP, baseP))
		deltaCovList = append(deltaCovList, cover.DeltaCov{FileName: "Total", BasePer: baseP.TotalPercentage(), NewPer: localP.TotalPercentage(), DeltaPer: totalDelta})
	}
	err = j.GithubComment.CreateGithubComment(commentPrefix, deltaCovList)
	if err != nil {
		return fmt.Errorf("Post comment to github failed: %s", err.Error())
	}

	return nil
}

// RunPostsubmit run a postsubmit job
func (j *Job) RunPostsubmit() error {
	return nil
}

// RunPeriodic run a periodic job
func (j *Job) RunPeriodic() error {
	return nil
}

//trim github filename to profile format:
//	src/qiniu.com/kodo/io/io/io_svr.go -> qiniu.com/kodo/io/io/io_svr.go
func trimGhFileToProfile(ghFiles []string) (pFiles []string) {
	//TODO: need compatible other situation
	logrus.Infof("trim PR changed file name to:")
	for _, f := range ghFiles {
		file := strings.TrimPrefix(f, "src/")
		logrus.Infof("%s", file)
		pFiles = append(pFiles, file)
	}
	return
}

// filter local profile with changed files and save to j.LocalArtifacts.ChangedProfileName
func (j *Job) WriteChangedCov(changedFiles []string) error {
	p, err := ioutil.ReadFile(j.LocalProfilePath)
	if err != nil {
		logrus.Printf("Open file %s failed", j.LocalProfilePath)
		return err
	}
	cp := j.LocalArtifacts.CreateChangedProfile()
	defer cp.Close()
	s := bufio.NewScanner(bytes.NewReader(p))
	s.Scan()
	writeLine(cp, s.Text())

	for s.Scan() {
		for _, file := range changedFiles {
			if strings.HasPrefix(s.Text(), file) {
				writeLine(cp, s.Text())
			}
		}
	}

	return nil
}

// writeLine writes a line in the given file, if the file pointer is not nil
func writeLine(file *os.File, content string) {
	if file != nil {
		fmt.Fprintln(file, content)
	}
}

func (j *Job) JobPrefixOnQiniu() string {
	return path.Join("pr-logs", "pull", j.Org+"_"+j.RepoName, j.PRNumStr, j.JobName, j.BuildId)
}

func (j *Job) HtmlProfile() string {
	return fmt.Sprintf("%s-%s-pr%s-coverage.html", j.Org, j.RepoName, j.PRNumStr)
}

func (j *Job) SetDeltaCovLinks(c cover.DeltaCovList) {
	c.Sort()
	for i := 0; i < len(c); i++ {
		qnKey := path.Join(j.JobPrefixOnQiniu(), "artifacts", j.HtmlProfile())
		authQnKey := j.QiniuClient.GetAccessURL(qnKey, time.Hour*24*7)
		c[i].SetLineCovLink(authQnKey + "#file" + strconv.Itoa(i))
		logrus.Printf("file %s html coverage link is: %s\n", c[i].FileName, c[i].GetLineCovLink())
	}
}

// CreateChangedCovHtml create changed file related coverage html base on the local artifact
func (j *Job) CreateChangedCovHtml() error {
	if j.LocalArtifacts.GetChangedProfileName() == "" {
		logrus.Errorf("param LocalArtifacts.ChangedProfileName is empty")
	}
	pathProfileCov := j.LocalArtifacts.GetChangedProfileName()
	pathHtmlCov := path.Join(os.Getenv("ARTIFACTS"), j.HtmlProfile())
	cmdTxt := fmt.Sprintf("go tool cover -html=%s -o %s", pathProfileCov, pathHtmlCov)
	logrus.Printf("Running command '%s'\n", cmdTxt)
	cmd := exec.Command("go", "tool", "cover", "-html="+pathProfileCov, "-o", pathHtmlCov)
	stdOut, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Printf("Error executing cmd: %v; combinedOutput=%s", err, stdOut)
	}
	return err
}

func getFilesAndCovList(fullDiff bool, prComment github.PrComment, localP, baseP cover.CoverageList) (changedFiles []string, deltaCovList cover.DeltaCovList, err error) {
	if !fullDiff {
		// get github pull request changed files' name
		var ghChangedFiles, err = prComment.GetPrChangedFiles()
		if err != nil {
			return nil, nil, fmt.Errorf("Get pull request changed file failed: %s", err.Error())
		}
		if len(ghChangedFiles) == 0 {
			logrus.Printf("0 files changed in github pull request, don't need to run coverage profile in presubmit.\n")
			return nil, nil, nil
		}
		changedFiles = trimGhFileToProfile(ghChangedFiles)

		// calculate diff cov between local and remote profile
		deltaCovList = cover.GetChFileDeltaCov(localP, baseP, changedFiles)
		logrus.Printf("Get changed files and delta cover list success. ChangedFiles: [%+v], DeltaCovList: [%+v]", changedFiles, deltaCovList)
		return changedFiles, deltaCovList, nil
	}
	deltaCovList = cover.GetDeltaCov(localP, baseP)
	for _, d := range deltaCovList {
		changedFiles = append(changedFiles, d.FileName)
	}
	logrus.Printf("Get all files and delta cover list success. Files: [%+v], DeltaCovList: [%+v]", changedFiles, deltaCovList)
	return changedFiles, deltaCovList, nil
}
