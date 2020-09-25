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

	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a service into service center",
	Long:  "Register a service into service center",
	Example: `
goc register [flags] 
`,
	Run: func(cmd *cobra.Command, args []string) {
		s := cover.ServiceUnderTest{
			Name:    name,
			Address: address,
		}
		res, err := cover.NewWorker(center).RegisterService(s)
		if err != nil {
			log.Fatalf("register service failed, err: %v", err)
		}
		fmt.Fprint(os.Stdout, string(res))
	},
}

var (
	name    string
	address string
)

func init() {
	registerCmd.Flags().StringVarP(&center, "center", "", "http://127.0.0.1:7777", "cover profile host center")
	registerCmd.Flags().StringVarP(&name, "name", "n", "", "service name")
	registerCmd.Flags().StringVarP(&address, "address", "a", "", "service address")
	registerCmd.MarkFlagRequired("name")
	registerCmd.MarkFlagRequired("address")
	rootCmd.AddCommand(registerCmd)
}
