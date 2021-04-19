package config

type gocConfig struct {
	Debug            bool
	CurPkgDir        string
	CurModProjectDir string
	TmpModProjectDir string
	TmpPkgDir        string
	BinaryName       string
}

var GocConfig gocConfig

type goConfig struct {
	BuildA                 bool
	BuildBuildmode         string // -buildmode flag
	BuildMod               string // -mod flag
	BuildModReason         string // reason -mod flag is set, if set by default
	BuildI                 bool   // -i flag
	BuildLinkshared        bool   // -linkshared flag
	BuildMSan              bool   // -msan flag
	BuildN                 bool   // -n flag
	BuildO                 string // -o flag
	BuildP                 int    // -p flag
	BuildPkgdir            string // -pkgdir flag
	BuildRace              bool   // -race flag
	BuildToolexec          string // -toolexec flag
	BuildToolchainName     string
	BuildToolchainCompiler func() string
	BuildToolchainLinker   func() string
	BuildTrimpath          bool // -trimpath flag
	BuildV                 bool // -v flag
	BuildWork              bool // -work flag
	BuildX                 bool // -x flag
	// from buildcontext
	Installsuffix string // -installSuffix
	BuildTags     string // -tags
	// from load
	BuildAsmflags   string
	BuildCompiler   string
	BuildGcflags    string
	BuildGccgoflags string
	BuildLdflags    string

	// mod related
	ModCacheRW bool
	ModFile    string
}

var GoConfig goConfig
