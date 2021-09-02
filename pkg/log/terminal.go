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

package log

import (
	"os"

	"k8s.io/kubectl/pkg/util/term"
)

func setupTTY() term.TTY {
	t := term.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
	}

	if !t.IsTerminalIn() {
		return t
	}

	t.Raw = true

	return t
}
