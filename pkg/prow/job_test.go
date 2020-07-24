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
	"errors"
	"fmt"
	"github.com/qiniu/goc/pkg/cover"
	"github.com/qiniu/goc/pkg/github"
	"github.com/qiniu/goc/pkg/mock/clients"
	"github.com/qiniu/goc/pkg/qiniu"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var (
	defaultContent = `mode: atomic
qiniu.com/kodo/bd/bdgetter/source.go:19.118,22.2 2 0
qiniu.com/kodo/bd/bdgetter/source.go:37.34,39.2 1 0
qiniu.com/kodo/bd/pfd/locker/app/qboxbdlocker/main.go:50.2,53.52 4 1
qiniu.com/kodo/bd/pfd/locker/bdlocker/locker.go:33.51,35.2 1 0`
	defaultLocalPath   = "local.cov"
	defaultChangedPath = "changed.cov"
)

func TestTrimGhFileToProfile(t *testing.T) {
	items := []struct {
		inputFiles  []string
		expectFiles []string
	}{
		{
			inputFiles:  []string{"src/qiniu.com/kodo/io/io/io_svr.go", "README.md"},
			expectFiles: []string{"qiniu.com/kodo/io/io/io_svr.go", "README.md"},
		},
	}

	for _, tc := range items {
		f := trimGhFileToProfile(tc.inputFiles)
		assert.Equal(t, f, tc.expectFiles)
	}
}

func setup(path, content string) {
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		logrus.WithError(err).Fatalf("write file %s failed", path)
	}
}

func TestWriteChangedCov(t *testing.T) {
	path := defaultLocalPath
	savePath := qiniu.ChangedProfileName
	content := defaultContent
	changedFiles := []string{"qiniu.com/kodo/bd/pfd/locker/bdlocker/locker.go"}
	expectContent := `mode: atomic
qiniu.com/kodo/bd/pfd/locker/bdlocker/locker.go:33.51,35.2 1 0
`

	setup(path, content)
	defer os.Remove(path)
	defer os.Remove(savePath)
	j := &Job{
		LocalProfilePath: path,
		LocalArtifacts:   &qiniu.ProfileArtifacts{ChangedProfileName: savePath},
	}
	j.WriteChangedCov(changedFiles)

	r, err := ioutil.ReadFile(savePath)
	if err != nil {
		logrus.WithError(err).Fatalf("read file %s failed", path)
	}
	assert.Equal(t, string(r), expectContent)
}

func TestRunPresubmitFulldiff(t *testing.T) {
	//param
	org := "qbox"
	repo := "kodo"
	prNum := "1"
	buildId := "1266322425771986946"
	jobName := "kodo-pull-integration-test"
	robotName := "qiniu-bot"
	githubCommentPrefix := ""
	githubTokenPath := "token"

	//mock local profile
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	localPath := defaultLocalPath
	localProfileContent := `mode: atomic
"qiniu.com/kodo/apiserver/server/main.go:32.49,33.13 1 30
"qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 0`
	setup(localPath, localProfileContent)
	defer os.Remove(path.Join(pwd, localPath))

	// mock qiniu
	conf := qiniu.Config{
		Bucket: "artifacts",
	}
	qc, router, _, teardown := qiniu.MockQiniuServer(&conf)
	defer teardown()
	qiniu.MockRouterAPI(router, localProfileContent, 0)

	ChangedProfilePath := "changed.cov"
	defer os.Remove(path.Join(pwd, ChangedProfilePath))

	//mock github client
	setup(githubTokenPath, "")
	defer os.Remove(path.Join(pwd, githubTokenPath))
	prClient := github.NewPrClient(githubTokenPath, org, repo, prNum, robotName, githubCommentPrefix)

	j := &Job{
		JobName:                jobName,
		Org:                    org,
		RepoName:               repo,
		PRNumStr:               prNum,
		BuildId:                buildId,
		PostSubmitJob:          "kodo-postsubmits-go-st-coverage",
		PostSubmitCoverProfile: "filterd.cov",
		LocalProfilePath:       localPath,
		LocalArtifacts:         &qiniu.ProfileArtifacts{ChangedProfileName: ChangedProfilePath},
		QiniuClient:            qc,
		GithubComment:          prClient,
		FullDiff:               true,
	}
	defer os.Remove(path.Join(os.Getenv("ARTIFACTS"), j.HtmlProfile()))

	err = j.RunPresubmit()
	assert.NoError(t, err)
}

