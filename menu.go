package main

import (
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func generateMenuItems(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	// ── VideoOptim ──────────────────────────────────────────────────────────
	vo := appMenu.AddSubmenu("VideoOptim")
	vo.AddText("About VideoOptim", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:about")
	})
	vo.AddSeparator()
	vo.AddText("Settings", keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:settings")
	})
	vo.AddSeparator()
	vo.AddText("Quit VideoOptim", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		wailsRuntime.Quit(app.ctx)
	})

	// ── File ─────────────────────────────────────────────────────────────────
	file := appMenu.AddSubmenu("File")
	file.AddText("Add Files…", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:open")
	})
	file.AddText("Add Folder…", keys.Combo("o", keys.ShiftKey, keys.CmdOrCtrlKey), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:open-folder")
	})
	file.AddSeparator()
	file.AddText("Clear List", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:clear")
	})

	return appMenu
}