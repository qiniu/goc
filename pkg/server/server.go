package server

import (
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// gocServer represents a goc server
type gocServer struct {
	port      int
	storePath string
	upgrader  websocket.Upgrader

	rpcAgents    sync.Map
	watchAgents  sync.Map
	watchCh      chan []byte
	watchClients sync.Map
}

type gocCliendId string

// gocCoveredAgent represents a covered client
type gocCoveredAgent struct {
	Id       string      `json:"id"`
	RemoteIP string      `json:"remoteip"`
	Hostname string      `json:"hostname"`
	CmdLine  string      `json:"cmdline"`
	Pid      string      `json:"pid"`
	rpc      *rpc.Client `json:"-"`

	exitCh chan int  `json:"-"`
	once   sync.Once `json:"-"` // 保护 close(exitCh) 只执行一次
}

//  api 客户端，不是 agent
type gocWatchClient struct {
	ws     *websocket.Conn
	exitCh chan int
	once   sync.Once
}

func RunGocServerUntilExit(host string) {
	gs := gocServer{
		storePath: "",
		upgrader: websocket.Upgrader{
			ReadBufferSize:   4096,
			WriteBufferSize:  4096,
			HandshakeTimeout: 45 * time.Second,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		watchCh: make(chan []byte, 4096),
	}

	r := gin.Default()
	v2 := r.Group("/v2")
	{
		v2.GET("/cover/profile", gs.getProfiles)
		v2.DELETE("/cover/profile", gs.resetProfiles)
		v2.GET("/rpcagents", gs.listAgents)
		v2.GET("/watchagents", nil)

		v2.GET("/cover/ws/watch", gs.watchProfileUpdate)

		// internal use only
		v2.GET("/internal/ws/rpcstream", gs.serveRpcStream)
		v2.GET("/internal/ws/watchstream", gs.serveWatchInternalStream)
	}

	go gs.watchLoop()

	r.Run(host)
}
