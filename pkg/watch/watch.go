package watch

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/log"
)

func Watch() {
	watchUrl := fmt.Sprintf("ws://%v/v2/cover/ws/watch", config.GocConfig.Host)
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
