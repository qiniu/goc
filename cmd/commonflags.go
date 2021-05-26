/*
 Copyright 2020 Qiniu Cloud (qiniu.com)

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

package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	target            string
	center            string
	agentPort         AgentPort
	debugGoc          bool
	debugInCISyncFile string
	buildFlags        string
	singleton         bool

	goRunExecFlag  string
	goRunArguments string
)

var coverMode = CoverMode{
	mode: "count",
}

// addBasicFlags adds a
func addBasicFlags(cmdset *pflag.FlagSet) {
	cmdset.StringVar(&center, "center", "http://127.0.0.1:7777", "cover profile host center")
	// bind to viper
	viper.BindPFlags(cmdset)
}

func addCommonFlags(cmdset *pflag.FlagSet) {
	addBasicFlags(cmdset)
	cmdset.Var(&coverMode, "mode", "coverage mode: set, count, atomic")
	cmdset.Var(&agentPort, "agentport", "a fixed port such as :8100 for registered service communicate with goc server. if not provided, using a random one")
	cmdset.BoolVar(&singleton, "singleton", false, "singleton mode, not register to goc center")
	cmdset.StringVar(&buildFlags, "buildflags", "", "specify the build flags")
	// bind to viper
	viper.BindPFlags(cmdset)
}

func addBuildFlags(cmdset *pflag.FlagSet) {
	addCommonFlags(cmdset)
	// bind to viper
	viper.BindPFlags(cmdset)
}

func addRunFlags(cmdset *pflag.FlagSet) {
	addBuildFlags(cmdset)
	cmdset.StringVar(&goRunExecFlag, "exec", "", "same as -exec flag in 'go run' command")
	cmdset.StringVar(&goRunArguments, "arguments", "", "same as 'arguments' in 'go run' command")
	// bind to viper
	viper.BindPFlags(cmdset)
}

// CoverMode represents the covermode when doing cover for source code
type CoverMode struct {
	mode string
}

func (m *CoverMode) String() string {
	return m.mode
}

// Set sets the value to the CoverMode struct, use 'count' as default if v is empty
func (m *CoverMode) Set(v string) error {
	if v == "" {
		m.mode = "count"
		return nil
	}
	if v != "set" && v != "count" && v != "atomic" {
		return fmt.Errorf("unknown mode")
	}
	m.mode = v
	return nil
}

// Type returns the type of CoverMode
func (m *CoverMode) Type() string {
	return "string"
}

// AgentPort is the struct to do agentPort check
type AgentPort struct {
	port string
}

func (agent *AgentPort) String() string {
	return agent.port
}

// Set sets the value to the AgentPort struct
func (agent *AgentPort) Set(v string) error {
	if v == "" {
		agent.port = ""
		return nil
	}
	_, _, err := net.SplitHostPort(v)
	if err != nil {
		return err
	}
	agent.port = v
	return nil
}

// Type returns the type of AgentPort
func (agent *AgentPort) Type() string {
	return "string"
}
