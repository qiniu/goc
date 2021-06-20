package websocket

import (
	"archive/tar"
	"bytes"
	"embed"
	"io"
	"os"
	"path/filepath"

	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/log"
)

//go:embed websocket.tar
var depTarFile embed.FS

// AddCustomWebsocketDep injects custom gorrila/websocket library into the temporary directory
//
// 从 embed 文件系统中解压 websocket.tar 文件，并依次写入临时工程中，作为一个单独的包存在。
// gorrila/websocket 是一个无第三方依赖的库，因此其位置可以随处移动，而不影响自身的编译。
func AddCustomWebsocketDep() {
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

		customWebsocketPath := config.GocConfig.GlobalCoverVarImportPathDir
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
