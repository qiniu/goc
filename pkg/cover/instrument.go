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

package cover

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

//go:embed templates/coverMain.go.tmpl
var coverMain string

//go:embed templates/coverParentFile.go.tmpl
var coverParentFile string

//go:embed templates/coverParentVars.go.tmpl
var coverParentVars string

// InjectCountersHandlers generate a file _cover_http_apis.go besides the main.go file
func InjectCountersHandlers(tc TestCover, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	if err := coverMainTmpl.Execute(f, tc); err != nil {
		return err
	}
	return nil
}

var coverMainTmpl = template.Must(template.New("coverMain").Parse(coverMain))

var coverParentFileTmpl = template.Must(template.New("coverParentFileTmpl").Parse(coverParentFile))

var coverParentVarsTmpl = template.Must(template.New("coverParentVarsTmpl").Parse(coverParentVars))

func InjectCacheCounters(covers map[string][]*PackageCover, cache map[string]*PackageCover) []error {
	var errs []error
	for k, v := range covers {
		if pkg, ok := cache[k]; ok {
			err := checkCacheDir(pkg.Package.Dir)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			_, pkgName := path.Split(k)
			err = injectCache(v, pkgName, fmt.Sprintf("%s/%s", pkg.Package.Dir, pkg.Package.GoFiles[0]))
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
	}
	return errs
}

// InjectCacheCounters generate a file _cover_http_apis.go besides the main.go file
func injectCache(covers []*PackageCover, pkg, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}

	if err := coverParentFileTmpl.Execute(f, pkg); err != nil {
		return err
	}

	if err := coverParentVarsTmpl.Execute(f, covers); err != nil {
		return err
	}
	return nil
}

func checkCacheDir(p string) error {
	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		err := os.Mkdir(p, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func injectGlobalCoverVarFile(ci *CoverInfo, content string) error {
	coverFile, err := os.Create(filepath.Join(ci.Target, ci.GlobalCoverVarImportPath, "cover.go"))
	if err != nil {
		return err
	}
	defer coverFile.Close()

	packageName := "package " + filepath.Base(ci.GlobalCoverVarImportPath) + "\n\n"

	_, err = coverFile.WriteString(packageName)
	if err != nil {
		return err
	}
	_, err = coverFile.WriteString(content)

	return err
}
