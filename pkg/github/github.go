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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/qiniu/goc/pkg/cover"
)

// CommentsPrefix is the prefix when commenting on Github Pull Requests
// It is also the flag when checking whether the target comment exists or not to avoid duplicate
const CommentsPrefix = "The following is the coverage report on the affected files."

// PrComment is the interface of the entry which is able to comment on Github Pull Requests
type PrComment interface {
	CreateGithubComment(commentPrefix string, diffCovList cover.DeltaCovList) (err error)
	PostComment(content, commentPrefix string) error
	EraseHistoryComment(commentPrefix string) error
	GetPrChangedFiles() (files []string, err error)
	GetCommentFlag() string
}

// GithubPrComment is the entry which is able to comment on Github Pull Requests
type GithubPrComment struct {
	RobotUserName string
	RepoOwner     string
	RepoName      string
	CommentFlag   string
	PrNumber      int
	Ctx           context.Context
	opt           *github.ListOptions
	GithubClient  *github.Client
}

// NewPrClient creates an Client which be able to comment on Github Pull Request
func NewPrClient(githubTokenPath, repoOwner, repoName, prNumStr, botUserName, commentFlag string) *GithubPrComment {
	var client *github.Client

	// performs automatic retries when connection error occurs or a 500-range response code received (except 501)
	retryClient := retryablehttp.NewClient()
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, retryClient.StandardClient())

	prNum, err := strconv.Atoi(prNumStr)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to convert prNumStr(=%v) to int.\n", prNumStr)
	}
	token, err := ioutil.ReadFile(githubTokenPath)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to get github token.\n")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.TrimSpace(string(token))},
	)
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)

	return &GithubPrComment{
		RobotUserName: botUserName,
		RepoOwner:     repoOwner,
		RepoName:      repoName,
		PrNumber:      prNum,
		CommentFlag:   commentFlag,
		Ctx:           ctx,
		opt:           &github.ListOptions{Page: 1},
		GithubClient:  client,
	}
}

// CreateGithubComment post github comment of diff coverage
func (c *GithubPrComment) CreateGithubComment(commentPrefix string, diffCovList cover.DeltaCovList) (err error) {
	if len(diffCovList) == 0 {
		logrus.Printf("Detect 0 files coverage diff, will not comment to github.")
		return nil
	}
	content := GenCommentContent(commentPrefix, diffCovList)

	err = c.PostComment(content, commentPrefix)
	if err != nil {
		logrus.WithError(err).Fatalf("Post comment to github failed.")
	}

	return
}

// PostComment post comment on github. It erased the old one if existed to avoid duplicate
func (c *GithubPrComment) PostComment(content, commentPrefix string) error {
	//step1: erase history similar comment to avoid too many comment for same job
	err := c.EraseHistoryComment(commentPrefix)
	if err != nil {
		return err
	}

	//step2: post comment with new result
	comment := &github.IssueComment{
		Body: &content,
	}
	_, _, err = c.GithubClient.Issues.CreateComment(c.Ctx, c.RepoOwner, c.RepoName, c.PrNumber, comment)
	if err != nil {
		return err
	}

	return nil
}

// EraseHistoryComment erase history similar comment before post again
func (c *GithubPrComment) EraseHistoryComment(commentPrefix string) error {
	comments, _, err := c.GithubClient.Issues.ListComments(c.Ctx, c.RepoOwner, c.RepoName, c.PrNumber, nil)
	if err != nil {
		logrus.Errorf("list PR comments failed.")
		return err
	}
	logrus.Infof("the count of history comments by %s is: %v", c.RobotUserName, len(comments))

	for _, cm := range comments {
		if *cm.GetUser().Login == c.RobotUserName && strings.HasPrefix(cm.GetBody(), commentPrefix) {
			_, err = c.GithubClient.Issues.DeleteComment(c.Ctx, c.RepoOwner, c.RepoName, *cm.ID)
			if err != nil {
				logrus.Errorf("delete PR comments %d failed.", *cm.ID)
				return err
			}
		}
	}

	return nil
}

// GetPrChangedFiles get github pull request changes file list
func (c *GithubPrComment) GetPrChangedFiles() (files []string, err error) {
	var commitFiles []*github.CommitFile
	for {
		f, resp, err := c.GithubClient.PullRequests.ListFiles(c.Ctx, c.RepoOwner, c.RepoName, c.PrNumber, c.opt)
		if err != nil {
			logrus.Errorf("Get PR changed file failed. repoOwner is: %s, repoName is: %s, prNum is: %d", c.RepoOwner, c.RepoName, c.PrNumber)
			return nil, err
		}
		commitFiles = append(commitFiles, f...)
		if resp.NextPage == 0 {
			break
		}
		c.opt.Page = resp.NextPage
	}
	logrus.Infof("get %d PR changed files:", len(commitFiles))
	for _, file := range commitFiles {
		files = append(files, *file.Filename)
		logrus.Infof("%s", *file.Filename)
	}
	return
}

// GetCommentFlag get CommentFlag from the GithubPrComment
func (c *GithubPrComment) GetCommentFlag() string {
	return c.CommentFlag
}

// GenCommentContent generate github comment content based on diff coverage and commentFlag
func GenCommentContent(commentPrefix string, delta cover.DeltaCovList) string {
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"File", "Base Coverage", "New Coverage", "Delta"})
	table.SetAutoFormatHeaders(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})
	for _, d := range delta {
		table.Append([]string{fmt.Sprintf("[%s](%s)", d.FileName, d.LineCovLink), d.BasePer, d.NewPer, d.DeltaPer})
	}
	table.Render()

	content := []string{
		commentPrefix,
		fmt.Sprintf("Say `/test %s` to re-run this coverage report", os.Getenv("JOB_NAME")),
		buf.String(),
	}

	return strings.Join(content, "\n")
}
