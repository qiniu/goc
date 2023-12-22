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
	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var coverCmd = &cobra.Command{
	Use:   "cover",
	Short: "Do cover for the target source",
	Long:  `Do cover for the target source. You can select different cover mode (set, count, atomic), default: count`,
	Example: `
# Do cover for the current path, default center: http://127.0.0.1:7777,  default cover mode: count.
goc cover

# Do cover for the current path, default cover mode: count.
goc cover --center=http://127.0.0.1:7777

# Do cover for the target path,  cover mode: atomic.
goc cover --center=http://127.0.0.1:7777 --target=/path/to/target --mode=atomic
`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		runCover(target)
	},
}

func runCover(target string) {
	buildFlags := viper.GetString("buildflags")
	ci := &cover.CoverInfo{
		Args:           buildFlags,
		GoPath:         "",
		Target:         target,
		Mode:           coverMode.String(),
		AgentPort:      agentPort.String(),
		Center:         center,
		Singleton:      singleton,
		OneMainPackage: false,
		CoverModName:   "coverPackageMod",
	}
	_ = cover.Execute(ci)
}

func init() {
	coverCmd.Flags().StringVar(&target, "target", ".", "target folder to cover")
	addCommonFlags(coverCmd.Flags())
	rootCmd.AddCommand(coverCmd)
}
