package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func screen(title string, objects ...fyne.CanvasObject) fyne.CanvasObject {
	content := []fyne.CanvasObject{widget.NewLabel(title), widget.NewSeparator()}
	content = append(content, objects...)
	return container.NewScroll(container.NewVBox(content...))
}

func messageScreen(message string) fyne.CanvasObject {
	return screen("ghgo", widget.NewLabel(message))
}

func prerequisiteScreen(title, message string) fyne.CanvasObject {
	return screen(title, widget.NewLabel(message))
}

func actionRow(objects ...fyne.CanvasObject) fyne.CanvasObject {
	return container.NewHBox(objects...)
}
