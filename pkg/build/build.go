package build

import (
	"github.com/qiniu/goc/v2/pkg/flag"
	"github.com/qiniu/goc/v2/pkg/log"
	"github.com/spf13/cobra"
)

// Build struct a build
// most configurations are stored in global variables: config.GocConfig & config.GoConfig
type Build struct {
}

// NewBuild creates a Build struct
//
// consumes args, get package dirs, read project meta info.
func NewBuild(cmd *cobra.Command, args []string) *Build {
	b := &Build{}
	remainedArgs := flag.BuildCmdArgsParse(cmd, args)
	flag.GetPackagesDir(remainedArgs)
	b.readProjectMetaInfo()
	b.displayProjectMetaInfo()

	return b
}

// Build starts go build
//
// 1. copy project to temp,
// 2. inject cover variables and functions into the project,
// 3. build the project in temp.
func (b *Build) Build() {
	b.copyProjectToTmp()
	defer b.clean()
	log.Donef("project copied to temporary directory")
}
