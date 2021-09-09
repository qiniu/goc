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
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/qiniu/goc/v2/pkg/log"
	"github.com/qiniu/goc/v2/pkg/server/store"
)

// gocServer represents a goc server
type gocServer struct {
	port  int
	store store.Store

	upgrader websocket.Upgrader

	agents sync.Map

	watchCh      chan []byte
	watchClients sync.Map

	idCount int64
	idL     sync.Mutex
}

type gocCliendId string

const (
	DISCONNECT   = 1 << iota
	RPCCONNECT   = 1 << iota
	WATCHCONNECT = 1 << iota
)

// gocCoveredAgent represents a covered client
type gocCoveredAgent struct {
	Id            string `json:"id"`
	RpcRemoteIP   string `json:"rpc_remoteip"`
	WatchRemoteIP string `json:"watch_remoteip"`
	Hostname      string `json:"hostname"`
	CmdLine       string `json:"cmdline"`
	Pid           string `json:"pid"`

	// 用户可以选择上报一些定制信息
	// 比如不同 namespace 的 statefulset POD，它们的 hostname/cmdline/pid 都是一样的，
	// 这时候将 extra 设置为 namespace 并上报，这个额外的信息在展示时将更友好
	Extra string `json:"extra"`

	Token  string `json:"token"`
	Status int    `json:"status"` // 表示该 agent 是否处于 connected 状态

	rpc *rpc.Client `json:"-"`

	exitCh             chan int `json:"-"`
	closeRpcConnOnce   func()   `json:"-"` // close rpc conn 只执行一次
	closeWatchConnOnce func()   `json:"-"` // close watch conn 只执行一次
}

func (agent *gocCoveredAgent) closeConnection() {
	if agent.closeRpcConnOnce != nil {
		agent.closeRpcConnOnce()
	}

	if agent.closeWatchConnOnce != nil {
		agent.closeWatchConnOnce()
	}
}

//  api 客户端，不是 agent
type gocWatchClient struct {
	ws     *websocket.Conn
	exitCh chan int
	once   sync.Once
}

func RunGocServerUntilExit(host string, s store.Store) error {
	gs := gocServer{
		store: s,
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

	// 从持久化存储上恢复 agent 列表
	gs.restoreAgents()

	r := gin.Default()
	v2 := r.Group("/v2")
	{
		v2.GET("/cover/profile", gs.getProfiles)
		v2.DELETE("/cover/profile", gs.resetProfiles)
		v2.GET("/agents", gs.listAgents)
		v2.DELETE("/agents", gs.removeAgents)

		v2.GET("/cover/ws/watch", gs.watchProfileUpdate)

		// internal use only
		v2.GET("/internal/register", gs.register)
		v2.GET("/internal/ws/rpcstream", gs.serveRpcStream)
		v2.GET("/internal/ws/watchstream", gs.serveWatchInternalStream)
	}

	go gs.watchLoop()

	return r.Run(host)
}

func (gs *gocServer) register(c *gin.Context) {
	// 检查插桩服务上报的信息
	hostname := c.Query("hostname")
	pid := c.Query("pid")
	cmdline := c.Query("cmdline")
	extra := c.Query("extra")

	if hostname == "" || pid == "" || cmdline == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "missing some params",
		})
		return
	}

	gs.idL.Lock()
	gs.idCount++
	globalId := gs.idCount
	gs.idL.Unlock()

	genToken := func(i int64) string {
		now := time.Now().UnixNano()
		random := rand.Int()

		raw := fmt.Sprintf("%v-%v-%v", i, random, now)
		sum := sha256.Sum256([]byte(raw))
		h := fmt.Sprintf("%x", sum[:16])

		return h
	}

	token := genToken(globalId)
	id := strconv.Itoa(int(globalId))

	agent := &gocCoveredAgent{
		Id:       id,
		Hostname: hostname,
		Pid:      pid,
		CmdLine:  cmdline,
		Token:    token,
		Status:   DISCONNECT,
		Extra:    extra,
	}

	// 持久化
	err := gs.saveAgentToStore(agent)
	if err != nil {
		log.Errorf("fail to save to store: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
	}
	// 维护 agent 连接
	gs.agents.Store(id, agent)

	log.Infof("one agent registered, id: %v, cmdline: %v, pid: %v, hostname: %v", id, agent.CmdLine, agent.Pid, agent.Hostname)

	c.JSON(http.StatusOK, gin.H{
		"id":    id,
		"token": token,
	})
}

func (gs *gocServer) saveAgentToStore(agent *gocCoveredAgent) error {

	value, err := json.Marshal(agent)
	if err != nil {
		return err
	}
	return gs.store.Set("/goc/agents/"+agent.Id, string(value))
}

func (gs *gocServer) removeAgentFromStore(id string) error {

	return gs.store.Remove("/goc/agents/" + id)
}

func (gs *gocServer) removeAllAgentsFromStore() error {

	return gs.store.RangeRemove("/goc/agents/")
}

func (gs *gocServer) restoreAgents() {
	pattern := "/goc/agents/"

	// ignore err, 这个 err 不需要处理，直接忽略
	rawagents, _ := gs.store.Range(pattern)

	var maxId int
	for _, rawagent := range rawagents {
		var agent gocCoveredAgent
		err := json.Unmarshal([]byte(rawagent), &agent)
		if err != nil {
			log.Fatalf("fail to unmarshal restore agents: %v", err)
		}

		id, err := strconv.Atoi(agent.Id)
		if err != nil {
			log.Fatalf("fail to transform id to number: %v", err)
		}
		if maxId < id {
			maxId = id
		}

		gs.agents.Store(agent.Id, &agent)
		log.Infof("restore one agent: %v, %v from store", id, agent.RpcRemoteIP)

		agent.RpcRemoteIP = ""
		agent.WatchRemoteIP = ""
		agent.Status = DISCONNECT
	}

	// 更新全局 id
	atomic.StoreInt64(&gs.idCount, int64(maxId))
}
