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

package qiniu

import (
	"os/exec"
	"strings"
)

// Command create new command
func Command(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// ShellCommand create new command run with shell
func ShellCommand(name string, args ...string) *exec.Cmd {
	param := make([]string, len(args)+1)
	param[0] = name
	copy(param[1:], args)
	return exec.Command(Shell, OptCmd, strings.Join(param, " "))
}

// Artifact return the build artifact name
func Artifact(name string) string {
	return name + ArtifactExt
}

// ToSlash Cast path from os style to golang 'import' code style
func ToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
