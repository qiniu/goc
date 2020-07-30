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

package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-github/github"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/qiniu/goc/pkg/cover"
)

const (
	// baseURLPath is a non-empty Client.BaseURL path to use during tests,
	// to ensure relative URLs are used for all endpoints. See issue #752.
	baseURLPath = "/api-v3"
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() (client *github.Client, router *httprouter.Router, serverURL string, teardown func()) {
	// router is the HTTP request multiplexer used with the test server.
	router = httprouter.New()

	// We want to ensure that tests catch mistakes where the endpoint URL is
	// specified as absolute rather than relative. It only makes a difference
	// when there's a non-empty base URL path. So, use that. See issue #752.
	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, router))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(os.Stderr, "FAIL: Client.BaseURL path prefix is not preserved in the request URL:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\t"+req.URL.String())
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tDid you accidentally use an absolute endpoint URL rather than relative?")
		fmt.Fprintln(os.Stderr, "\tSee https://github.com/google/go-github/issues/752 for information.")
		http.Error(w, "Client.BaseURL path prefix is not preserved in the request URL.", http.StatusInternalServerError)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	// client is the GitHub client being tested and is
	// configured to use test server.
	client = github.NewClient(nil)
	url, _ := url.Parse(server.URL + baseURLPath + "/")
	client.BaseURL = url
	client.UploadURL = url

	return client, router, server.URL, server.Close
}

func TestNewPrClient(t *testing.T) {
	items := []struct {
		token       string
		repoOwner   string
		repoName    string
		prNumStr    string
		botUserName string
		commentFlag string
		expectPrNum int
	}{
		{token: "github_test.go", repoOwner: "qiniu", repoName: "goc", prNumStr: "1", botUserName: "qiniu-bot", commentFlag: "test", expectPrNum: 1},
	}

	for _, tc := range items {
		prClient := NewPrClient(tc.token, tc.repoOwner, tc.repoName, tc.prNumStr, tc.botUserName, tc.commentFlag)
		assert.Equal(t, tc.expectPrNum, prClient.PrNumber)
	}
}

func TestCreateGithubComment(t *testing.T) {
	client, router, _, teardown := setup()
	defer teardown()

	var coverList = cover.DeltaCovList{{FileName: "fake-coverage", BasePer: "50.0%", NewPer: "75.0%", DeltaPer: "25.0%"}}
	expectContent := GenCommentContent("", coverList)
	comment := &github.IssueComment{
		Body: &expectContent,
	}

	// create comment: https://developer.github.com/v3/issues/comments/#create-a-comment
	router.HandlerFunc("POST", "/repos/qiniu/goc/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.IssueComment)
		json.NewDecoder(r.Body).Decode(v)
		assert.Equal(t, v, comment)

		fmt.Fprint(w, `{"id":1}`)
	})

	// list comment: https://developer.github.com/v3/issues/comments/#list-comments-on-an-issue
	router.HandlerFunc("GET", "/repos/qiniu/goc/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"id":1,"user": {"login": "qiniu-bot"}}]`)
	})

	// delete comment: https://developer.github.com/v3/issues/comments/#edit-a-comment
	router.HandlerFunc("DELETE", "/repos/qiniu/goc/issues/comments/1", func(w http.ResponseWriter, r *http.Request) {
	})

	p := GithubPrComment{
		RobotUserName: "qiniu-bot",
		RepoOwner:     "qiniu",
		RepoName:      "goc",
		CommentFlag:   "",
		PrNumber:      1,
		Ctx:           context.Background(),
		opt:           nil,
		GithubClient:  client,
	}

	p.CreateGithubComment("", coverList)
}

func TestCreateGithubCommentError(t *testing.T) {
	p := &GithubPrComment{}
	err := p.CreateGithubComment("", cover.DeltaCovList{})
	assert.NoError(t, err)
}

func TestGetPrChangedFiles(t *testing.T) {
	client, router, _, teardown := setup()
	defer teardown()

	var expectFiles = []string{"src/qiniu.com/kodo/s3apiv2/bucket/bucket.go"}

	// list files API: https://developer.github.com/v3/pulls/#list-pull-requests-files
	router.HandlerFunc("GET", "/repos/qiniu/goc/pulls/1/files", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"filename":"src/qiniu.com/kodo/s3apiv2/bucket/bucket.go"}]`)
	})

	p := GithubPrComment{
		RobotUserName: "qiniu-bot",
		RepoOwner:     "qiniu",
		RepoName:      "goc",
		CommentFlag:   "",
		PrNumber:      1,
		Ctx:           context.Background(),
		opt:           nil,
		GithubClient:  client,
	}
	changedFiles, err := p.GetPrChangedFiles()
	assert.Equal(t, err, nil)
	assert.Equal(t, changedFiles, expectFiles)
}

func TestGetCommentFlag(t *testing.T) {
	p := GithubPrComment{
		CommentFlag: "flag",
	}
	flag := p.GetCommentFlag()
	assert.Equal(t, flag, p.CommentFlag)
}
