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

package cover

import (
	"fmt"
	"os"
	"path"
	"text/template"
)

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

const coverMain = `
// Code generated by goc system. DO NOT EDIT.

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	{{range $i, $pkgCover := .DepsCover}}
	_cover{{$i}} {{$pkgCover.Package.ImportPath | printf "%q"}}
	{{end}}

	{{range $k, $pkgCover := .CacheCover}}
	{{$pkgCover.Package.ImportPath | printf "%q"}}
	{{end}}

)

func init() {
	go registerHandlers()
}

func loadValues() (map[string][]uint32, map[string][]testing.CoverBlock) {
	var (
		coverCounters = make(map[string][]uint32)
		coverBlocks   = make(map[string][]testing.CoverBlock)
	)

	{{range $i, $pkgCover := .DepsCover}}
	{{range $file, $cover := $pkgCover.Vars}}
	loadFileCover(coverCounters, coverBlocks, {{printf "%q" $cover.File}}, _cover{{$i}}.{{$cover.Var}}.Count[:], _cover{{$i}}.{{$cover.Var}}.Pos[:], _cover{{$i}}.{{$cover.Var}}.NumStmt[:])
	{{end}}
	{{end}}

	{{range $file, $cover := .MainPkgCover.Vars}}
	loadFileCover(coverCounters, coverBlocks, {{printf "%q" $cover.File}}, {{$cover.Var}}.Count[:], {{$cover.Var}}.Pos[:], {{$cover.Var}}.NumStmt[:])
	{{end}}

	{{range $k, $pkgCover := .CacheCover}}
	{{range $v, $cover := $pkgCover.Vars}}
	loadFileCover(coverCounters, coverBlocks, {{printf "%q" $cover.File}}, {{$pkgCover.Package.Name}}.{{$v}}.Count[:], {{$pkgCover.Package.Name}}.{{$v}}.Pos[:], {{$pkgCover.Package.Name}}.{{$v}}.NumStmt[:])
	{{end}}
	{{end}}

	return coverCounters, coverBlocks
}

func loadFileCover(coverCounters map[string][]uint32, coverBlocks map[string][]testing.CoverBlock, fileName string, counter []uint32, pos []uint32, numStmts []uint16) {
	if 3*len(counter) != len(pos) || len(counter) != len(numStmts) {
		panic("coverage: mismatched sizes")
	}
	if coverCounters[fileName] != nil {
		// Already registered.
		return
	}
	coverCounters[fileName] = counter
	block := make([]testing.CoverBlock, len(counter))
	for i := range counter {
		block[i] = testing.CoverBlock{
			Line0: pos[3*i+0],
			Col0:  uint16(pos[3*i+2]),
			Line1: pos[3*i+1],
			Col1:  uint16(pos[3*i+2] >> 16),
			Stmts: numStmts[i],
		}
	}
	coverBlocks[fileName] = block
}

func clearValues() {

	{{range $i, $pkgCover := .DepsCover}}
	{{range $file, $cover := $pkgCover.Vars}}
	clearFileCover(_cover{{$i}}.{{$cover.Var}}.Count[:])
	{{end}}
	{{end}}

	{{range $file, $cover := .MainPkgCover.Vars}}
	clearFileCover({{$cover.Var}}.Count[:])
	{{end}}

	{{range $k, $pkgCover := .CacheCover}}
	{{range $v, $cover := $pkgCover.Vars}}
	clearFileCover({{$pkgCover.Package.Name}}.{{$v}}.Count[:])
	{{end}}
	{{end}}

}

func clearFileCover(counter []uint32) {
	for i := range counter {
		counter[i] = 0
	}
}

func registerHandlers() {
	ln, host, err := listen()
	if err != nil {
		log.Fatalf("profile listen failed, err:%v", err)
	}
	log.Println("profile listen on", host)
	profileAddr := "http://" + host
	if resp, err := registerSelf(profileAddr); err != nil {
		log.Fatalf("register address %v failed, err: %v, response: %v", profileAddr, err, string(resp))
	}
	go genProfileAddr(host)
	//clear and profile is mutex,
	var lock atomic.Value
	mux := http.NewServeMux()
	// Coverage reports the current code coverage as a fraction in the range [0, 1].
	// If coverage is not enabled, Coverage returns 0.
	mux.HandleFunc("/v1/cover/coverage", func(w http.ResponseWriter, r *http.Request) {
		lock.Store(false)
		counters, _ := loadValues()
		lock.Store(true)

		var n, d int64
		for _, counter := range counters {
			for i := range counter {
				if atomic.LoadUint32(&counter[i]) > 0 {
					n++
				}
				d++
			}
		}
		if d == 0 {
			fmt.Fprint(w, 0)
			return
		}
		fmt.Fprintf(w, "%f", float64(n)/float64(d))
	})

	// coverprofile reports a coverage profile with the coverage percentage
	mux.HandleFunc("/v1/cover/profile", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "mode: {{.Mode }} \n")
		lock.Store(false)
		counters, blocks := loadValues()

		lock.Store(true)
		var active, total int64
		var count uint32
		for name, counts := range counters {
			block := blocks[name]
			for i := range counts {
				stmts := int64(block[i].Stmts)
				total += stmts
				count = atomic.LoadUint32(&counts[i]) // For -mode=atomic.
				if count > 0 {
					active += stmts
				}
				_, err := fmt.Fprintf(w, "%s:%d.%d,%d.%d %d %d\n", name,
					block[i].Line0, block[i].Col0,
					block[i].Line1, block[i].Col1,
					stmts,
					count)
				if err != nil {
					fmt.Fprintf(w, "invalid block format, err: %v", err)
					return
				}
			}
		}
	})

	mux.HandleFunc("/v1/cover/clear", func(w http.ResponseWriter, r *http.Request) {
		f := lock.Load().(bool)
		if f {
			clearValues()
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w,"clear call successfully")
		}else {
			http.Error(w, "clear call failed cause by coverage profile dump", http.StatusExpectationFailed)
		}
	})

	log.Fatal(http.Serve(ln, mux))
}

func registerSelf(address string) ([]byte, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/cover/register?name=%s&address=%s", {{.Center | printf "%q"}}, os.Args[0], address), nil)
	if err != nil {
		log.Fatalf("http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil && isNetworkError(err) {
		log.Printf("[WARN]error occured:%v, try again", err)
		resp, err = http.DefaultClient.Do(req)
	}
	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("registed faile, err:%v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body, err:%v", err)
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("registed failed, response code %d", resp.StatusCode)
	}

	return body, err
}

func isNetworkError(err error) bool {
	if err == io.EOF {
		return true
	}
	_, ok := err.(net.Error)
	return ok
}

func listen() (ln net.Listener, host string, err error) {
	// 获取上次使用的监听地址
	if previousAddr := getPreviousAddr(); previousAddr != "" {
		ss := strings.Split(previousAddr, ":")
		// listen on all network interface
		ln, err = net.Listen("tcp4", ":"+ss[len(ss)-1])
		if err == nil {
			host = previousAddr
			return
		}
	}

	ln, err = net.Listen("tcp4", ":0")
	if err != nil {
		return
	}

	adds, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	var localIPV4 string
	var nonLocalIPV4 string
	for _, addr := range adds {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			if ipNet.IP.IsLoopback() {
				localIPV4 = ipNet.IP.String()
			} else {
				nonLocalIPV4 = ipNet.IP.String()
			}
		}
	}
	if nonLocalIPV4 != "" {
		host = fmt.Sprintf("%s:%d", nonLocalIPV4, ln.Addr().(*net.TCPAddr).Port)
	} else {
		host = fmt.Sprintf("%s:%d", localIPV4, ln.Addr().(*net.TCPAddr).Port)
	}
	return
}

func getPreviousAddr() string {
	file, err := os.Open(os.Args[0] + "_profile_listen_addr")
	if err != nil {
		return ""
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	addr, _, _ := reader.ReadLine()
	return string(addr)
}

func genProfileAddr(profileAddr string) {
	fn := os.Args[0] + "_profile_listen_addr"
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, strings.TrimPrefix(profileAddr, "http://"))
}
`

var coverParentFileTmpl = template.Must(template.New("coverParentFileTmpl").Parse(coverParentFile))

const coverParentFile = `
// Code generated by goc system. DO NOT EDIT.

package {{.}}

`

var coverParentVarsTmpl = template.Must(template.New("coverParentVarsTmpl").Parse(coverParentVars))

const coverParentVars = `

import (

	{{range $i, $pkgCover := .}}
	_cover{{$i}} {{$pkgCover.Package.ImportPath | printf "%q"}}
	{{end}} 

)

{{range $i, $pkgCover := .}}
{{range $v, $cover := $pkgCover.Vars}}
var {{$v}} = &_cover{{$i}}.{{$cover.Var}}
{{end}}
{{end}}
	
`

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
