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

package cover

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"encoding/json"
	"os"
	"strings"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"testing"

	"{{.GlobalCoverVarImportPath}}/websocket"

	_cover "{{.GlobalCoverVarImportPath}}"
)

var (
	waitDelay time.Duration = 5 * time.Second
	host      string        = "{{.Host}}"
)

var (
	token string
	id string
	cond = sync.NewCond(&sync.Mutex{})
	register_extra = os.Getenv("GOC_REGISTER_EXTRA")
)

func init() {
	// init host
	host_env := os.Getenv("GOC_CUSTOM_HOST")
	if host_env != "" {
		host = host_env
	}

	var dialer = websocket.DefaultDialer

	go func() {
		register(host)

		cond.L.Lock()
		cond.Broadcast()
		cond.L.Unlock()

		// 永不退出，出错后统一操作为：延时 + conitnue
		for {
			// 直接将 token 放在 ws 地址中
			v := url.Values{}
			v.Set("token", token)
			v.Set("id", id)
			v.Encode()

			rpcstreamUrl := fmt.Sprintf("ws://%v/v2/internal/ws/rpcstream?%v", host, v.Encode())
			ws, resp, err := dialer.Dial(rpcstreamUrl, nil)
			if err != nil {
				if resp != nil {
					tmp, _ := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					log.Printf("[goc][Error] rpc fail to dial to goc server: %v, body: %v", err, string(tmp))
					
					if isOffline(tmp) {
						log.Printf("[goc][Error] needs re-register")
						register(host)
					}
				} else {
					log.Printf("[goc][Error] rpc fail to dial to goc server: %v", err)
				}
				time.Sleep(waitDelay)
				continue
			}
			log.Printf("[goc][Info] rpc connected to goc server")

			rwc := &ReadWriteCloser{ws: ws}
			s := rpc.NewServer()
			s.Register(&GocAgent{})
			s.ServeCodec(jsonrpc.NewServerCodec(rwc))

			// exit rpc server, close ws connection
			ws.Close()
			time.Sleep(waitDelay)
			log.Printf("[goc][Error] rpc connection to goc server broken", )
		}
	}()
}

// register
func register (host string) {
	for {
		// 获取进程元信息用于注册
		ps, err := getRegisterInfo()
		if err != nil {
			time.Sleep(waitDelay)
			continue
		}

		// 注册，直接将元信息放在 ws 地址中
		v := url.Values{}
		v.Set("hostname", ps.hostname)
		v.Set("pid", strconv.Itoa(ps.pid))
		v.Set("cmdline", ps.cmdline)
		v.Set("extra", register_extra)
		v.Encode()

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%v/v2/internal/register?%v", host, v.Encode()), nil)
		if err != nil {
			log.Printf("[goc][Error] register generate register http request: %v", err)
			time.Sleep(waitDelay)
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("[goc][Error] register fail to goc server: %v", err)
			time.Sleep(waitDelay)
			continue				
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[goc][Error] fail to get register resp from the goc server: %v", err)
			time.Sleep(waitDelay)
			continue	
		}

		if resp.StatusCode != 200 {
			log.Printf("[goc][Error] wrong register http statue code: %v", resp.StatusCode)
			time.Sleep(waitDelay)
			continue
		}

		registerResp := struct {
			Token string `json:"token"`
			Id string 	 `json:"id"`
		}{}

		err = json.Unmarshal(body, &registerResp)
		if err != nil {
			log.Printf("[goc][Error] register response json unmarshal failed: %v", err)
			time.Sleep(waitDelay)
			continue				
		}

		// register success
		token = registerResp.Token
		id = registerResp.Id
		break
	}
}

// check if offline failed
func isOffline(data []byte) bool {
	val := struct {
		Code int `json:"code"`
	}{}
	err := json.Unmarshal(data, &val)
	if err != nil {
		return true
	}
	if val.Code == 1 {
		return true
	}
	return false
}

// rpc
type GocAgent struct {
}

type ProfileReq string

type ProfileRes string

// return a profile of now
func (ga *GocAgent) GetProfile(req *ProfileReq, res *ProfileRes) error {
	if *req != "getprofile" {
		*res = ""
		return fmt.Errorf("wrong command")
	}

	w := new(strings.Builder)
	fmt.Fprint(w, "mode: {{.Mode}}\n")

	counters, blocks := loadValues()
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
				return err
			}
		}
	}

	*res = ProfileRes(w.String())

	return nil
}

// reset profile to 0
func (ga *GocAgent) ResetProfile(req *ProfileReq, res *ProfileRes) error {
	if *req != "resetprofile" {
		*res = ""
		return fmt.Errorf("wrong command")
	}

	resetValues()

	*res = `ok`

	return nil
}

// get cover Values

func loadValues() (map[string][]uint32, map[string][]testing.CoverBlock) {
	var (
		coverCounters = make(map[string][]uint32)
		coverBlocks   = make(map[string][]testing.CoverBlock)
	)

	{{range $i, $pkgCover := .Covers}}
	{{range $file, $cover := $pkgCover.Vars}}
	loadFileCover(coverCounters, coverBlocks, "{{$cover.File}}", _cover.{{$cover.Var}}.Count[:], _cover.{{$cover.Var}}.Pos[:], _cover.{{$cover.Var}}.NumStmt[:])
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

// reset counters
func resetValues() {
	{{range $i, $pkgCover := .Covers}}
	{{range $file, $cover := $pkgCover.Vars}}
	clearFileCover(_cover.{{$cover.Var}}.Count[:])
	{{end}}
	{{end}}
}

func clearFileCover(counter []uint32) {
	for i := range counter {
		counter[i] = 0
	}
}


// get process meta info for register
type processInfo struct {
	hostname string
	pid      int
	cmdline  string
}

func getRegisterInfo() (*processInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("[goc][Error] fail to get hostname: %v", hostname)
		return nil, err
	}

	pid := os.Getpid()

	cmdline := strings.Join(os.Args, " ")

	return &processInfo{
		hostname: hostname,
		pid:      pid,
		cmdline:  cmdline,
	}, nil
}

/// websocket rpc readwriter closer

type ReadWriteCloser struct {
	ws *websocket.Conn
	r  io.Reader
	w  io.WriteCloser
}

func (rwc *ReadWriteCloser) Read(p []byte) (n int, err error) {
	if rwc.r == nil {
		var _ int
		_, rwc.r, err = rwc.ws.NextReader()
		if err != nil {
			return 0, err
		}
	}
	for n = 0; n < len(p); {
		var m int
		m, err = rwc.r.Read(p[n:])
		n += m
		if err == io.EOF {
			// done
			rwc.r = nil
			break
		}
		// ???
		if err != nil {
			break
		}
	}
	return
}

func (rwc *ReadWriteCloser) Write(p []byte) (n int, err error) {
	if rwc.w == nil {
		rwc.w, err = rwc.ws.NextWriter(websocket.TextMessage)
		if err != nil {
			return 0, err
		}
	}
	for n = 0; n < len(p); {
		var m int
		m, err = rwc.w.Write(p)
		n += m
		if err != nil {
			break
		}
	}
	if err != nil || n == len(p) {
		err = rwc.Close()
	}
	return
}

func (rwc *ReadWriteCloser) Close() (err error) {
	if rwc.w != nil {
		err = rwc.w.Close()
		rwc.w = nil
	}
	return err
}
