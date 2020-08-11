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

package build

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
)

func (b *Build) cpGoModulesProject() {
	for _, v := range b.Pkgs {
		if v.Name == "main" {
			dst := b.TmpDir
			src := v.Module.Dir

			if err := copy.Copy(src, dst); err != nil {
				log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
			}
			break
		} else {
			continue
		}
	}
}

func (b *Build) updateGoModFile() (string, error) {
	tempModfile := filepath.Join(b.TmpDir, "go.mod")
	buf, err := ioutil.ReadFile(tempModfile)
	if err != nil {
		return "", err
	}
	oriGoModFile, err := modfile.Parse(tempModfile, buf, nil)
	if err != nil {
		return "", err
	}

	for index := range oriGoModFile.Replace {
		replace := oriGoModFile.Replace[index]
		oldPath := replace.Old.Path
		oldVersion := replace.Old.Version
		newPath := replace.New.Path
		newVersion := replace.New.Version
		if strings.HasPrefix(replace.New.Path, "..") {
			fullPath := filepath.Join(b.ModRoot, newPath)
			absPath, err := filepath.Abs(fullPath)
			if err != nil {
				return "", err
			}
			err = oriGoModFile.DropReplace(oldPath, oldVersion)
			err = oriGoModFile.AddReplace(oldPath, oldVersion, absPath, newVersion)
			if err != nil {
				return "", err
			}
		}
	}
	oriGoModFile.Cleanup()
	newGoModFile, err := oriGoModFile.Format()
	if err != nil {
		return "", err
	}
	return string(newGoModFile), nil
}
