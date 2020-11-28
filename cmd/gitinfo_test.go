package cmd

import (
	"bytes"
	"fmt"
	"github.com/qiniu/goc/pkg/cover"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)


var bin = "package main\n\nimport (\n\t\"fmt\"\n\t\"net/http\"\n)\n\nfunc sayhelloName(w http.ResponseWriter, r *http.Request) {\n\tfmt.Fprintf(w, \"Hello astaxie!\") \n}\n\nfunc main() {\n\thttp.HandleFunc(\"/\", sayhelloName) \n\terr := http.ListenAndServe(\":9991\", nil) \n\tif err != nil {\n\t}\n}"

type gitInfoFunc func(cmd *cobra.Command, args []string)

var ids []int

func GetRunMainDir() *string {
	appName := "goc"
	cfgDir := ""
	if dir, err := os.Getwd(); err != nil {
		log.Fatal(err)
	}else {
		log.Println(dir)
		cfgDir = strings.Split(dir,appName)[0] + appName
	}
	return &cfgDir
}


func doGitInfoLocal(cmd *cobra.Command, args []string)  {
	cmdStr := "sh -c go build ."
	runDir := GetRunMainDir()

	if info, err, _ := cover.RunCommandStr(*runDir, cmdStr); err != nil {
		log.Println(err)
		log.Println(info)
	}

	dir := "TMP"
	mainDir := path.Join( *runDir,dir)
	defer func() {
		if err  := os.RemoveAll(mainDir) ; err != nil {
			log.Println(err)
		}
		for pid := range ids {
			cmdStr := fmt.Sprintf( "kill -9  %d",pid)
			_, _, _ = cover.RunCommandStr(*runDir, cmdStr)
		}
	}()
	if err := os.MkdirAll(mainDir,0777) ; err != nil {
		log.Println(err)
	}
	testMain := path.Join( mainDir,"main.go")
	err := ioutil.WriteFile(testMain, []byte( args[0] ), 0777)
	if err != nil {
		log.Println(err)
	}


	cmdStr = "sh -c  cp ../goc  . "
	if info, err,_ := cover.RunCommandStr(mainDir, cmdStr); err != nil {
		log.Println(err)
		log.Println(info)
	}

	cmdStr = "sh -c  ./goc build   --agentport :9090 . "
	if info, err,_ := cover.RunCommandStr(mainDir, cmdStr); err != nil {
		log.Println(err)
		log.Println(info)
	}

	go func() {
		cmdStr = "sh -c   ./goc server  "
		if info, err, pidServer  := cover.RunCommandStr(mainDir, cmdStr); err != nil {
			log.Println(err)
			log.Println(info)
		}else{
			log.Println(info)
			log.Println("pidServer",pidServer)
			ids = append(ids,pidServer)
		}
	}()

	go func() {
		cmdStr = "sh -c   ./TMP & "
		if _, err,pidMain := cover.RunCommandStr(mainDir, cmdStr); err != nil {
			log.Println(err)
		}else {
			log.Println("pidMain",pidMain)
			ids = append(ids,pidMain)
		}
	}()

	time.Sleep(3*time.Second)
	cmdStr = "sh -c  ./goc gitinfo --address http://127.0.0.1:9090  "
	if info, err ,_:= cover.RunCommandStr(mainDir, cmdStr); err != nil {
		log.Println(err)
		log.Println(info)
	}else {
		//log.Println(info)
		fmt.Println(info)
	}
}

func captureGitInfoStdout(f gitInfoFunc, cmd *cobra.Command, args []string) string {
	r, w, err := os.Pipe()
	if err != nil {
		logrus.WithError(err).Fatal("os pipe fail")
	}
	stdout := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = stdout
	}()

	f(cmd, args)
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func TestDoGitInfo(t *testing.T) {

	log.SetFlags(log.Lshortfile | log.LstdFlags)
	_ = gitInfoCmd.Flags().Set("service", "")
	_ = gitInfoCmd.Flags().Set("address", "")
	_ = gitInfoCmd.Flags().Set("path", ".")
	//gitInfoCmd.Run()
	out := captureGitInfoStdout( doGitInfoLocal , gitInfoCmd, []string{bin})
	log.Println(out)
	assert.NotNil(t, out)

}