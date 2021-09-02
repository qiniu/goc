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

package server

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	"github.com/gin-gonic/gin"
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
			"msg": "missing some params",
		})
		return
	}
	// 计算插桩服务 id
	agentId := gs.generateAgentId(remoteIP.String(), hostname, cmdline, pid)
	// 检查 id 是否重复
	if _, ok := gs.rpcAgents.Load(agentId); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "the rpc agent already exists",
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
		log.Errorf("fail to establish websocket connection with rpc agent: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
	}

	// send close msg and close ws connection
	defer func() {
		deadline := 1 * time.Second
		// 发送 close msg
		gs.wsclose(ws, deadline)
		time.Sleep(deadline)
		// 从维护的 websocket 链接字典中移除
		gs.rpcAgents.Delete(agentId)
		ws.Close()
		log.Infof("close rpc connection, %v", hostname)
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
				log.Errorf("rpc ping to %v failed: %v", hostname, err)
				break
			}
		}

		gocA.once.Do(func() {
			close(gocA.exitCh)
		})
	}()

	log.Infof("one rpc agent established, %v, cmdline: %v, pid: %v, hostname: %v", ws.RemoteAddr(), cmdline, pid, hostname)
	// new rpc agent
	// 在这里 websocket server 作为 rpc 的客户端，
	// 发送 rpc 请求，
	// 由被插桩服务返回 rpc 应答
	rwc := &ReadWriteCloser{ws: ws}
	codec := jsonrpc.NewClientCodec(rwc)

	gocA.rpc = rpc.NewClientWithCodec(codec)
	gocA.Id = string(agentId)
	gs.rpcAgents.Store(agentId, gocA)
	// wait for exit
	<-gocA.exitCh
}

// generateAgentId generate id based on agent's meta infomation
func (gs *gocServer) generateAgentId(args ...string) gocCliendId {
	var path string
	for _, arg := range args {
		path += arg
	}
	sum := sha256.Sum256([]byte(path))
	h := fmt.Sprintf("%x", sum[:6])

	return gocCliendId(h)
}
