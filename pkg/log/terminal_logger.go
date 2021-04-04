package log

import (
	"fmt"
	"io"
	"os"
	"sync"

	goansi "github.com/k0kubun/go-ansi"
	"github.com/mgutz/ansi"
	"go.uber.org/zap/zapcore"
)

// goansi works nicer on Windows platform
var stdout = goansi.NewAnsiStdout()
var stderr = goansi.NewAnsiStderr()

type terminalLogger struct {
	mutex       sync.Mutex
	level       zapcore.Level
	loadingText *loadingText
}

type levelFuncType int32

const (
	fatalFn levelFuncType = iota
	infoFn
	errorFn
	warnFn
	debugFn
	doneFn
)

type levelFuncInfo struct {
	tag    string
	color  string
	level  zapcore.Level
	stream io.Writer
}

var levelFuncMap = map[levelFuncType]*levelFuncInfo{
	doneFn: {
		tag:    "[done] √ ",
		color:  "green+b",
		level:  zapcore.InfoLevel,
		stream: stdout,
	},
	debugFn: {
		tag:    "[debug]  ",
		color:  "green+b",
		level:  zapcore.DebugLevel,
		stream: stdout,
	},
	infoFn: {
		tag:    "[info]   ",
		color:  "cyan+b",
		level:  zapcore.InfoLevel,
		stream: stdout,
	},
	warnFn: {
		tag:    "[warn]   ",
		color:  "magenta+b",
		level:  zapcore.WarnLevel,
		stream: stdout,
	},
	errorFn: {
		tag:    "[error]  ",
		color:  "yellow+b",
		level:  zapcore.ErrorLevel,
		stream: stdout,
	},
	fatalFn: {
		tag:    "[fatal]  ",
		color:  "red+b",
		level:  zapcore.FatalLevel,
		stream: stdout,
	},
}

func (t *terminalLogger) writeMessage(funcType levelFuncType, message string) {
	funcInfo := levelFuncMap[funcType]
	if t.level <= funcInfo.level {
		// 如果当前有消息在加载，需先暂停
		if t.loadingText != nil {
			t.loadingText.stop()
		}

		funcInfo.stream.Write([]byte(ansi.Color(funcInfo.tag, funcInfo.color)))
		funcInfo.stream.Write([]byte(message))

		// 恢复加载
		if t.loadingText != nil && funcType != fatalFn {
			t.loadingText.start()
		}
	}
}

// StartWait prints a waiting message until StopWait is called
func (t *terminalLogger) StartWait(message string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// 撤销之前的加载
	if t.loadingText != nil {
		t.loadingText.stop()
		t.loadingText = nil
	}

	// 创建新的加载字符串
	t.loadingText = &loadingText{
		message: message,
		stream:  goansi.NewAnsiStdout(),
	}

	t.loadingText.start()
}

// StopWait stops waiting
func (t *terminalLogger) StopWait() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.loadingText != nil {
		t.loadingText.stop()
		t.loadingText = nil
	}
}

func (t *terminalLogger) Sync() {

}

func (t *terminalLogger) Debugf(format string, args ...interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.writeMessage(debugFn, fmt.Sprintf(format, args...)+"\n")
}

func (t *terminalLogger) Donef(format string, args ...interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.writeMessage(doneFn, fmt.Sprintf(format, args...)+"\n")
}

func (t *terminalLogger) Infof(format string, args ...interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.writeMessage(infoFn, fmt.Sprintf(format, args...)+"\n")
}

func (t *terminalLogger) Errorf(format string, args ...interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.writeMessage(errorFn, fmt.Sprintf(format, args...)+"\n")
}

func (t *terminalLogger) Warnf(format string, args ...interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.writeMessage(warnFn, fmt.Sprintf(format, args...)+"\n")
}

func (t *terminalLogger) Fatalf(format string, args ...interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.writeMessage(fatalFn, fmt.Sprintf(format, args...)+"\n")

	os.Exit(1)
}
