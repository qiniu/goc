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
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/goc/v2/pkg/log"
	"golang.org/x/tools/cover"
	"k8s.io/test-infra/gopherage/pkg/cov"
)

// listAgents return all service informations
func (gs *gocServer) listAgents(c *gin.Context) {
	idQuery := c.Query("id")
	ifInIdMap := idMaps(idQuery)

	agents := make([]*gocCoveredAgent, 0)

	gs.agents.Range(func(key, value interface{}) bool {
		// check if id is in the query ids
		if !ifInIdMap(key.(string)) {
			return true
		}

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

func (gs *gocServer) removeAgents(c *gin.Context) {
	idQuery := c.Query("id")
	ifInIdMap := idMaps(idQuery)

	errs := ""
	gs.agents.Range(func(key, value interface{}) bool {

		// check if id is in the query ids
		id := key.(string)
		if !ifInIdMap(id) {
			return true
		}

		agent, ok := value.(*gocCoveredAgent)
		if !ok {
			return false
		}

		err := gs.removeAgentFromStore(id)
		if err != nil {
			log.Errorf("fail to remove agent: %v", id)
			err := fmt.Errorf("fail to remove agent: %v, err: %v", id, err)
			errs = errs + err.Error()
			return true
		}
		agent.closeConnection()
		gs.agents.Delete(key)

		return true
	})

	if errs != "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": errs,
		})
	} else {
		c.JSON(http.StatusOK, nil)
	}
}

// getProfiles get and merge all agents' informations
//
// it is synchronous
func (gs *gocServer) getProfiles(c *gin.Context) {
	idQuery := c.Query("id")
	ifInIdMap := idMaps(idQuery)

	skippatternRaw := c.Query("skippattern")
	var skippattern []string
	if skippatternRaw != "" {
		skippattern = strings.Split(skippatternRaw, ",")
	}

	extra := c.Query("extra")
	isExtra := filterExtra(extra)

	var mu sync.Mutex
	var wg sync.WaitGroup

	mergedProfiles := make([][]*cover.Profile, 0)

	gs.agents.Range(func(key, value interface{}) bool {
		// check if id is in the query ids
		if !ifInIdMap(key.(string)) {
			// not in
			return true
		}

		agent, ok := value.(*gocCoveredAgent)
		if !ok {
			return false
		}

		// check if extra matches
		if !isExtra(agent.Extra) {
			// not match
			return true
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
				// lock-free
				rpc := agent.rpc
				if rpc == nil || agent.Status == DISCONNECT {
					done <- nil
					return
				}
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
				agent.closeRpcConnOnce()
			case err := <-done:
				// 调用 rpc 发生错误
				if err != nil {
					// 关闭链接
					agent.closeRpcConnOnce()
				}
			}
			// append profile
			profile, err := convertProfile([]byte(res))
			if err != nil {
				log.Errorf("fail to convert the received profile from: %v, reasson: %v. let's close the connection", agent.Id, err)
				// 关闭链接
				agent.closeRpcConnOnce()
				return
			}

			// check if skippattern matches
			newProfile := filterProfileByPattern(skippattern, profile)

			mu.Lock()
			defer mu.Unlock()
			mergedProfiles = append(mergedProfiles, newProfile)
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
	idQuery := c.Query("id")
	ifInIdMap := idMaps(idQuery)

	extra := c.Query("extra")
	isExtra := filterExtra(extra)

	gs.agents.Range(func(key, value interface{}) bool {

		// check if id is in the query ids
		if !ifInIdMap(key.(string)) {
			// not in
			return true
		}

		agent, ok := value.(*gocCoveredAgent)
		if !ok {
			return false
		}

		// check if extra matches
		if !isExtra(agent.Extra) {
			// not match
			return true
		}

		var req ProfileReq = "resetprofile"
		var res ProfileRes
		go func() {
			// lock-free
			rpc := agent.rpc
			if rpc == nil || agent.Status == DISCONNECT {
				return
			}
			err := rpc.Call("GocAgent.ResetProfile", req, &res)
			if err != nil {
				log.Errorf("fail to reset profile from: %v, reasson: %v. let's close the connection", agent.Id, err)
				// 关闭链接
				agent.closeRpcConnOnce()
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

func filterProfileByPattern(skippattern []string, profiles []*cover.Profile) []*cover.Profile {

	if len(skippattern) == 0 {
		return profiles
	}

	var out = make([]*cover.Profile, 0)
	for _, profile := range profiles {
		skip := false
		for _, pattern := range skippattern {
			if strings.Contains(profile.FileName, pattern) {
				skip = true
				break
			}
		}

		if !skip {
			out = append(out, profile)
		}
	}

	return out
}

func idMaps(idQuery string) func(key string) bool {
	idMap := make(map[string]bool)
	if len(strings.TrimSpace(idQuery)) == 0 {
	} else {
		ids := strings.Split(idQuery, ",")
		for _, id := range ids {
			idMap[id] = true
		}
	}

	inIdMaps := func(key string) bool {
		// if no id in query, then all id agent will be return
		if len(idMap) == 0 {
			return true
		}
		// other
		_, ok := idMap[key]
		if !ok {
			return false
		} else {
			return true
		}
	}

	return inIdMaps
}

func filterExtra(extraPattern string) func(string) bool {

	re := regexp.MustCompile(extraPattern)

	return func(extra string) bool {
		return re.Match([]byte(extra))
	}
}
