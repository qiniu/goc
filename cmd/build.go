package cmd

import (
	"time"

	"github.com/qiniu/goc/v2/pkg/log"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use: "build",
	Run: func(cmd *cobra.Command, args []string) {
		log.StartWait("doing something")
		time.Sleep(time.Second * 3)
		log.Infof("building")
		time.Sleep(time.Second * 3)
		log.Infof("making temp dir")
		time.Sleep(time.Second * 3)
		log.StopWait()
		log.Donef("done")
		log.Infof("hello")
		log.Errorf("hello")
		log.Warnf("hello")
		log.Debugf("hello")
		log.Fatalf("fail to excute: %v, %v", "ee", "ddd")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
