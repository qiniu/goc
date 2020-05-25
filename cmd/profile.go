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
	"log"
	"os"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get coverage profile from service registry center",
	Run: func(cmd *cobra.Command, args []string) {
		res, err := cover.NewWorker().Profile(center)
		if err != nil {
			log.Fatalf("call host %v failed, err: %v, response: %v", center, err, string(res))
		}

		if output == "" {
			fmt.Fprint(os.Stdout, string(res))
		} else {
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

var output string

func init() {
	profileCmd.Flags().StringVarP(&output, "output", "o", "", "download cover profile")
	profileCmd.Flags().StringVarP(&center, "center", "", "http://127.0.0.1:7777", "cover profile host center")
	rootCmd.AddCommand(profileCmd)
}
