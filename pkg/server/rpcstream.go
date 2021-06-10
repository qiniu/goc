package server

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/qiniu/goc/v2/pkg/log"
)

// GocCommandArg defines server -> client arg
type GocCommandArg struct {
	Type    string
	Content string
}

// GocCommandReply defines server -> client reply
type GocCommandReply struct {
	Type    string
	Code    string
	Content string
}

func (gs *gocServer) serveRpcStream(c *gin.Context) {
	// 检查插桩服务上报的信息
	remoteIP, _ := c.RemoteIP()
	hostname := c.Query("hostname")
	pid := c.Query("pid")
	cmdline := c.Query("cmdline")

	if hostname == "" || pid == "" || cmdline == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "missing some param",
		})
		return
	}
	// 计算插桩服务 id
	clientId := gs.generateClientId(remoteIP.String(), hostname, cmdline, pid)
	// 检查 id 是否重复
	if _, ok := gs.rpcClients.Load(clientId); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "client already exist",
		})
		return
	}

	gocC := gocCoveredClient{
		RemoteIP: remoteIP.String(),
		Hostname: hostname,
		Pid:      pid,
		CmdLine:  cmdline,
		exitCh:   make(chan int),
	}

	// upgrade to websocket
	ws, err := gs.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("fail to establish websocket connection with client: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
	}

	// send close msg and close ws connection
	defer func() {
		deadline := 1 * time.Second
		gs.wsclose(ws, deadline)
		time.Sleep(deadline)
		ws.Close()
	}()

	// set pong handler
	ws.SetReadDeadline(time.Now().Add(PongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	// set ping goroutine to ping every PingWait time
	go func() {
		ticker := time.Tick(PingWait)
		for range ticker {
			if err := gs.wsping(ws, PongWait); err != nil {
				log.Errorf("ping to %v failed: %v", ws.RemoteAddr(), err)
				break
			}
		}
		gs.wsclose(ws, 1)
	}()

	// new rpc client
	// 在这里 websocket server 作为 rpc 的客户端，
	// 发送 rpc 请求，
	// 由被插桩服务返回 rpc 应答
	rwc := &ReadWriteCloser{ws: ws}
	codec := jsonrpc.NewClientCodec(rwc)

	gocC.rpc = rpc.NewClientWithCodec(codec)
	gocC.Id = string(clientId)
	gs.rpcClients.Store(clientId, gocC)
	// wait for exit
	for {
		<-gocC.exitCh
	}
}

func (gs *gocServer) wsping(ws *websocket.Conn, deadline time.Duration) error {
	return ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(deadline*time.Second))
}

func (gs *gocServer) wsclose(ws *websocket.Conn, deadline time.Duration) error {
	return ws.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(deadline*time.Second))
}

// generateClientId generate id based on client's meta infomation
func (gs *gocServer) generateClientId(args ...string) gocCliendId {
	var path string
	for _, arg := range args {
		path += arg
	}
	sum := sha256.Sum256([]byte(path))
	h := fmt.Sprintf("%x", sum[:6])

	return gocCliendId(h)
}
