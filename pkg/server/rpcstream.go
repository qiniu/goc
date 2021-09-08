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
	"sync"
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
	rpcRemoteIP, _ := c.RemoteIP()
	id := c.Query("id")
	token := c.Query("token")

	rawagent, ok := gs.agents.Load(id)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  "agent not registered",
			"code": 1,
		})
		return
	}

	agent := rawagent.(*gocCoveredAgent)
	if agent.Token != token {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  "register token not match",
			"code": 1,
		})
		return
	}

	// 更新 agent 信息
	agent.RpcRemoteIP = rpcRemoteIP.String()
	agent.exitCh = make(chan int)
	agent.Status &= ^DISCONNECT // 取消 DISCONNECT 的状态
	agent.Status |= RPCCONNECT  // 设置为 RPC CONNECT 状态
	// 注册销毁函数
	var once sync.Once
	agent.closeRpcConnOnce = func() {
		once.Do(func() {
			// 为什么只是关闭 channel？其它资源如何释放？
			// close channel 后，本 goroutine 会进入到 defer
			close(agent.exitCh)
		})
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

		// 取消 RPC CONNECT 状态
		agent.Status &= ^RPCCONNECT
		if agent.Status == 0 {
			agent.Status = DISCONNECT
		}

		ws.Close()
		log.Infof("close rpc connection, %v", agent.Hostname)
		// reset rpc client
		agent.rpc = nil
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
				log.Errorf("rpc ping to %v failed: %v", agent.Hostname, err)
				break
			}
		}

		agent.closeRpcConnOnce()
	}()

	log.Infof("one rpc agent established, %v, cmdline: %v, pid: %v, hostname: %v", ws.RemoteAddr(), agent.CmdLine, agent.Pid, agent.Hostname)
	// new rpc agent
	// 在这里 websocket server 作为 rpc 的客户端，
	// 发送 rpc 请求，
	// 由被插桩服务返回 rpc 应答
	rwc := &ReadWriteCloser{ws: ws}
	codec := jsonrpc.NewClientCodec(rwc)

	agent.rpc = rpc.NewClientWithCodec(codec)

	// wait for exit
	<-agent.exitCh
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
