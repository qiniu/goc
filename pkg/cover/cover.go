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
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

// TestCover is a collection of all counters
type TestCover struct {
	Mode         string
	Center       string // cover profile host center
	MainPkgCover *PackageCover
	DepsCover    []*PackageCover
	CacheCover   map[string]*PackageCover
}

// PackageCover holds all the generate coverage variables of a package
type PackageCover struct {
	Package *Package
	Vars    map[string]*FileVar
}

// FileVar holds the name of the generated coverage variables targeting the named file.
type FileVar struct {
	File string
	Var  string
}

// Package map a package output by go list
// this is subset of pakcage struct in: https://github.com/golang/go/blob/master/src/cmd/go/internal/load/pkg.go#L58
type Package struct {
	Dir        string `json:"Dir"`        // directory containing package sources
	ImportPath string `json:"ImportPath"` // import path of package in dir
	Name       string `json:"Name"`       // package name
	Target     string `json:",omitempty"` // installed target for this package (may be executable)
	Root       string `json:",omitempty"` // Go root, Go path dir, or module root dir containing this package

	Module   *ModulePublic `json:",omitempty"`         // info about package's module, if any
	Goroot   bool          `json:"Goroot,omitempty"`   // is this package in the Go root?
	Standard bool          `json:"Standard,omitempty"` // is this package part of the standard Go library?
	DepOnly  bool          `json:"DepOnly,omitempty"`  // package is only a dependency, not explicitly listed

	// Source files
	GoFiles  []string `json:"GoFiles,omitempty"`  // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles []string `json:"CgoFiles,omitempty"` // .go source files that import "C"

	// Dependency information
	Deps      []string          `json:"Deps,omitempty"` // all (recursively) imported dependencies
	Imports   []string          `json:",omitempty"`     // import paths used by this package
	ImportMap map[string]string `json:",omitempty"`     // map from source import to ImportPath (identity entries omitted)

	// Error information
	Incomplete bool            `json:"Incomplete,omitempty"` // this package or a dependency has an error
	Error      *PackageError   `json:"Error,omitempty"`      // error loading package
	DepsErrors []*PackageError `json:"DepsErrors,omitempty"` // errors loading dependencies
}

type ModulePublic struct {
	Path      string        `json:",omitempty"` // module path
	Version   string        `json:",omitempty"` // module version
	Versions  []string      `json:",omitempty"` // available module versions
	Replace   *ModulePublic `json:",omitempty"` // replaced by this module
	Time      *time.Time    `json:",omitempty"` // time version was created
	Update    *ModulePublic `json:",omitempty"` // available update (with -u)
	Main      bool          `json:",omitempty"` // is this the main module?
	Indirect  bool          `json:",omitempty"` // module is only indirectly needed by main module
	Dir       string        `json:",omitempty"` // directory holding local copy of files, if any
	GoMod     string        `json:",omitempty"` // path to go.mod file describing module, if any
	GoVersion string        `json:",omitempty"` // go version used in module
	Error     *ModuleError  `json:",omitempty"` // error loading module
}

type ModuleError struct {
	Err string // error text
}

// PackageError is the error info for a package when list failed
type PackageError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error (if present, file:line:col)
	Err         string   // the error itself
}

// ListPackages list all packages under specific via go list command
func ListPackages(dir string, args []string, newgopath string) map[string]*Package {
	cmd := exec.Command("go", args...)
	log.Printf("go list cmd is: %v", cmd.Args)
	cmd.Dir = dir
	if newgopath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", newgopath))
	}
	out, _ := cmd.Output()
	// if err != nil {
	//	 log.Fatalf("excute `go list -json ./...` command failed, err: %v, out: %v", err, string(out))
	// }

	dec := json.NewDecoder(bytes.NewReader(out))
	pkgs := make(map[string]*Package, 0)
	for {
		var pkg Package
		if err := dec.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("reading go list output: %v", err)
		}
		if pkg.Error != nil {
			log.Fatalf("list package %s failed with output: %v", pkg.ImportPath, pkg.Error)
		}

		// for _, err := range pkg.DepsErrors {
		// 	log.Fatalf("dependency package list failed, err: %v", err)
		// }

		pkgs[pkg.ImportPath] = &pkg
	}
	return pkgs
}

