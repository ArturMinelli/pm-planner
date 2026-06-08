package main

import (
	"embed"
	"flag"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	var daemonMode bool
	flag.BoolVar(&daemonMode, "daemon", false, "run background reminder daemon")
	flag.Parse()

	if daemonMode {
		if err := runDaemon(); err != nil {
			println("Daemon error:", err.Error())
			os.Exit(1)
		}
		return
	}

	desktopApp := NewApp()

	err := wails.Run(&options.App{
		Title:     "PM — Planner",
		Width:     980,
		Height:    720,
		MinWidth:  720,
		MinHeight: 540,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: desktopApp.Startup,
		Bind: []interface{}{
			desktopApp,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
