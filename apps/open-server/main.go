package main

import (
	"log"

	"github.com/eviltomorrow/open-terminal/apps/open-server/cmd"
	"github.com/eviltomorrow/open-terminal/lib/buildinfo"
	"github.com/eviltomorrow/open-terminal/lib/system"
)

var (
	AppName     = "king-storage"
	MainVersion = "unknown"
	GitSha      = "unknown"
	BuildTime   = "unknown"
)

func init() {
	buildinfo.AppName = AppName
	buildinfo.MainVersion = MainVersion
	buildinfo.GitSha = GitSha
	buildinfo.BuildTime = BuildTime
}

func main() {
	if err := system.LoadRuntime(); err != nil {
		log.Fatalf("[F] App: load system runtime failure, nest error: %v", err)
	}

	if err := cmd.RunApp(); err != nil {
		log.Fatalf("[F] App: run app failure, nest error: %v", err)
	}
}