// AddCounters add counters for all go files under the package
func AddCounters(pkg *Package, newgopath string) (*PackageCover, error) {
	coverVarMap := declareCoverVars(pkg)

	// to construct: go tool cover -mode=atomic -o dest src (note: dest==src)
	var args = []string{"tool", "cover", "-mode=atomic"}
	for file, coverVar := range coverVarMap {
		var newArgs = args
		newArgs = append(newArgs, "-var", coverVar.Var)
		longPath := path.Join(pkg.Dir, file)
		newArgs = append(newArgs, "-o", longPath, longPath)
		cmd := exec.Command("go", newArgs...)
		if newgopath != "" {
			cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", newgopath))
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("execuate go tool cover -mode=atomic -var %s -o %s %s failed, err: %v, out: %s", coverVar.Var, longPath, longPath, err, string(out))
		}
	}

	return &PackageCover{
		Package: pkg,
		Vars:    coverVarMap,
	}, nil
}

// declareCoverVars attaches the required cover variables names
// to the files, to be used when annotating the files.
func declareCoverVars(p *Package) map[string]*FileVar {
	coverVars := make(map[string]*FileVar)
	coverIndex := 0
	// We create the cover counters as new top-level variables in the package.
	// We need to avoid collisions with user variables (GoCover_0 is unlikely but still)
	// and more importantly with dot imports of other covered packages,
	// so we append 12 hex digits from the SHA-256 of the import path.
	// The point is only to avoid accidents, not to defeat users determined to
	// break things.
	sum := sha256.Sum256([]byte(p.ImportPath))
	h := fmt.Sprintf("%x", sum[:6])
	for _, file := range p.GoFiles {
		// These names appear in the cmd/cover HTML interface.
		var longFile = path.Join(p.ImportPath, file)
		coverVars[file] = &FileVar{
			File: longFile,
			Var:  fmt.Sprintf("GoCover_%d_%x", coverIndex, h),
		}
		coverIndex++
	}

	for _, file := range p.CgoFiles {
		// These names appear in the cmd/cover HTML interface.
		var longFile = path.Join(p.ImportPath, file)
		coverVars[file] = &FileVar{
			File: longFile,
			Var:  fmt.Sprintf("GoCover_%d_%x", coverIndex, h),
		}
		coverIndex++
	}

	return coverVars
}

func declareCacheVars(in *PackageCover) map[string]*FileVar {
	sum := sha256.Sum256([]byte(in.Package.ImportPath))
	h := fmt.Sprintf("%x", sum[:5])

	vars := make(map[string]*FileVar)
	coverIndex := 0
	for _, v := range in.Vars {
		cacheVar := fmt.Sprintf("GoCacheCover_%d_%x", coverIndex, h)
		vars[cacheVar] = v
		coverIndex++
	}
	return vars
}

func CacheInternalCover(in *PackageCover) *PackageCover {
	c := &PackageCover{}
	vars := declareCacheVars(in)
	c.Package = in.Package
	c.Vars = vars
	return c
}

func AddCacheCover(pkg *Package, in *PackageCover) *PackageCover {
	c := &PackageCover{}
	sum := sha256.Sum256([]byte(pkg.ImportPath))
	h := fmt.Sprintf("%x", sum[:6])
	goFile := fmt.Sprintf("cache_vars_auto_generated_%x.go", h)
	p := &Package{
		Dir:        fmt.Sprintf("%s/cache_%x", pkg.Dir, h),
		ImportPath: fmt.Sprintf("%s/cache_%x", pkg.ImportPath, h),
		Name:       fmt.Sprintf("cache_%x", h),
	}
	p.GoFiles = append(p.GoFiles, goFile)
	c.Package = p
	c.Vars = declareCacheVars(in)
	return c
}

