package coverdef

import (
	"fmt"
	"time"
	"os"
	"log"
	"strconv"
	"net/url"

	"{{.GlobalCoverVarImportPath}}/websocket"
)

var (
	watchChannel = make(chan *blockInfo, 1024)

	watchEnabled = false

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

			watchstreamUrl := fmt.Sprintf("ws://%v/v2/internal/ws/watchstream?%v", host, v.Encode())
			ws, _, err := dialer.Dial(watchstreamUrl, nil)
			if err != nil {
				log.Printf("[goc][Error] watch fail to dial to goc server: %v", err)
				time.Sleep(waitDelay)
				continue
			}

			// 连接成功
			watchEnabled = true
			log.Printf("[goc][Info] watch connected to goc server")

			ticker := time.NewTicker(time.Second)
			closeFlag := false
			go func() {
				for {
					// 必须调用一下以触发 ping 的自动处理
					_, _, err := ws.ReadMessage()
					if err != nil {
						break
					}
				}
				closeFlag = true
			}()

			Loop:
			for {
				select {
				case block := <-watchChannel:
					i := block.i

					cov := fmt.Sprintf("%s:%d.%d,%d.%d %d %d", block.name,
						block.pos[3*i+0], uint16(block.pos[3*i+2]),
						block.pos[3*i+1], uint16(block.pos[3*i+2] >> 16),
						1,
						0)

					err = ws.WriteMessage(websocket.TextMessage, []byte(cov))
					if err != nil {
						watchEnabled = false
						log.Println("[goc][Error] push coverage failed: %v", err)
						time.Sleep(waitDelay)
						break Loop
					}
				case <-ticker.C:
					if closeFlag == true {
						break Loop
					}
				}
			}
		}
	}()
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

//

type blockInfo struct {
	name string
	pos  []uint32
	i    int
}

// UploadCoverChangeEvent_{{.Random}} is non-blocking
func UploadCoverChangeEvent_{{.Random}}(name string, pos []uint32, i int) {

	if watchEnabled == false {
		return
	}

	// make sure send is non-blocking
	select {
	case watchChannel <- &blockInfo{
		name: name,
		pos:  pos,
		i:    i,
	}:
	default:
	}
}
