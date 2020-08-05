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
	"log"

	"github.com/spf13/cobra"
	"golang.org/x/tools/cover"
	"k8s.io/test-infra/gopherage/pkg/cov"
	"k8s.io/test-infra/gopherage/pkg/util"
)

var mergeCmd = &cobra.Command{
	Use:   "merge [files...]",
	Short: "Merge multiple coherent Go coverage files into a single file.",
	Long: `merge will merge multiple Go coverage files into a single coverage file.
merge requires that the files are 'coherent', meaning that if they both contain references to the
same paths, then the contents of those source files were identical for the binary that generated
each file.
`,
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd, args)
	},
}

var outputMergeProfile string

func init() {
	mergeCmd.Flags().StringVarP(&outputMergeProfile, "output", "o", "mergeprofile.cov", "output file")

	rootCmd.AddCommand(mergeCmd)
}

func run(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatalln("Expected at least one coverage file.")
	}

	profiles := make([][]*cover.Profile, len(args))
	for _, path := range args {
		profile, err := util.LoadProfile(path)
		if err != nil {
			log.Fatalf("failed to open %s: %v", path, err)
		}
		profiles = append(profiles, profile)
	}

	merged, err := cov.MergeMultipleProfiles(profiles)
	if err != nil {
		log.Fatalf("failed to merge files: %v", err)
	}

	err = util.DumpProfile(outputMergeProfile, merged)
	if err != nil {
		log.Fatalln(err)
	}
}
