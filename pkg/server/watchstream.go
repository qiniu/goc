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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/qiniu/goc/v2/pkg/log"
)

func (gs *gocServer) serveWatchInternalStream(c *gin.Context) {
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
	if _, ok := gs.watchAgents.Load(agentId); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "the watch agent already exist",
		})
		return
	}
	// upgrade to websocket
	ws, err := gs.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("fail to establish websocket connection with watch agent: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
	}

	// send close msg and close ws connection
	defer func() {
		gs.watchAgents.Delete(agentId)
		ws.Close()
		log.Infof("close watch connection, %v", hostname)
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
				log.Errorf("watch ping to %v failed: %v", hostname, err)
				break
			}
		}
	}()

	log.Infof("one watch agent established, %v, cmdline: %v, pid: %v, hostname: %v", ws.RemoteAddr(), cmdline, pid, hostname)

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Errorf("read from %v: %v", hostname, err)
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
