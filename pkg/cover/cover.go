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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/qiniu/goc/pkg/cover/internal/tool"
	"github.com/sirupsen/logrus"
)

var (
	// ErrCoverPkgFailed represents the error that fails to inject the package
	ErrCoverPkgFailed = errors.New("fail to inject code to project")
	// ErrCoverListFailed represents the error that fails to list package dependencies
	ErrCoverListFailed = errors.New("fail to list package dependencies")
)

// TestCover is a collection of all counters
type TestCover struct {
	Mode                     string
	AgentPort                string
	Center                   string // cover profile host center
	MainPkgCover             *PackageCover
	DepsCover                []*PackageCover
	CacheCover               map[string]*PackageCover
	GlobalCoverVarImportPath string
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
// this is subset of package struct in: https://github.com/golang/go/blob/master/src/cmd/go/internal/load/pkg.go#L58
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
	GoFiles         []string `json:",omitempty"` // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles        []string `json:",omitempty"` // .go source files that import "C"
	CompiledGoFiles []string `json:",omitempty"` // .go output from running cgo on CgoFiles
	IgnoredGoFiles  []string `json:",omitempty"` // .go source files ignored due to build constraints
	CFiles          []string `json:",omitempty"` // .c source files
	CXXFiles        []string `json:",omitempty"` // .cc, .cpp and .cxx source files
	MFiles          []string `json:",omitempty"` // .m source files
	HFiles          []string `json:",omitempty"` // .h, .hh, .hpp and .hxx source files
	FFiles          []string `json:",omitempty"` // .f, .F, .for and .f90 Fortran source files
	SFiles          []string `json:",omitempty"` // .s source files
	SwigFiles       []string `json:",omitempty"` // .swig files
	SwigCXXFiles    []string `json:",omitempty"` // .swigcxx files
	SysoFiles       []string `json:",omitempty"` // .syso system object files added to package

	// Dependency information
	Deps      []string          `json:"Deps,omitempty"` // all (recursively) imported dependencies
	Imports   []string          `json:",omitempty"`     // import paths used by this package
	ImportMap map[string]string `json:",omitempty"`     // map from source import to ImportPath (identity entries omitted)

	// Error information
	Incomplete bool            `json:"Incomplete,omitempty"` // this package or a dependency has an error
	Error      *PackageError   `json:"Error,omitempty"`      // error loading package
	DepsErrors []*PackageError `json:"DepsErrors,omitempty"` // errors loading dependencies
}

// ModulePublic represents the package info of a module
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

// ModuleError represents the error loading module
type ModuleError struct {
	Err string // error text
}

// PackageError is the error info for a package when list failed
type PackageError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error (if present, file:line:col)
	Err         string   // the error itself
}

// CoverBuildInfo retreives some info from build
type CoverInfo struct {
	Target                   string
	GoPath                   string
	IsMod                    bool
	ModRootPath              string
	GlobalCoverVarImportPath string // path for the injected global cover var file
	OneMainPackage           bool
	Args                     string
	Mode                     string
	AgentPort                string
	Center                   string
}

//Execute inject cover variables for all the .go files in the target folder
func Execute(coverInfo *CoverInfo) error {
	target := coverInfo.Target
	newGopath := coverInfo.GoPath
	// oneMainPackage := coverInfo.OneMainPackage
	args := coverInfo.Args
	mode := coverInfo.Mode
	agentPort := coverInfo.AgentPort
	center := coverInfo.Center
	globalCoverVarImportPath := coverInfo.GlobalCoverVarImportPath

	if coverInfo.IsMod {
		globalCoverVarImportPath = filepath.Join(coverInfo.ModRootPath, globalCoverVarImportPath)
	} else {
		globalCoverVarImportPath = filepath.Base(globalCoverVarImportPath)
	}

	if !isDirExist(target) {
		log.Errorf("Target directory %s not exist", target)
		return ErrCoverPkgFailed
	}
	listArgs := []string{"-json"}
	if len(args) != 0 {
		listArgs = append(listArgs, args)
	}
	listArgs = append(listArgs, "./...")
	pkgs, err := ListPackages(target, strings.Join(listArgs, " "), newGopath)
	if err != nil {
		log.Errorf("Fail to list all packages, the error: %v", err)
		return err
	}

	var seen = make(map[string]*PackageCover)
	// var seenCache = make(map[string]*PackageCover)
	allDecl := ""
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			log.Printf("handle package: %v", pkg.ImportPath)
			// inject the main package
			mainCover, mainDecl := AddCounters(pkg, mode, globalCoverVarImportPath)
			allDecl += mainDecl
			// new a testcover for this service
			tc := TestCover{
				Mode:                     mode,
				AgentPort:                agentPort,
				Center:                   center,
				MainPkgCover:             mainCover,
				GlobalCoverVarImportPath: globalCoverVarImportPath,
			}

			// handle its dependency
			// var internalPkgCache = make(map[string][]*PackageCover)
			tc.CacheCover = make(map[string]*PackageCover)
			for _, dep := range pkg.Deps {
				if packageCover, ok := seen[dep]; ok {
					tc.DepsCover = append(tc.DepsCover, packageCover)
					continue
				}

				//only focus package neither standard Go library nor dependency library
				if depPkg, ok := pkgs[dep]; ok {
					packageCover, depDecl := AddCounters(depPkg, mode, globalCoverVarImportPath)
					allDecl += depDecl
					tc.DepsCover = append(tc.DepsCover, packageCover)
					seen[dep] = packageCover
				}
			}

			// inject Http Cover APIs
			var httpCoverApis = fmt.Sprintf("%s/http_cover_apis_auto_generated.go", pkg.Dir)
			if err := InjectCountersHandlers(tc, httpCoverApis); err != nil {
				log.Errorf("failed to inject counters for package: %s, err: %v", pkg.ImportPath, err)
				return ErrCoverPkgFailed
			}
		}
	}

	return injectGlobalCoverVarFile(coverInfo, allDecl)
}

