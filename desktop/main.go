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
	var overlayMode bool
	var overlayPayload string
	flag.BoolVar(&daemonMode, "daemon", false, "run background reminder daemon")
	flag.BoolVar(&overlayMode, "overlay", false, "show a reminder overlay")
	flag.StringVar(&overlayPayload, "payload", "", "base64 JSON overlay payload")
	flag.Parse()

	if daemonMode {
		if err := runDaemon(); err != nil {
			println("Daemon error:", err.Error())
			os.Exit(1)
		}
		return
	}
	if overlayMode {
		if err := runOverlay(overlayPayload); err != nil {
			println("Overlay error:", err.Error())
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
