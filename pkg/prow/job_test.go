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
	defaultLocalPath = "local.cov"
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
		LocalArtifacts:   &qiniu.Artifacts{ChangedProfileName: savePath},
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
		LocalArtifacts:         &qiniu.Artifacts{ChangedProfileName: ChangedProfilePath},
		QiniuClient:            qc,
		GithubComment:          prClient,
		FullDiff:               true,
	}
	defer os.Remove(path.Join(os.Getenv("ARTIFACTS"), j.HtmlProfile()))

	err = j.RunPresubmit()
	assert.NoError(t, err)
}

func TestRunPresubmitErrorGetRemoteCoverProfile(t *testing.T) {
	j := &Job{
		LocalProfilePath: "unkown",
	}
	err := j.RunPresubmit()
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestRunPresubmitGetEmptyProfile(t *testing.T) {
	path := "local.cov"
	setup(path, defaultContent)
	defer os.Remove(path)
	j := &Job{
		LocalProfilePath: path,
		QiniuClient:      &clients.MockQnClient{},
	}
	err := j.RunPresubmit()
	assert.NoError(t, err)
}

type MockErrorListSubDirsQnClient struct {
	*clients.MockQnClient
}

func (s *MockErrorListSubDirsQnClient) ListSubDirs(prefix string) ([]string, error) {
	return nil, errors.New("mock error")
}

func TestRunPresubmitErrorGetBaseProfile(t *testing.T) {
	path := "local.cov"
	setup(path, defaultContent)
	defer os.Remove(path)
	j := &Job{
		LocalProfilePath: path,
		QiniuClient:      &MockErrorListSubDirsQnClient{},
	}
	err := j.RunPresubmit()
	assert.Contains(t, err.Error(), "mock error")
}

type MockProfileQnClient struct {
	*clients.MockQnClient
}

func (s *MockProfileQnClient) ListSubDirs(prefix string) ([]string, error) {
	return []string{"1"}, nil
}

func (s *MockProfileQnClient) ReadObject(key string) ([]byte, error) {
	logrus.Info(key)
	if key == "logs/1/finished.json" {
		return []byte(`{"timestamp":1590750306,"passed":true,"result":"SUCCESS","repo-version":"76433418ea48aae57af028f9cb2fa3735ce08c7d"}`), nil
	}
	return []byte(""), nil
}

// 无法实现coverList报错
//func TestRunPresubmitErrorReadBaseProfile(t *testing.T) {
//	path := "local.cov"
//	setup(path, defaultContent)
//	defer os.Remove(path)
//	j := &Job{
//		LocalProfilePath: path,
//		QiniuClient:      &MockProfileQnClient{},
//	}
//	err := j.RunPresubmit()
//	assert.Contains(t, err.Error(), "mock error")
//}

//type MockErrorPrComment struct {
//	*clients.MockPrComment
//}
//
//func (s *MockErrorPrComment) GetPrChangedFiles() (files []string, err error) {
//	return []string{"aaa"}, nil
//}
//
//func TestGetFilesAndCovList(t *testing.T){
//	items := []struct {
//		fullDiff bool
//		prComment github.PrComment
//		localP cover.CoverageList
//		baseP cover.CoverageList
//	}{
//		{
//			inputFiles:  []string{"src/qiniu.com/kodo/io/io/io_svr.go", "README.md"},
//			expectFiles: []string{"qiniu.com/kodo/io/io/io_svr.go", "README.md"},
//		},
//	}
//
//	for _, tc := range items {
//		f := trimGhFileToProfile(tc.inputFiles)
//		assert.Equal(t, f, tc.expectFiles)
//	}
//	getFilesAndCovList
//}
