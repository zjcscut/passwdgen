package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"passwdgen/theme"
	"passwdgen/ui"
)

const AppId = "passwdgen"

func main() {
	application := app.NewWithID(AppId)
	application.Preferences().SetBool("__MainWindowInit__", false)
	mainWindow := ui.InitMainWindow()
	mainWindow.SetIcon(theme.Logo)
	mainWindow.CenterOnScreen()
	mainWindow.SetMaster()
	mainWindow.Resize(fyne.NewSize(
		mainWindow.Canvas().Size().Width,
		mainWindow.Canvas().Size().Height,
	))
	application.Preferences().SetBool("__MainWindowInit__", true)
	mainWindow.ShowAndRun()
}
