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

package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/qiniu/goc/pkg/github"
	"github.com/qiniu/goc/pkg/prow"
	"github.com/qiniu/goc/pkg/qiniu"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "do coverage profile diff analysis, it can also work with prow and post comments to github pull request if needed",
	Example: `	# Diff two local coverage profile and display
	goc diff --new-profile=<xxxx> --base-profile=<xxxx> 

	# Diff local coverage profile with the remote one in prow job using default qiniu-credential
	goc diff --prow-postsubmit-job=<xxx> --new-profile=<xxx> 

	# Calculate and display full diff coverage between new-profile and base-profile, not concerned github changed files
	goc diff --prow-postsubmit-job=<xxx> --new-profile=<xxx> --full-diff=true 

	# Diff local coverage profile with the remote one in prow job
	goc diff --prow-postsubmit-job=<xxx> --prow-remote-profile-name=<xxx> 
    		 --qiniu-credential=<xxx> --new-profile=<xxxx> 

	# Diff coverage profile with the remote one in prow job, and post comments to github PR
	goc diff --prow-postsubmit-job=<xxx> --prow-profile=<xxx> 
    		 --github-token=<xxx> --github-user=<xxx> --github-comment-prefix=<xxx> 
    		 --qiniu-credential=<xxx> --coverage-threshold-percentage=<xxx> --new-profile=<xxxx> 
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if baseProfile != "" {
			doDiffForLocalProfiles(cmd, args)
		} else if prowPostSubmitJob != "" {
			doDiffUnderProw(cmd, args)
		} else {
			logrus.Fatalf("either base-profile or prow-postsubmit-job must be provided")
		}
	},
}

var (
	newProfile        string
	baseProfile       string
	coverageThreshold int

	prowPostSubmitJob string
	prowProfile       string

	githubToken         string
	githubUser          string
	githubCommentPrefix string

	qiniuCredential string

	robotName string
	fullDiff  string
)

func init() {
	diffCmd.Flags().StringVarP(&newProfile, "new-profile", "n", "", "local profile which works as the target to analysis")
	diffCmd.MarkFlagRequired("new-profile")
	diffCmd.Flags().StringVarP(&baseProfile, "base-profile", "b", "", "another local profile which works as baseline to compare with the target")
	diffCmd.Flags().IntVarP(&coverageThreshold, "coverage-threshold-percentage", "", 0, "coverage threshold percentage")
	diffCmd.Flags().StringVarP(&prowPostSubmitJob, "prow-postsubmit-job", "", "", "prow postsubmit job which used to find the base profile")
	diffCmd.Flags().StringVarP(&prowProfile, "prow-remote-profile-name", "", "filtered.cov", "the name of profile in prow postsubmit job, which used as the base profile to compare")
	diffCmd.Flags().StringVarP(&githubToken, "github-token", "", "/etc/github/oauth", "path to token to access github repo")
	diffCmd.Flags().StringVarP(&githubUser, "github-user", "", "", "github user name when comments in github")
	diffCmd.Flags().StringVarP(&githubCommentPrefix, "github-comment-prefix", "", "", "specific comment flag you provided")
	diffCmd.Flags().StringVarP(&qiniuCredential, "qiniu-credential", "", "/etc/qiniuconfig/qiniu.json", "path to credential file to access qiniu cloud")
	diffCmd.Flags().StringVarP(&robotName, "robot-name", "", "qiniu-bot", "github user name for coverage robot")
	diffCmd.Flags().StringVarP(&fullDiff, "full-diff", "", "", "empty means set false, not empty means set true. When set true, will calculate and display full diff coverage between new-profile and base-profile")

	rootCmd.AddCommand(diffCmd)
}

//goc diff --new-profile=./new.cov --base-profile=./base.cov
//+------------------------------------------------------+---------------+--------------+--------+
//|                         File                         | Base Coverage | New Coverage | Delta  |
//+------------------------------------------------------+---------------+--------------+--------+
//| qiniu.com/kodo/bd/pfd/pfdstg/cursor/mgr.go           |     53.5%     |    50.5%     | -3.0%  |
//| qiniu.com/kodo/bd/pfd/pfdstg/svr/getstripe.go        |     0.5%      |     0.0%     | -0.5%  |
//| Total                                                |     35.7%     |    35.7%     | -0.0%  |
//+------------------------------------------------------+---------------+--------------+--------+
func doDiffForLocalProfiles(cmd *cobra.Command, args []string) {
	localP, err := cover.ReadFileToCoverList(newProfile)
	if err != nil {
		logrus.Fatal(err)
	}

	baseP, err := cover.ReadFileToCoverList(baseProfile)
	if err != nil {
		logrus.Fatal(err)
	}

	//calculate diff file cov and display
	rows := cover.GetDeltaCov(localP, baseP)
	rows.Sort()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"File", "Base Coverage", "New Coverage", "Delta"})
	table.SetAutoFormatHeaders(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})
	for _, row := range rows {
		table.Append([]string{row.FileName, row.BasePer, row.NewPer, row.DeltaPer})
	}
	totalDelta := cover.PercentStr(cover.TotalDelta(localP, baseP))
	table.Append([]string{"Total", baseP.TotalPercentage(), localP.TotalPercentage(), totalDelta})
	table.Render()
}

func doDiffUnderProw(cmd *cobra.Command, args []string) {
	var (
		prNumStr  = os.Getenv("PULL_NUMBER")
		pullSha   = os.Getenv("PULL_PULL_SHA")
		baseSha   = os.Getenv("PULL_BASE_SHA")
		repoOwner = os.Getenv("REPO_OWNER")
		repoName  = os.Getenv("REPO_NAME")
		jobType   = os.Getenv("JOB_TYPE")
		jobName   = os.Getenv("JOB_NAME")
		buildStr  = os.Getenv("BUILD_NUMBER")
		artifacts = os.Getenv("ARTIFACTS")
		fd        bool
	)
	logrus.Printf("Running coverage for PR = %s; PR commit SHA = %s;base SHA = %s", prNumStr, pullSha, baseSha)

	switch jobType {
	case "periodic":
		logrus.Printf("job type %s, do nothing", jobType)
	case "postsubmit":
		logrus.Printf("job type %s, do nothing", jobType)
	case "presubmit":
		if githubToken == "" {
			logrus.Fatalf("github token not provided")
		}
		prClient := github.NewPrClient(githubToken, repoOwner, repoName, prNumStr, robotName, githubCommentPrefix)

		if qiniuCredential == "" {
			logrus.Fatalf("qiniu credential not provided")
		}
		var qc *qiniu.Client
		var conf qiniu.Config
		files, err := ioutil.ReadFile(*&qiniuCredential)
		if err != nil {
			logrus.WithError(err).Fatal("Error reading qiniu config file")
		}
		if err := json.Unmarshal(files, &conf); err != nil {
			logrus.Fatal("Error unmarshal qiniu config file")
		}
		if conf.Bucket == "" {
			logrus.Fatal("no qiniu bucket provided")
		}
		if conf.AccessKey == "" || conf.SecretKey == "" {
			logrus.Fatal("either qiniu access key or secret key was not provided")
		}
		if conf.Domain == "" {
			logrus.Fatal("no qiniu bucket domain was provided")
		}
		qc = qiniu.NewClient(&conf)

		localArtifacts := qiniu.Artifacts{
			Directory:          artifacts,
			ProfileName:        newProfile,
			ChangedProfileName: qiniu.ChangedProfileName,
		}

		if fullDiff != "" {
			fd = true
		}

		job := prow.Job{
			JobName:                jobName,
			BuildId:                buildStr,
			Org:                    repoOwner,
			RepoName:               repoName,
			PRNumStr:               prNumStr,
			PostSubmitJob:          prowPostSubmitJob,
			LocalProfilePath:       newProfile,
			PostSubmitCoverProfile: prowProfile,
			QiniuClient:            qc,
			LocalArtifacts:         &localArtifacts,
			GithubComment:          prClient,
			FullDiff:               fd,
		}
		if err := job.RunPresubmit(); err != nil {
			logrus.Fatalf("run presubmit job failed, err: %v", err)
		}
	default:
		logrus.Printf("Unknown job type: %s, do nothing.", jobType)
	}
}
