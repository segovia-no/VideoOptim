package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "VideoOptim",
		Width:     760,
		Height:    540,
		MinWidth:  600,
		MinHeight: 400,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:         &options.RGBA{R: 246, G: 246, B: 246, A: 1},
		OnStartup:                app.startup,
		EnableDefaultContextMenu: false,
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
			},
			Appearance: mac.DefaultAppearance,
			About: &mac.AboutInfo{
				Title:   "VideoOptim",
				Message: "© 2026 VideoOptim\nVideo compression for macOS",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
