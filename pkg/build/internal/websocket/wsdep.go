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

package websocket

import (
	"archive/tar"
	"bytes"
	"embed"
	"io"
	"os"
	"path/filepath"

	"github.com/qiniu/goc/v2/pkg/log"
)

//go:embed websocket.tar
var depTarFile embed.FS

// AddCustomWebsocketDep injects custom gorrila/websocket library into the temporary directory
//
// 从 embed 文件系统中解压 websocket.tar 文件，并依次写入临时工程中，作为一个单独的包存在。
// gorrila/websocket 是一个无第三方依赖的库，因此其位置可以随处移动，而不影响自身的编译。
func AddCustomWebsocketDep(customWebsocketPath string) {
	data, err := depTarFile.ReadFile("websocket.tar")
	if err != nil {
		log.Fatalf("cannot find the websocket.tar in the embed file: %v", err)
	}

	buf := bytes.NewBuffer(data)
	tr := tar.NewReader(buf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("cannot untar the websocket.tar: %v", err)
		}

		fpath := filepath.Join(customWebsocketPath, hdr.Name)
		if hdr.FileInfo().IsDir() {
			// 处理目录
			err := os.MkdirAll(fpath, hdr.FileInfo().Mode())
			if err != nil {
				log.Fatalf("fail to untar the websocket.tar: %v", err)
			}
		} else {
			// 处理文件
			fdir := filepath.Dir(fpath)
			err := os.MkdirAll(fdir, hdr.FileInfo().Mode())
			if err != nil {
				log.Fatalf("fail to untar the websocket.tar: %v", err)
			}

			f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				log.Fatalf("fail to untar the websocket.tar: %v", err)
			}
			defer f.Close()

			_, err = io.Copy(f, tr)

			if err != nil {
				log.Fatalf("fail to untar the websocket.tar: %v", err)
			}
		}
	}
}
