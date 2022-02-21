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

package build

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/qiniu/goc/v2/pkg/log"
	"github.com/tongjingran/copy"
	"golang.org/x/mod/modfile"
)

// copyProjectToTmp copies project files to the temporary directory
//
// It will ignore .git and irregular files, only copy source(text) files
func (b *Build) copyProjectToTmp() {
	curProject := b.CurModProjectDir
	tmpProject := b.TmpModProjectDir

	if _, err := os.Stat(tmpProject); !os.IsNotExist(err) {
		log.Infof("find previous temporary directory, delete")
		err := os.RemoveAll(tmpProject)
		if err != nil {
			log.Fatalf("fail to remove preivous temporary directory: %v", err)
		}
	}

	log.StartWait("coping project")
	err := os.MkdirAll(tmpProject, os.ModePerm)
	if err != nil {
		log.Fatalf("fail to create temporary directory: %v", err)
	}

	// copy
	if err := copy.Copy(curProject, tmpProject, copy.Options{Skip: skipCopy}); err != nil {
		log.Fatalf("fail to copy the folder from %v to %v, the err: %v", curProject, tmpProject, err)
	}

	log.StopWait()
}

// TmpFolderName generates a directory name according to the path
func TmpFolderName(path string) string {
	sum := sha256.Sum256([]byte(path))
	h := fmt.Sprintf("%x", sum[:6])

	return "gocbuild" + h
}

// skipCopy skip copy .git dir and irregular files
func skipCopy(src string, info os.FileInfo) (bool, error) {
	irregularModeType := os.ModeNamedPipe | os.ModeSocket | os.ModeDevice | os.ModeCharDevice | os.ModeIrregular
	if strings.HasSuffix(src, "/.git") {
		log.Debugf("skip .git dir [%s]", src)
		return true, nil
	}
	if info.Mode()&irregularModeType != 0 {
		log.Debugf("skip file [%s], the file mode is [%s]", src, info.Mode().String())
		return true, nil
	}
	return false, nil
}

// clean clears the temporary project
func (b *Build) clean() {
	if !b.Debug {
		if err := os.RemoveAll(b.TmpModProjectDir); err != nil {
			log.Fatalf("fail to delete the temporary project: %v", err)
		}
		log.Donef("delete the temporary project")
	} else {
		log.Debugf("--debug is enabled, keep the temporary project")
	}
}

// updateGoModFile rewrites the go.mod file in the temporary directory,
//
// if it has a 'replace' directive, and the directive has a relative local path,
// it will be rewritten with a absolute path.
//
// ex.
//
// suppose original project is located at /path/to/aa/bb/cc, go.mod contains a directive:
// 'replace github.com/qiniu/bar => ../home/foo/bar'
//
// after the project is copied to temporary directory, it should be rewritten as
// 'replace github.com/qiniu/bar => /path/to/aa/bb/home/foo/bar'
func (b *Build) updateGoModFile() (updateFlag bool, newModFile []byte) {
	tempModfile := filepath.Join(b.TmpModProjectDir, "go.mod")
	buf, err := ioutil.ReadFile(tempModfile)
	if err != nil {
		log.Fatalf("cannot find go.mod file in temporary directory: %v", err)
	}
	oriGoModFile, err := modfile.Parse(tempModfile, buf, nil)
	if err != nil {
		log.Fatalf("cannot parse go.mod: %v", err)
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
			fullPath := filepath.Join(b.CurModProjectDir, newPath)
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

	if updateFlag {
		log.Infof("go.mod needs rewrite")
		err := os.WriteFile(tempModfile, newModFile, os.ModePerm)
		if err != nil {
			log.Fatalf("fail to update go.mod: %v", err)
		}
		b.IsModEdit = true
	}
	return
}
