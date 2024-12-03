package main

import (
	"fyne.io/fyne/v2/app"
	"steganography-tool/internal/ui"
)

func main() {

	myApp := app.New()
	stegano := ui.NewSteganoUI(myApp)
	stegano.ShowAndRun()

}
