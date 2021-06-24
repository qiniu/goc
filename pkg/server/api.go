package server

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/goc/v2/pkg/log"
	"golang.org/x/tools/cover"
	"k8s.io/test-infra/gopherage/pkg/cov"
)

// listAgents return all service informations
func (gs *gocServer) listAgents(c *gin.Context) {
	agents := make([]*gocCoveredAgent, 0)

	gs.rpcAgents.Range(func(key, value interface{}) bool {
		agent, ok := value.(*gocCoveredAgent)
		if !ok {
			return false
		}
		agents = append(agents, agent)
		return true
	})

	c.JSON(http.StatusOK, gin.H{
		"items": agents,
	})
}

// getProfiles get and merge all agents' informations
//
// it is synchronous
func (gs *gocServer) getProfiles(c *gin.Context) {
	var mu sync.Mutex
	var wg sync.WaitGroup

	mergedProfiles := make([][]*cover.Profile, 0)

	gs.rpcAgents.Range(func(key, value interface{}) bool {
		agent, ok := value.(*gocCoveredAgent)
		if !ok {
			return false
		}

		wg.Add(1)
		// 并发 rpc，且每个 rpc 设超时时间 10 second
		go func() {
			defer wg.Done()

			timeout := time.Duration(10 * time.Second)
			done := make(chan error, 1)

			var req ProfileReq = "getprofile"
			var res ProfileRes
			go func() {
				err := agent.rpc.Call("GocAgent.GetProfile", req, &res)
				if err != nil {
					log.Errorf("fail to get profile from: %v, reasson: %v. let's close the connection", agent.Id, err)
				}
				done <- err
			}()

			select {
			// rpc 超时
			case <-time.After(timeout):
				log.Warnf("rpc call timeout: %v", agent.Hostname)
				// 关闭链接
				agent.once.Do(func() {
					close(agent.exitCh)
				})
			case err := <-done:
				// 调用 rpc 发生错误
				if err != nil {
					// 关闭链接
					agent.once.Do(func() {
						close(agent.exitCh)
					})
				}
			}
			// append profile
			profile, err := convertProfile([]byte(res))
			if err != nil {
				log.Errorf("fail to convert the received profile from: %v, reasson: %v. let's close the connection", agent.Id, err)
				// 关闭链接
				agent.once.Do(func() {
					close(agent.exitCh)
				})
				return
			}
			mu.Lock()
			defer mu.Unlock()
			mergedProfiles = append(mergedProfiles, profile)
		}()

		return true
	})

	// 一直等待并发的 rpc 都回应
	wg.Wait()

	merged, err := cov.MergeMultipleProfiles(mergedProfiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}

	var buff bytes.Buffer
	err = cov.DumpProfile(merged, &buff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": buff.String(),
	})
}

// resetProfiles reset all profiles in agent
//
// it is async, the function will return immediately
func (gs *gocServer) resetProfiles(c *gin.Context) {
	gs.rpcAgents.Range(func(key, value interface{}) bool {
		agent, ok := value.(gocCoveredAgent)
		if !ok {
			return false
		}

		var req ProfileReq = "resetprofile"
		var res ProfileRes
		go func() {
			err := agent.rpc.Call("GocAgent.ResetProfile", req, &res)
			if err != nil {
				log.Errorf("fail to reset profile from: %v, reasson: %v. let's close the connection", agent.Id, err)
				// 关闭链接
				agent.once.Do(func() {
					close(agent.exitCh)
				})
			}
		}()

		return true
	})
}

// watchProfileUpdate watch the profile change
//
// any profile change will be updated on this websocket connection.
func (gs *gocServer) watchProfileUpdate(c *gin.Context) {
	// upgrade to websocket
	ws, err := gs.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("fail to establish websocket connection with watch client: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
	}

	log.Infof("watch client connected")

	id := time.Now().String()
	gwc := &gocWatchClient{
		ws:     ws,
		exitCh: make(chan int),
	}
	gs.watchClients.Store(id, gwc)
	// send close msg and close ws connection
	defer func() {
		gs.watchClients.Delete(id)
		ws.Close()
		gwc.once.Do(func() { close(gwc.exitCh) })
		log.Infof("watch client disconnected")
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
				break
			}
		}

		gwc.once.Do(func() { close(gwc.exitCh) })
	}()

	<-gwc.exitCh
}
