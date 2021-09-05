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
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/qiniu/goc/v2/pkg/log"
)

func (gs *gocServer) serveWatchInternalStream(c *gin.Context) {
	// 检查插桩服务上报的信息
	watchRemoteIP, _ := c.RemoteIP()
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
	agent.WatchRemoteIP = watchRemoteIP.String()
	agent.Status &= ^DISCONNECT  // 取消 DISCONNECT 的状态
	agent.Status |= WATCHCONNECT // 设置为 RPC CONNECT 状态
	var once sync.Once

	// upgrade to websocket
	ws, err := gs.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("fail to establish websocket connection with watch agent: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
	}

	// 注册销毁函数
	agent.closeWatchConnOnce = func() {
		once.Do(func() {
			// 关闭 ws 连接后，ws.ReadMessage() 会出错退出 goroutine，进入 defer
			ws.Close()
		})
	}

	// send close msg and close ws connection
	defer func() {
		// 取消 WATCH CONNECT 状态
		agent.Status &= ^WATCHCONNECT
		if agent.Status == 0 {
			agent.Status = DISCONNECT
		}

		agent.closeWatchConnOnce()

		log.Infof("close watch connection, %v", agent.Hostname)
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
				log.Errorf("watch ping to %v failed: %v", agent.Hostname, err)
				break
			}
		}
	}()

	log.Infof("one watch agent established, %v, cmdline: %v, pid: %v, hostname: %v", ws.RemoteAddr(), agent.CmdLine, agent.Pid, agent.Hostname)

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Errorf("read from %v: %v", agent.Hostname, err)
			break
		}
		if mt == websocket.TextMessage {
			gs.watchCh <- message
		}
	}
}

func (gs *gocServer) watchLoop() {
	for {
		msg := <-gs.watchCh
		gs.watchClients.Range(func(key, value interface{}) bool {
			// 这里是客户端的 ws 连接，不是 agent ws 连接
			gwc := value.(*gocWatchClient)
			err := gwc.ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				gwc.ws.Close()
				gwc.once.Do(func() { close(gwc.exitCh) })
			}

			return true
		})
	}
}
