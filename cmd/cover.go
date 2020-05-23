/*
 Copyright 2020 Qiniu Cloud (七牛云)

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
	"strings"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
)

var coverCmd = &cobra.Command{
	Use:   "cover",
	Short: "do cover for the target source ",
	Run: func(cmd *cobra.Command, args []string) {
		doCover(cmd, args, "", "")
	},
}

var (
	target string
	center string
)

func init() {
	coverCmd.Flags().StringVarP(&center, "center", "", "http://127.0.0.1:7777", "cover profile host center")
	coverCmd.Flags().StringVarP(&target, "target", "", ".", "target folder to cover")

	rootCmd.AddCommand(coverCmd)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func doCover(cmd *cobra.Command, args []string, newgopath string, newtarget string) {
	if newtarget != "" {
		target = newtarget
	}
	if !isDirExist(target) {
		log.Fatalf("target directory %s not exist", target)
	}

	listArgs := []string{"list", "-json"}
	if len(args) != 0 {
		listArgs = append(listArgs, args...)
	}
	listArgs = append(listArgs, "./...")
	pkgs := cover.ListPackages(target, listArgs, newgopath)

	var seen = make(map[string]*cover.PackageCover)
	var seenCache = make(map[string]*cover.PackageCover)
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			log.Printf("handle package: %v", pkg.ImportPath)
			// inject the main package
			mainCover, err := cover.AddCounters(pkg, newgopath)
			if err != nil {
				log.Fatalf("failed to add counters for pkg %s, err: %v", pkg.ImportPath, err)
			}

			// new a testcover for this service
			tc := cover.TestCover{
				Mode:         "atomic",
				Center:       center,
				MainPkgCover: mainCover,
			}

			// handle its dependency
			var internalPkgCache = make(map[string][]*cover.PackageCover)
			tc.CacheCover = make(map[string]*cover.PackageCover)
			for _, dep := range pkg.Deps {

				if packageCover, ok := seen[dep]; ok {
					tc.DepsCover = append(tc.DepsCover, packageCover)
					continue
				}

				//only focus package neither standard Go library nor dependency library
				if depPkg, ok := pkgs[dep]; ok {

					if findInternal(dep) {

						//scan exist cache cover to tc.CacheCover
						if cache, ok := seenCache[dep]; ok {
							log.Printf("cache cover exist: %s", cache.Package.ImportPath)
							tc.CacheCover[cache.Package.Dir] = cache
							continue
						}

						// add counter for internal package
						inPkgCover, err := cover.AddCounters(depPkg, newgopath)
						if err != nil {
							log.Fatalf("failed to add counters for internal pkg %s, err: %v", depPkg.ImportPath, err)
						}
						parentDir := getInternalParent(depPkg.Dir)
						parentImportPath := getInternalParent(depPkg.ImportPath)

						//if internal parent dir or import is root path, ignore the dep. the dep is Go library nor dependency library
						if parentDir == "" {
							continue
						}
						if parentImportPath == "" {
							continue
						}

						pkg := &cover.Package{
							ImportPath: parentImportPath,
							Dir:        parentDir,
						}

						// Some internal package have same parent dir or import path
						// Cache all vars by internal parent dir for all child internal counter vars
						cacheCover := cover.AddCacheCover(pkg, inPkgCover)
						if v, ok := tc.CacheCover[cacheCover.Package.Dir]; ok {
							for cVar, val := range v.Vars {
								cacheCover.Vars[cVar] = val
							}
							tc.CacheCover[cacheCover.Package.Dir] = cacheCover
						} else {
							tc.CacheCover[cacheCover.Package.Dir] = cacheCover
						}

						// Cache all internal vars to internal parent package
						inCover := cover.CacheInternalCover(inPkgCover)
						if v, ok := internalPkgCache[cacheCover.Package.Dir]; ok {
							v = append(v, inCover)
							internalPkgCache[cacheCover.Package.Dir] = v
						} else {
							var covers []*cover.PackageCover
							covers = append(covers, inCover)
							internalPkgCache[cacheCover.Package.Dir] = covers
						}
						seenCache[dep] = cacheCover
						continue
					}

					packageCover, err := cover.AddCounters(depPkg, newgopath)
					if err != nil {
						log.Fatalf("failed to add counters for pkg %s, err: %v", depPkg.ImportPath, err)
					}
					tc.DepsCover = append(tc.DepsCover, packageCover)
					seen[dep] = packageCover
				}
			}

			if errs := cover.InjectCacheCounters(internalPkgCache, tc.CacheCover); len(errs) > 0 {
				log.Fatalf("failed to inject cache counters for package: %s, err: %v", pkg.ImportPath, errs)
			}

			// inject Http Cover APIs
			var httpCoverApis = fmt.Sprintf("%s/http_cover_apis_auto_generated.go", pkg.Dir)
			if err := cover.InjectCountersHandlers(tc, httpCoverApis); err != nil {
				log.Fatalf("failed to inject counters for package: %s, err: %v", pkg.ImportPath, err)
			}
		}
	}
}

func isDirExist(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// Refer: https://github.com/golang/go/blob/master/src/cmd/go/internal/load/pkg.go#L1334:6
// findInternal looks for the final "internal" path element in the given import path.
// If there isn't one, findInternal returns ok=false.
// Otherwise, findInternal returns ok=true and the index of the "internal".
func findInternal(path string) bool {
	// Three cases, depending on internal at start/end of string or not.
	// The order matters: we must return the index of the final element,
	// because the final one produces the most restrictive requirement
	// on the importer.
	switch {
	case strings.HasSuffix(path, "/internal"):
		return true
	case strings.Contains(path, "/internal/"):
		return true
	case path == "internal", strings.HasPrefix(path, "internal/"):
		return true
	}
	return false
}

func getInternalParent(path string) string {
	switch {
	case strings.HasSuffix(path, "/internal"):
		return strings.Split(path, "/internal")[0]
	case strings.Contains(path, "/internal/"):
		return strings.Split(path, "/internal/")[0]
	case path == "internal":
		return ""
	case strings.HasPrefix(path, "internal/"):
		return strings.Split(path, "internal/")[0]
	}
	return ""
}
