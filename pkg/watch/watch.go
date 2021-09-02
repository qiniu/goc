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

package watch

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/qiniu/goc/v2/pkg/log"
)

func Watch(host string) {
	watchUrl := fmt.Sprintf("ws://%v/v2/cover/ws/watch", host)
	c, _, err := websocket.DefaultDialer.Dial(watchUrl, nil)
	if err != nil {
		log.Fatalf("cannot connect to goc server: %v", err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatalf("cannot read message: %v", err)
		}

		log.Infof("profile update: %v", string(message))
	}
}
