package cmd

import (
	"github.com/qiniu/goc/v2/pkg/log"
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
		runMerge(args, outputMergeProfile)
	},
}

var outputMergeProfile string

func init() {
	mergeCmd.Flags().StringVarP(&outputMergeProfile, "output", "o", "mergeprofile.cov", "output file")

	rootCmd.AddCommand(mergeCmd)
}

func runMerge(args []string, output string) {
	if len(args) == 0 {
		log.Fatalf("Expected at least one coverage file.")
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

	err = util.DumpProfile(output, merged)
	if err != nil {
		log.Fatalf("fail to dump the merged file: %v", err)
	}
}
