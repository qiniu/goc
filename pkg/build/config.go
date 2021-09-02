package build

import (
	"github.com/spf13/pflag"
)

type GocOption func(*Build)

func WithHost(host string) GocOption {
	return func(b *Build) {
		b.Host = host
	}
}

func WithMode(mode string) GocOption {
	return func(b *Build) {
		b.Mode = mode
	}
}

func WithArgs(args []string) GocOption {
	return func(b *Build) {
		b.Args = args
	}
}

func WithFlagSets(sets *pflag.FlagSet) GocOption {
	return func(b *Build) {
		b.FlagSets = sets
	}
}

func WithBuild() GocOption {
	return func(b *Build) {
		b.BuildType = 0
	}
}

func WithInstall() GocOption {
	return func(b *Build) {
		b.BuildType = 1
	}
}

func WithDebug(enable bool) GocOption {
	return func(b *Build) {
		b.Debug = enable
	}
}
