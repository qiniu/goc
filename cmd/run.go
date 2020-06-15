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
	"io/ioutil"
	"net"

	"github.com/qiniu/goc/pkg/build"
	"github.com/qiniu/goc/pkg/cover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run covers and runs the named main Go package",
	Long: `Run covers and runs the named main Go package, 
It is exactly behave as 'go run .' in addition of some internal goc features.`,
	Example: `
goc run .
`,
	Run: func(cmd *cobra.Command, args []string) {
		gocBuild := build.NewBuild(buildFlags, packages, buildOutput)
		gocBuild.GoRunExecFlag = goRunExecFlag
		gocBuild.GoRunArguments = goRunArguments
		defer gocBuild.Clean()

		// only save services in memory
		cover.DefaultStore = cover.NewMemoryStore()

		// start goc server
		var l = newLocalListener()
		go cover.GocServer(ioutil.Discard).RunListener(l)
		gocServer := fmt.Sprintf("http://%s", l.Addr().String())
		fmt.Printf("[goc] goc server started: %s \n", gocServer)

		// execute covers for the target source with original buildFlags and new GOPATH( tmp:original )
		cover.Execute(buildFlags, gocBuild.NewGOPATH, gocBuild.TmpDir, mode, gocServer)

		gocBuild.Run()
	},
}

func init() {
	addRunFlags(runCmd.Flags())
	rootCmd.AddCommand(runCmd)
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			log.Fatalf("failed to listen on a port: %v", err)
		}
	}
	return l
}
