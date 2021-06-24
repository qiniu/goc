package cover

import (
	"fmt"
	"io"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"testing"

	"{{.GlobalCoverVarImportPath}}/websocket"

	_cover "{{.GlobalCoverVarImportPath}}"
)

var (
	waitDelay time.Duration = 10 * time.Second
	host      string        = "{{.Host}}"
)

func init() {
	// init host
	host_env := os.Getenv("GOC_CUSTOM_HOST")
	if host_env != "" {
		host = host_env
	}

	var dialer = websocket.DefaultDialer

	go func() {
		// 永不退出，出错后统一操作为：延时 + conitnue
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
			v.Encode()

			rpcstreamUrl := fmt.Sprintf("ws://%v/v2/internal/ws/rpcstream?%v", host, v.Encode())
			ws, _, err := dialer.Dial(rpcstreamUrl, nil)
			if err != nil {
				log.Printf("[goc][Error] rpc fail to dial to goc server: %v", err)
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

	cmdline := os.Args[0]

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