// CoverageList is a collection and summary over multiple file Coverage objects
type CoverageList struct {
	*Coverage
	Groups          []Coverage
	ConcernedFiles  map[string]bool
	CovThresholdInt int
}

// Coverage stores test coverage summary data for one file
type Coverage struct {
	FileName      string
	NCoveredStmts int
	NAllStmts     int
	LineCovLink   string
}

type codeBlock struct {
	fileName      string // the file the code block is in
	numStatements int    // number of statements in the code block
	coverageCount int    // number of times the block is covered
}

func CovList(f io.Reader) (g *CoverageList, err error) {
	scanner := bufio.NewScanner(f)
	scanner.Scan() // discard first line
	g = NewCoverageList("", map[string]bool{}, 0)

	for scanner.Scan() {
		row := scanner.Text()
		blk, err := toBlock(row)
		if err != nil {
			return nil, err
		}
		blk.addToGroupCov(g)
	}
	return
}

// NewCoverageList constructs new (file) group Coverage
func NewCoverageList(name string, concernedFiles map[string]bool, covThresholdInt int) *CoverageList {
	return &CoverageList{
		Coverage:        newCoverage(name),
		Groups:          []Coverage{},
		ConcernedFiles:  concernedFiles,
		CovThresholdInt: covThresholdInt,
	}
}

func newCoverage(name string) *Coverage {
	return &Coverage{name, 0, 0, ""}
}

// convert a line in profile file to a codeBlock struct
func toBlock(line string) (res *codeBlock, err error) {
	slice := strings.Split(line, " ")
	if len(slice) != 3 {
		return nil, fmt.Errorf("the profile line %s is not expected", line)
	}
	blockName := slice[0]
	nStmts, _ := strconv.Atoi(slice[1])
	coverageCount, _ := strconv.Atoi(slice[2])
	return &codeBlock{
		fileName:      blockName[:strings.Index(blockName, ":")],
		numStatements: nStmts,
		coverageCount: coverageCount,
	}, nil
}

// add blk Coverage to file group Coverage
func (blk *codeBlock) addToGroupCov(g *CoverageList) {
	if g.size() == 0 || g.lastElement().Name() != blk.fileName {
		// when a new file name is processed
		coverage := newCoverage(blk.fileName)
		g.append(coverage)
	}
	cov := g.lastElement()
	cov.NAllStmts += blk.numStatements
	if blk.coverageCount > 0 {
		cov.NCoveredStmts += blk.numStatements
	}
}

func (g *CoverageList) size() int {
	return len(g.Groups)
}

func (g *CoverageList) lastElement() *Coverage {
	return &g.Groups[g.size()-1]
}

func (g *CoverageList) append(c *Coverage) {
	g.Groups = append(g.Groups, *c)
}

// Group returns the collection of file Coverage objects
func (g *CoverageList) Group() *[]Coverage {
	return &g.Groups
}

// Map returns maps the file name to its coverage for faster retrieval
// & membership check
func (g *CoverageList) Map() map[string]Coverage {
	m := make(map[string]Coverage)
	for _, c := range g.Groups {
		m[c.Name()] = c
	}
	return m
}

// Name returns the file name
func (c *Coverage) Name() string {
	return c.FileName
}

// Percentage returns the percentage of statements covered
func (c *Coverage) Percentage() string {
	ratio, err := c.Ratio()
	if err == nil {
		return PercentStr(ratio)
	}

	return "N/A"
}

func (c *Coverage) Ratio() (ratio float32, err error) {
	if c.NAllStmts == 0 {
		err = fmt.Errorf("[%s] has 0 statement", c.Name())
	} else {
		ratio = float32(c.NCoveredStmts) / float32(c.NAllStmts)
	}
	return
}

// PercentStr converts a fraction number to percentage string representation
func PercentStr(f float32) string {
	return fmt.Sprintf("%.1f%%", f*100)
}
