/*
 Copyright 2021 Qiniu Cloud (qiniu.com)
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
	"github.com/qiniu/goc/v2/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get coverage profile from service registry center",
	Long:  `Get code coverage profile for the services under test at runtime.`,
	//Run: profile,
}

var (
	profileHost     string
	profileOutput   string // --output flag
	profileIds      []string
	profilePackages string
	profileExtra    string
)

func init() {

	add1Flags := func(f *pflag.FlagSet) {
		f.StringVar(&profileHost, "host", "127.0.0.1:7777", "specify the host of the goc server")
		f.StringSliceVar(&profileIds, "id", nil, "specify the ids of the services")
		f.StringVar(&profileExtra, "extra", "", "specify the regex expression of extra, only profile with extra information will be downloaded")
	}

	add2Flags := func(f *pflag.FlagSet) {
		f.StringVarP(&profileOutput, "output", "o", "", "download cover profile")
		f.StringVar(&profilePackages, "packages", "", "specify the regex expression of packages, only profile of these packages will be downloaded")
	}

	add1Flags(getProfileCmd.Flags())
	add2Flags(getProfileCmd.Flags())

	add1Flags(clearProfileCmd.Flags())

	profileCmd.AddCommand(getProfileCmd)
	profileCmd.AddCommand(clearProfileCmd)
	rootCmd.AddCommand(profileCmd)
}

var getProfileCmd = &cobra.Command{
	Use: "get",
	Run: getProfile,
}

func getProfile(cmd *cobra.Command, args []string) {
	client.GetProfile(profileHost, profileIds, profilePackages, profileExtra, profileOutput)
}

var clearProfileCmd = &cobra.Command{
	Use: "clear",
	Run: clearProfile,
}

func clearProfile(cmd *cobra.Command, args []string) {
	client.ClearProfile(profileHost, profileIds, profileExtra)
}
