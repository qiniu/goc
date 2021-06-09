package cover

import (
	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/log"
)

// Inject injects cover variables for all the .go files in the target directory
func Inject() {
	log.StartWait("injecting cover variables")

	// var seen := make(map[string]*PackageCover)

	for _, pkg := range config.GocConfig.Pkgs {
		if pkg.Name == "main" {
			log.Infof("handle package: %v", pkg.ImportPath)
		}
	}
	log.StopWait()
	log.Donef("cover variables injected")
}

// declareCoverVars attaches the required cover variables names
// to the files, to be used when annotating the files.
func declareCoverVars(p *Package) map[string]*FileVar {

	return nil
}
