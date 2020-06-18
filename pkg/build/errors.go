package build

import (
	"errors"
)

var (
	ErrShouldNotReached           = errors.New("should never be reached")
	ErrGocShouldExecInProject     = errors.New("goc not executed in project directory")
	ErrWrongPackageTypeForInstall = errors.New("packages only support \".\" and \"./...\"")
	ErrWrongPackageTypeForBuild   = errors.New("packages only support \".\"")
	ErrTooManyArgs                = errors.New("too many args")
	ErrWrongCallSequence          = errors.New("function should be called in a specified sequence")
	ErrNoplaceToInstall           = errors.New("no go env")
)
