package flag

import (
	"flag"

	"github.com/qiniu/goc/v2/pkg/config"
)

func addBuildFlags(cmdSet *flag.FlagSet) {
	cmdSet.BoolVar(&config.GoConfig.BuildA, "a", false, "")
	cmdSet.BoolVar(&config.GoConfig.BuildN, "n", false, "")
	cmdSet.IntVar(&config.GoConfig.BuildP, "p", 4, "")
	cmdSet.BoolVar(&config.GoConfig.BuildV, "v", false, "")
	cmdSet.BoolVar(&config.GoConfig.BuildX, "x", false, "")
	cmdSet.StringVar(&config.GoConfig.BuildBuildmode, "buildmode", "default", "")
	cmdSet.StringVar(&config.GoConfig.BuildMod, "mod", "", "")
	cmdSet.StringVar(&config.GoConfig.Installsuffix, "installsuffix", "", "")

	// 类型和 go 原生的不一样，这里纯粹是为了 parse 并传递给 go
	cmdSet.StringVar(&config.GoConfig.BuildAsmflags, "asmflags", "", "")
	cmdSet.StringVar(&config.GoConfig.BuildCompiler, "compiler", "", "")
	cmdSet.StringVar(&config.GoConfig.BuildGcflags, "gcflags", "", "")
	cmdSet.StringVar(&config.GoConfig.BuildGccgoflags, "gccgoflags", "", "")
	// mod related
	cmdSet.BoolVar(&config.GoConfig.ModCacheRW, "modcacherw", false, "")
	cmdSet.StringVar(&config.GoConfig.ModFile, "modfile", "", "")
	cmdSet.StringVar(&config.GoConfig.BuildLdflags, "ldflags", "", "")
	cmdSet.BoolVar(&config.GoConfig.BuildLinkshared, "linkshared", false, "")
	cmdSet.StringVar(&config.GoConfig.BuildPkgdir, "pkgdir", "", "")
	cmdSet.BoolVar(&config.GoConfig.BuildRace, "race", false, "")
	cmdSet.BoolVar(&config.GoConfig.BuildMSan, "msan", false, "")
	cmdSet.StringVar(&config.GoConfig.BuildTags, "tags", "", "")
	cmdSet.StringVar(&config.GoConfig.BuildToolexec, "toolexec", "", "")
	cmdSet.BoolVar(&config.GoConfig.BuildTrimpath, "trimpath", false, "")
	cmdSet.BoolVar(&config.GoConfig.BuildWork, "work", false, "")
}

func addOutputFlags(cmdSet *flag.FlagSet) {
	cmdSet.StringVar(&config.GoConfig.BuildO, "o", "", "")
}
