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

package client

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/qiniu/goc/v2/pkg/client/rest"
	"github.com/qiniu/goc/v2/pkg/client/rest/profile"
	"github.com/qiniu/goc/v2/pkg/log"
)

func GetProfile(host string, ids []string, packages string, extra string, output string) {
	gocClient := rest.NewV2Client(host)

	profiles, err := gocClient.Profile().Get(ids,
		profile.WithPackagePattern(packages),
		profile.WithExtraPattern(extra))
	if err != nil {
		log.Fatalf("fail to get profile from the goc server: %v, response: %v", err, profiles)
	}

	if output == "" {
		fmt.Fprint(os.Stdout, profiles)
	} else {
		var dir, filename string = filepath.Split(output)
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
		_, err = io.Copy(f, bytes.NewReader([]byte(profiles)))
		if err != nil {
			log.Fatalf("failed to write file: %v, err: %v", output, err)
		}
	}
}

func ClearProfile(host string, ids []string, extra string) {
	gocClient := rest.NewV2Client(host)

	err := gocClient.Profile().Delete(ids,
		profile.WithExtraPattern(extra))

	if err != nil {
		log.Fatalf("fail to clear the profile: %v", err)
	}
}
