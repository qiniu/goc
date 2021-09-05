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

package cover

import (
	"fmt"
	"time"
	"os"
	"io/ioutil"
	"log"
	"net/url"

	"{{.GlobalCoverVarImportPath}}/websocket"

	_cover "{{.GlobalCoverVarImportPath}}"
)

func init() {
	// init host
	host_env := os.Getenv("GOC_CUSTOM_HOST")
	if host_env != "" {
		host = host_env
	}

	var dialer = websocket.DefaultDialer

	go func() {
		cond.L.Lock()
		cond.Wait()
		cond.L.Unlock()

		for {			
			// 直接将 token 放在 ws 地址中
			v := url.Values{}
			v.Set("token", token)
			v.Set("id", id)
			v.Encode()

			watchstreamUrl := fmt.Sprintf("ws://%v/v2/internal/ws/watchstream?%v", host, v.Encode())
			ws, resp, err := dialer.Dial(watchstreamUrl, nil)
			if err != nil {
				if resp != nil {
					tmp, _ := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					log.Printf("[goc][Error] watch fail to dial to goc server: %v, body: %v", err, string(tmp))
				} else {
					log.Printf("[goc][Error] watch fail to dial to goc server: %v", err)
				}
				time.Sleep(waitDelay)
				continue
			}

			// 连接成功
			_cover.WatchEnabled_{{.Random}} = true
			log.Printf("[goc][Info] watch connected to goc server")

			ticker := time.NewTicker(time.Second)
			closeFlag := false
			go func() {
				for {
					// 必须调用一下以触发 ping 的自动处理
					_, _, err := ws.ReadMessage()
					if err != nil {
						break
					}
				}
				closeFlag = true
			}()

			Loop:
			for {
				select {
				case block := <-_cover.WatchChannel_{{.Random}}:
					i := block.I

					cov := fmt.Sprintf("%s:%d.%d,%d.%d %d %d", block.Name,
						block.Pos[3*i+0], uint16(block.Pos[3*i+2]),
						block.Pos[3*i+1], uint16(block.Pos[3*i+2] >> 16),
						block.Stmts,
						1)

					err = ws.WriteMessage(websocket.TextMessage, []byte(cov))
					if err != nil {
						_cover.WatchEnabled_{{.Random}} = false
						log.Println("[goc][Error] push coverage failed: %v", err)
						time.Sleep(waitDelay)
						break Loop
					}
				case <-ticker.C:
					if closeFlag == true {
						break Loop
					}
				}
			}
		}
	}()
}
