package build

import (
	"errors"
)

var (
	// ErrShouldNotReached represents the logic should not be reached in normal flow
	ErrShouldNotReached = errors.New("should never be reached")
	// ErrGocShouldExecInProject represents goc currently not support for the project
	ErrGocShouldExecInProject = errors.New("goc not support for such project directory")
	// ErrWrongPackageTypeForInstall represents goc install command only support limited arguments
	ErrWrongPackageTypeForInstall = errors.New("packages only support \".\" and \"./...\"")
	// ErrWrongPackageTypeForBuild represents goc build command only support limited arguments
	ErrWrongPackageTypeForBuild = errors.New("packages only support \".\"")
	// ErrTooManyArgs represents goc CLI only support limited arguments
	ErrTooManyArgs = errors.New("too many args")
	// ErrInvalidWorkingDir represents the working directory is invalid
	ErrInvalidWorkingDir = errors.New("the working directory is invalid")
	// ErrEmptyTempWorkingDir represent the error that temporary working directory is empty
	ErrEmptyTempWorkingDir = errors.New("temporary working directory is empty")
	// ErrNoPlaceToInstall represents the err that no place to install the generated binary
	ErrNoPlaceToInstall = errors.New("don't know where to install")
)