func TestRunPresubmitError(t *testing.T) {
	items := []struct {
		prepare bool // 是否需要准备本地cov
		j       Job
		err     string
	}{
		{
			prepare: false,
			j: Job{
				LocalProfilePath: "unkown",
			},
			err: "no such file or directory",
		},
		{
			prepare: true,
			j: Job{
				LocalProfilePath: defaultLocalPath,
				QiniuClient:      &clients.MockQnClient{},
			},
		},
		{
			prepare: true,
			j: Job{
				LocalProfilePath: defaultLocalPath,
				QiniuClient:      &clients.MockQnClient{ListSubDirsErr: errors.New("mock error")},
			},
			err: "mock error",
		},
		{
			prepare: true,
			j: Job{
				LocalProfilePath: defaultLocalPath,
				QiniuClient:      &MockProfileQnClient{},
				GithubComment:    &clients.MockPrComment{GetPrChangedFilesRes: []string{"qiniu.com/kodo/apiserver/server/main.go"}},
				FullDiff:         true,
				LocalArtifacts:   &qiniu.ProfileArtifacts{ChangedProfileName: defaultChangedPath},
			},
			err: "",
		},
	}
	for _, tc := range items {
		if tc.prepare {
			path := defaultLocalPath
			setup(path, defaultContent)
			defer os.Remove(path)
			defer os.Remove(defaultChangedPath)
		}
		err := tc.j.RunPresubmit()
		if tc.err == "" {
			assert.NoError(t, err)
		} else {
			assert.Contains(t, err.Error(), tc.err)
		}
	}
}

type MockProfileQnClient struct {
	*clients.MockQnClient
}

func (s *MockProfileQnClient) ListSubDirs(prefix string) ([]string, error) {
	return []string{defaultContent}, nil
}

func (s *MockProfileQnClient) ReadObject(key string) ([]byte, error) {
	logrus.Info(key)
	if key == "logs/1/finished.json" {
		return []byte(`{"timestamp":1590750306,"passed":true,"result":"SUCCESS","repo-version":"76433418ea48aae57af028f9cb2fa3735ce08c7d"}`), nil
	}
	return []byte(""), nil
}

func TestGetFilesAndCovList(t *testing.T) {
	items := []struct {
		fullDiff   bool
		prComment  github.PrComment
		localP     cover.CoverageList
		baseP      cover.CoverageList
		err        string
		lenFiles   int
		lenCovList int
	}{
		{
			fullDiff:  true,
			prComment: &clients.MockPrComment{},
			localP: cover.CoverageList{
				{FileName: "qiniu.com/kodo/apiserver/server/main.go", NCoveredStmts: 2, NAllStmts: 2},
				{FileName: "qiniu.com/kodo/apiserver/server/test.go", NCoveredStmts: 2, NAllStmts: 2},
			},
			baseP: cover.CoverageList{
				{FileName: "qiniu.com/kodo/apiserver/server/main.go", NCoveredStmts: 1, NAllStmts: 2},
				{FileName: "qiniu.com/kodo/apiserver/server/test.go", NCoveredStmts: 1, NAllStmts: 2},
			},
			lenFiles:   2,
			lenCovList: 2,
		},
		{
			fullDiff:  false,
			prComment: &clients.MockPrComment{GetPrChangedFilesErr: errors.New("mock error")},
			err:       "mock error",
		},
		{
			fullDiff:   false,
			prComment:  &clients.MockPrComment{},
			lenFiles:   0,
			lenCovList: 0,
		},
		{
			fullDiff:  false,
			prComment: &clients.MockPrComment{GetPrChangedFilesRes: []string{"qiniu.com/kodo/apiserver/server/main.go"}},
			localP: cover.CoverageList{
				{FileName: "qiniu.com/kodo/apiserver/server/main.go", NCoveredStmts: 2, NAllStmts: 2},
				{FileName: "qiniu.com/kodo/apiserver/server/test.go", NCoveredStmts: 2, NAllStmts: 2},
			},
			baseP: cover.CoverageList{
				{FileName: "qiniu.com/kodo/apiserver/server/main.go", NCoveredStmts: 1, NAllStmts: 2},
				{FileName: "qiniu.com/kodo/apiserver/server/test.go", NCoveredStmts: 1, NAllStmts: 2},
			},
			lenFiles:   1,
			lenCovList: 1,
		},
	}

	for i, tc := range items {
		fmt.Println(i)
		files, covList, err := getFilesAndCovList(tc.fullDiff, tc.prComment, tc.localP, tc.baseP)
		if err != nil {
			assert.Equal(t, err.Error(), tc.err)
		} else {
			assert.Equal(t, len(files), tc.lenFiles)
			assert.Equal(t, len(covList), tc.lenCovList)
		}
	}
}

func TestSetDeltaCovLinks(t *testing.T) {
	covList := cover.DeltaCovList{{FileName: "file1", BasePer: "5%", NewPer: "5%", DeltaPer: "0"}}
	j := &Job{
		QiniuClient: &clients.MockQnClient{},
	}
	j.SetDeltaCovLinks(covList)
}

// functions to be done

func TestRunPostsubmit(t *testing.T) {
	j := &Job{}
	err := j.RunPostsubmit()
	assert.NoError(t, err)
}

func TestRunPeriodic(t *testing.T) {
	j := &Job{}
	err := j.RunPeriodic()
	assert.NoError(t, err)
}

func TestFetch(t *testing.T) {
	j := &Job{}
	res := j.Fetch("buidID", "name")
	assert.Equal(t, res, []byte{})
}
