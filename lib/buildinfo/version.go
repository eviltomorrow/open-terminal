package buildinfo

import (
	"runtime"

	"github.com/fatih/color"
)

// -X with build
var (
	MainVersion string
	GoVersion   = runtime.Version()
	GoOSArch    = runtime.GOOS + "/" + runtime.GOARCH
	GitSha      string
	BuildTime   string
)

var (
	bluebold  = color.New(color.FgBlue, color.Bold)
	whitebold = color.New(color.FgWhite, color.Bold)
)

func Version() string {
	var s1 = bluebold.Sprintf("Version: ")
	var s2 = whitebold.Sprintf("%s %s (commit-id=%s)", AppName, MainVersion, GitSha)
	var s3 = bluebold.Sprintf("Runtime: ")
	var s4 = whitebold.Sprintf("%s %s RELEASE.%s", GoVersion, GoOSArch, BuildTime)
	return s1 + s2 + "\r\n" + s3 + s4
}
