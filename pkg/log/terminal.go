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
