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
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
)

var gitInfoCmd = &cobra.Command{
	Use:   "gitinfo",
	Short: "Get git info from service registry center",
	Long:  `Get git info from service registry center to distinguish coverage data by commitid & branch under test at runtime.`,
	Example: `
		./goc gitinfo --debug  --service awesomeProject  --path .
        ./goc gitinfo --debug  --address http://127.0.0.1:9090 --path .

		result: {"CommitID":"d3fb54fcee7dadcc3d499562a0ff6ebf5d1c5323\n","Branch":"test\n"}%    
`,
	Run: func(cmd *cobra.Command, args []string) {
		p := cover.GitInfoParam{
			Path:  Path,
			Service: svrList,
			Address: addrList,
		}
		res, err := cover.NewWorker(center).GitInfo(p)
		if err != nil {
			log.Fatalf("call host %v failed, err: %v, response: %v", center, err, string(res))
		}
		fmt.Fprint(os.Stdout, string(res))
	},
}

var (
	Path          string // --service flag
)
func init() {
	addBasicFlags(gitInfoCmd.Flags())
	gitInfoCmd.Flags().StringVarP(&Path,"path","",".","work space path of server, default .")
	gitInfoCmd.Flags().StringSliceVarP(&svrList, "service", "", nil, "service name to clear profile, see 'goc list' for all services.")
	gitInfoCmd.Flags().StringSliceVarP(&addrList, "address", "", nil, "address to clear profile, see 'goc list' for all addresses.")
	rootCmd.AddCommand(gitInfoCmd)
}
