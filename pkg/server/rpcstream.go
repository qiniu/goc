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

// serveRpcStream holds connection between goc server and agent.
//
// 1. goc server 作为 rpc 客户端
//
// 2. 每个链接的 goc agent 作为 rpc 服务端
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

	gocA := &gocCoveredAgent{
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
		// 发送 close msg
		gs.wsclose(ws, deadline)
		time.Sleep(deadline)
		// 从维护的 websocket 链接字典中移除
		gs.rpcClients.Delete(clientId)
		ws.Close()
		log.Infof("close connection, %v", hostname)
	}()

	// set pong handler
	ws.SetReadDeadline(time.Now().Add(PongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	// set ping goroutine to ping every PingWait time
	go func() {
		ticker := time.NewTicker(PingWait)
		defer ticker.Stop()

		for range ticker.C {
			if err := gs.wsping(ws, PongWait); err != nil {
				log.Errorf("ping to %v failed: %v", hostname, err)
				break
			}
		}

		gocA.once.Do(func() {
			close(gocA.exitCh)
		})
	}()

	log.Infof("one client established, %v, cmdline: %v, pid: %v, hostname: %v", ws.RemoteAddr(), cmdline, pid, hostname)
	// new rpc client
	// 在这里 websocket server 作为 rpc 的客户端，
	// 发送 rpc 请求，
	// 由被插桩服务返回 rpc 应答
	rwc := &ReadWriteCloser{ws: ws}
	codec := jsonrpc.NewClientCodec(rwc)

	gocA.rpc = rpc.NewClientWithCodec(codec)
	gocA.Id = string(clientId)
	gs.rpcClients.Store(clientId, gocA)
	// wait for exit
	<-gocA.exitCh
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
