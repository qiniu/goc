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

	log "github.com/sirupsen/logrus"
	"github.com/tongjingran/copy"
	"golang.org/x/mod/modfile"
)

func (b *Build) cpGoModulesProject() {
	for _, v := range b.Pkgs {
		if v.Name == "main" {
			dst := b.TmpDir
			src := v.Module.Dir

			if err := copy.Copy(src, dst, copy.Options{Skip: skipCopy}); err != nil {
				log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
			}
			break
		} else {
			continue
		}
	}
}

// updateGoModFile rewrites the go.mod file in the temporary directory,
// if it has a 'replace' directive, and the directive has a relative local path
// it will be rewritten with a absolute path.
// ex.
// suppose original project is located at /path/to/aa/bb/cc, go.mod contains a directive:
// 'replace github.com/qiniu/bar => ../home/foo/bar'
// after the project is copied to temporary directory, it should be rewritten as
// 'replace github.com/qiniu/bar => /path/to/aa/bb/home/foo/bar'
func (b *Build) updateGoModFile() (updateFlag bool, newModFile []byte, err error) {
	tempModfile := filepath.Join(b.TmpDir, "go.mod")
	buf, err := ioutil.ReadFile(tempModfile)
	if err != nil {
		return
	}
	oriGoModFile, err := modfile.Parse(tempModfile, buf, nil)
	if err != nil {
		return
	}

	updateFlag = false
	for index := range oriGoModFile.Replace {
		replace := oriGoModFile.Replace[index]
		oldPath := replace.Old.Path
		oldVersion := replace.Old.Version
		newPath := replace.New.Path
		newVersion := replace.New.Version
		// replace to a local filesystem does not have a version
		// absolute path no need to rewrite
		if newVersion == "" && !filepath.IsAbs(newPath) {
			var absPath string
			fullPath := filepath.Join(b.ModRoot, newPath)
			absPath, _ = filepath.Abs(fullPath)
			// DropReplace & AddReplace will not return error
			// so no need to check the error
			_ = oriGoModFile.DropReplace(oldPath, oldVersion)
			_ = oriGoModFile.AddReplace(oldPath, oldVersion, absPath, newVersion)
			updateFlag = true
		}
	}
	oriGoModFile.Cleanup()
	// Format will not return error, so ignore the returned error
	// func (f *File) Format() ([]byte, error) {
	//     return Format(f.Syntax), nil
	// }
	newModFile, _ = oriGoModFile.Format()
	return
}
