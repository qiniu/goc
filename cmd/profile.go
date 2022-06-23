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
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/qiniu/goc/pkg/cover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get coverage profile from service registry center",
	Long:  `Get code coverage profile for the services under test at runtime.`,
	Example: `
# Get coverage counter from default register center http://127.0.0.1:7777, the result output to stdout.
goc profile

# Get coverage counter from specified register center, the result output to specified file.
goc profile --center=http://192.168.1.1:8080 --output=./coverage.cov

# Get coverage counter of several specified services. You can get all available service names from command 'goc list'. Use 'service' and 'address' flag at the same time may cause ambiguity, please use them separately.
goc profile --service=service1,service2,service3

# Get coverage counter of several specified addresses. You can get all available addresses from command 'goc list'. Use 'service' and 'address' flag at the same time may cause ambiguity, please use them separately.
goc profile --address=address1,address2,address3

# Only get the coverage data of files matching the special patterns
goc profile --coverfile=pattern1,pattern2,pattern3

# Force fetching all available profiles.
goc profile --force
`,
	Run: func(cmd *cobra.Command, args []string) {
		p := cover.ProfileParam{
			Force:             force,
			Service:           svrList,
			Address:           addrList,
			CoverFilePatterns: coverFilePatterns,
			SkipFilePatterns:  skipFilePatterns,
		}
		res, err := cover.NewWorker(center).Profile(p)
		if err != nil {
			log.Fatalf("Goc server %v return an error: %v", center, err)
		}

		if output == "" {
			fmt.Fprint(os.Stdout, string(res))
		} else {
			var dir, filename string = path.Split(output)
			if dir != "" {
				err = os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					log.Fatalf("failed to create directory %s, err:%v", dir, err)
				}
			}
			if filename == "" {
				output += "coverage.cov"
			}

			f, err := os.Create(output)
			if err != nil {
				log.Fatalf("failed to create file %s, err:%v", output, err)
			}
			defer f.Close()
			_, err = io.Copy(f, bytes.NewReader(res))
			if err != nil {
				log.Fatalf("failed to write file: %v, err: %v", output, err)
			}
		}
	},
}

var (
	svrList           []string // --service flag
	addrList          []string // --address flag
	force             bool     // --force flag
	output            string   // --output flag
	coverFilePatterns []string // --coverfile flag
	skipFilePatterns  []string // --skipfile flag
)

func init() {
	profileCmd.Flags().StringVarP(&output, "output", "o", "", "download cover profile")
	profileCmd.Flags().StringSliceVarP(&svrList, "service", "", nil, "service name to fetch profile, see 'goc list' for all services.")
	profileCmd.Flags().StringSliceVarP(&addrList, "address", "", nil, "address to fetch profile, see 'goc list' for all addresses.")
	profileCmd.Flags().BoolVarP(&force, "force", "f", true, "force fetching all available profiles")
	profileCmd.Flags().StringSliceVarP(&coverFilePatterns, "coverfile", "", nil, "only output coverage data of the files matching the patterns")
	profileCmd.Flags().StringSliceVarP(&skipFilePatterns, "skipfile", "", nil, "skip the files matching the patterns when outputing coverage data")
	addBasicFlags(profileCmd.Flags())
	rootCmd.AddCommand(profileCmd)
}