// ListPackages list all packages under specific via go list command
// The argument newgopath is if you need to go list in a different GOPATH
func ListPackages(dir string, args string, newgopath string) (map[string]*Package, error) {
	cmd := exec.Command("/bin/bash", "-c", "go list "+args)
	log.Printf("go list cmd is: %v", cmd.Args)
	cmd.Dir = dir
	if newgopath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", newgopath))
	}
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	out, err := cmd.Output()
	if err != nil {
		log.Errorf("excute `go list -json ./...` command failed, err: %v, stdout: %v, stderr: %v", err, string(out), errbuf.String())
		return nil, ErrCoverListFailed
	}
	log.Infof("\n%v", errbuf.String())
	dec := json.NewDecoder(bytes.NewReader(out))
	pkgs := make(map[string]*Package, 0)
	for {
		var pkg Package
		if err := dec.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("reading go list output: %v", err)
			return nil, ErrCoverListFailed
		}
		if pkg.Error != nil {
			log.Errorf("list package %s failed with output: %v", pkg.ImportPath, pkg.Error)
			return nil, ErrCoverPkgFailed
		}

		// for _, err := range pkg.DepsErrors {
		// 	log.Fatalf("dependency package list failed, err: %v", err)
		// }

		pkgs[pkg.ImportPath] = &pkg
	}
	return pkgs, nil
}

// AddCounters is different from official go tool cover
// 1. only inject covervar++ into source file
// 2. no declarartions for these covervars
// 3. return the declarations as string
func AddCounters(pkg *Package, mode string, globalCoverVarImportPath string) (*PackageCover, string) {
	coverVarMap := declareCoverVars(pkg)

	decl := ""
	for file, coverVar := range coverVarMap {
		decl += "\n" + tool.Annotate(path.Join(pkg.Dir, file), mode, coverVar.Var, globalCoverVarImportPath) + "\n"
	}

	return &PackageCover{
		Package: pkg,
		Vars:    coverVarMap,
	}, decl
}

func isDirExist(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// Refer: https://github.com/golang/go/blob/master/src/cmd/go/internal/load/pkg.go#L1334:6
// hasInternalPath looks for the final "internal" path element in the given import path.
// If there isn't one, hasInternalPath returns ok=false.
// Otherwise, hasInternalPath returns ok=true and the index of the "internal".
func hasInternalPath(path string) bool {
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

func buildCoverCmd(file string, coverVar *FileVar, pkg *Package, mode, newgopath string) *exec.Cmd {
	// to construct: go tool cover -mode=atomic -o dest src (note: dest==src)
	var newArgs = []string{"tool", "cover"}
	newArgs = append(newArgs, "-mode", mode)
	newArgs = append(newArgs, "-var", coverVar.Var)
	longPath := path.Join(pkg.Dir, file)
	newArgs = append(newArgs, "-o", longPath, longPath)
	cmd := exec.Command("go", newArgs...)
	if newgopath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", newgopath))
	}
	return cmd
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

func cacheInternalCover(in *PackageCover) *PackageCover {
	c := &PackageCover{}
	vars := declareCacheVars(in)
	c.Package = in.Package
	c.Vars = vars
	return c
}

func addCacheCover(pkg *Package, in *PackageCover) *PackageCover {
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
type CoverageList []Coverage

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

// CovList converts profile to CoverageList struct
func CovList(f io.Reader) (g CoverageList, err error) {
	scanner := bufio.NewScanner(f)
	scanner.Scan() // discard first line
	g = NewCoverageList()

	for scanner.Scan() {
		row := scanner.Text()
		blk, err := toBlock(row)
		if err != nil {
			return nil, err
		}
		blk.addToGroupCov(&g)
	}
	return
}

// ReadFileToCoverList coverts profile file to CoverageList struct
func ReadFileToCoverList(path string) (g CoverageList, err error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Errorf("Open file %s failed!", path)
		return nil, err
	}
	g, err = CovList(bytes.NewReader(f))
	return
}

// NewCoverageList return empty CoverageList
func NewCoverageList() CoverageList {
	return CoverageList{}

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

func (g CoverageList) size() int {
	return len(g)
}

func (g CoverageList) lastElement() *Coverage {
	return &g[g.size()-1]
}

func (g *CoverageList) append(c *Coverage) {
	*g = append(*g, *c)
}

// Sort sorts CoverageList with filenames
func (g CoverageList) Sort() {
	sort.SliceStable(g, func(i, j int) bool {
		return g[i].Name() < g[j].Name()
	})
}

// TotalPercentage returns the total percentage of coverage
func (g CoverageList) TotalPercentage() string {
	ratio, err := g.TotalRatio()
	if err == nil {
		return PercentStr(ratio)
	}
	return "N/A"
}

// TotalRatio returns the total ratio of covered statements
func (g CoverageList) TotalRatio() (ratio float32, err error) {
	var total Coverage
	for _, c := range g {
		total.NCoveredStmts += c.NCoveredStmts
		total.NAllStmts += c.NAllStmts
	}
	return total.Ratio()
}

// Map returns maps the file name to its coverage for faster retrieval
// & membership check
func (g CoverageList) Map() map[string]Coverage {
	m := make(map[string]Coverage)
	for _, c := range g {
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

// Ratio calculates the ratio of statements in a profile
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
