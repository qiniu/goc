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
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/qiniu/goc/pkg/build"
	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Do cover for all go files and execute go install command",
	Long: `This install command is a little different from the official one, for instance:
* 'goc install -- ./...' is equal to 'goc cover && go install ./...'
* 'goc install --center=http://127.0.0.1:7777 -- -static ./...' is equal to 'goc cover --center=http://127.0.0.1:7777 && go install -static ./...'`,
	Run: func(cmd *cobra.Command, args []string) {
		newgopath, newwd, tmpdir, pkgs := build.MvProjectsToTmp(target, args)
		doCover(cmd, args, newgopath, tmpdir)
		doInstall(args, newgopath, newwd, pkgs)
	},
}

func init() {
	installCmd.Flags().StringVarP(&center, "center", "", "http://127.0.0.1:7777", "cover profile host center")

	rootCmd.AddCommand(installCmd)
}

func doInstall(args []string, newgopath string, newworkingdir string, pkgs map[string]*cover.Package) {
	log.Println("Go building in temp...")
	newArgs := []string{"install"}
	newArgs = append(newArgs, args...)
	cmd := exec.Command("go", newArgs...)
	cmd.Dir = newworkingdir

	// Change the temp GOBIN, to force binary install to original place
	cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%v", build.FindWhereToInstall(pkgs)))
	if newgopath != "" {
		// Change to temp GOPATH for go install command
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%v", newgopath))
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Fail to execute: go install %v. The error is: %v, the stdout/stderr is: %v", strings.Join(args, " "), err, string(out))
	}
	log.Printf("Go install successful. Binary installed in: %v", build.FindWhereToInstall(pkgs))
}
