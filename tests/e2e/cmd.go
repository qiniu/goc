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

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LongRunCmd defines a cmd which run for a long time
type LongRunCmd struct {
	cancel    context.CancelFunc
	cmd       *exec.Cmd
	stderrBuf bytes.Buffer
	stdoutBuf bytes.Buffer
	err       error
	done      bool

	once sync.Once
}

// NewLongRunCmd defines a command which will be run forever
//
// You can specify the whole command arg list, the directory where the command
// will be run in, and the env list.
//
// args 命令列表
//
// dir 命令运行所在的目录
//
// envs 额外的环境变量
func NewLongRunCmd(args []string, dir string, envs []string) *LongRunCmd {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, args[0]+".exe", args[1:]...)
	}
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), envs...)

	var stderrBuf bytes.Buffer
	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	return &LongRunCmd{
		cmd:       cmd,
		cancel:    cancel,
		stderrBuf: stderrBuf,
		stdoutBuf: stdoutBuf,
	}
}

// Run in backend
func (l *LongRunCmd) Run() {
	go func() {
		err := l.cmd.Start()
		if err != nil {
			l.err = err
		}

		err = l.cmd.Wait()
		if err != nil {
			l.err = err
		}
		l.err = nil
		l.done = true
	}()
	time.Sleep(time.Millisecond * 100)
}

func (l *LongRunCmd) Stop() {
	l.once.Do(func() {
		l.cancel()
	})
}

func (l *LongRunCmd) CheckExitStatus() error {
	if l.done == true {
		return l.err
	} else {
		return fmt.Errorf("running")
	}
}

func (l *LongRunCmd) GetStdoutStdErr() (string, string) {
	return l.stdoutBuf.String(), l.stderrBuf.String()
}

// RunShortRunCmd defines a cmd which run and exits immediately
//
// args 命令列表
//
// dir 命令运行所在的目录
//
// envs 额外的环境变量
func RunShortRunCmd(args []string, dir string, envs []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, args[0]+".exe", args[1:]...)
	}
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), envs...)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// SearchSymbolInBinary searches if the specified symbol is compiled into the bianry
func SearchSymbolInBinary(dir string, binary string, symbol string) (bool, error) {
	if runtime.GOOS == "windows" {
		binary = binary + ".exe"
	}
	output, err := RunShortRunCmd([]string{"go", "tool", "objdump", "-s", symbol, binary}, dir, nil)
	if err != nil {
		return false, fmt.Errorf("cannot lookup into the binary: %w", err)
	}

	return strings.Contains(output, symbol), nil
}
