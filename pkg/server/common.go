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
	"io"
	"time"

	"github.com/gorilla/websocket"
)

const (
	PongWait = 10 * time.Second
	PingWait = 5 * time.Second
)

type ProfileReq string

type ProfileRes string

func (gs *gocServer) wsping(ws *websocket.Conn, deadline time.Duration) error {
	return ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(deadline))
}

func (gs *gocServer) wsclose(ws *websocket.Conn, deadline time.Duration) error {
	return ws.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(deadline))
}

type ReadWriteCloser struct {
	ws *websocket.Conn
	r  io.Reader
	w  io.WriteCloser
}

func (rwc *ReadWriteCloser) Read(p []byte) (n int, err error) {
	if rwc.r == nil {
		var _ int
		_, rwc.r, err = rwc.ws.NextReader()
		if err != nil {
			return 0, err
		}
	}
	for n = 0; n < len(p); {
		var m int
		m, err = rwc.r.Read(p[n:])
		n += m
		if err == io.EOF {
			// done
			rwc.r = nil
			break
		}
		// ???
		if err != nil {
			break
		}
	}
	return
}

func (rwc *ReadWriteCloser) Write(p []byte) (n int, err error) {
	if rwc.w == nil {
		rwc.w, err = rwc.ws.NextWriter(websocket.TextMessage)
		if err != nil {
			return 0, err
		}
	}
	for n = 0; n < len(p); {
		var m int
		m, err = rwc.w.Write(p)
		n += m
		if err != nil {
			break
		}
	}
	if err != nil || n == len(p) {
		err = rwc.Close()
	}
	return
}

func (rwc *ReadWriteCloser) Close() (err error) {
	if rwc.w != nil {
		err = rwc.w.Close()
		rwc.w = nil
	}
	return err
}
