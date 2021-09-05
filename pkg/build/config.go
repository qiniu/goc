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

package build

import (
	"github.com/spf13/pflag"
)

type gocOption func(*Build)

func WithHost(host string) gocOption {
	return func(b *Build) {
		b.Host = host
	}
}

func WithMode(mode string) gocOption {
	return func(b *Build) {
		b.Mode = mode
	}
}

func WithArgs(args []string) gocOption {
	return func(b *Build) {
		b.Args = args
	}
}

func WithFlagSets(sets *pflag.FlagSet) gocOption {
	return func(b *Build) {
		b.FlagSets = sets
	}
}

func WithBuild() gocOption {
	return func(b *Build) {
		b.BuildType = 0
	}
}

func WithInstall() gocOption {
	return func(b *Build) {
		b.BuildType = 1
	}
}

func WithDebug(enable bool) gocOption {
	return func(b *Build) {
		b.Debug = enable
	}
}
