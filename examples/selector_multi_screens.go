package main

import (
	"fmt"
	"github.com/kris-nova/grill"
)

func main() {

	mainScreen := mainScreen()

	// Get the option the user selected
	option, err := renderScreen(mainScreen)
	if err != nil {
		panic(err)
	}
	i := option.GetValInterface()
	fmt.Println(option.Label, i)
}

func renderScreen(screen *grill.Selector) (*grill.Option, error) {

	// clear selection from selector/screen
	// TODO: Maybe this should be method in grill.Selector
	// like ClearSelection or something
	for _, option := range screen.Options {
		option.Selected = false
	}

	// Render the menu
	err := screen.Render()
	if err != nil {
		panic(err)
	}

	// Get the option the user selected
	option, err := screen.GetSelectedOption()
	if err != nil {
		panic(err)
	}
	nextScreen, ok := option.GetValInterface().(*grill.Selector)
	if ok {
		option, err = renderScreen(nextScreen)
	}

	return option, err

}

func mainScreen() *grill.Selector {

	selector := grill.NewSelector()

	title := grill.NewTitle(`
-----------------------------------
             OS Selector
-----------------------------------
`)
	title.TitleFgColor = grill.COLOR_WHITE
	title.TitleBgColor = grill.COLOR_BLACK
	selector.AddTitle(title)

	selector.NewAddOption("Linux", linuxScreen(selector))
	selector.NewAddOption("BSD", bsdScreen(selector))
	selector.NewAddOption("Window$", "Blue Screen of Death!")

	// Define a custom cursor here
	// You can have as many steps as you want
	cursor := grill.NewCursor()
	cursor.NewAddStep("<[|]")
	cursor.NewAddStep("<[/]")
	cursor.NewAddStep("<[-]")
	cursor.NewAddStep("<[\\]")
	cursor.CursorFgColor = grill.COLOR_WHITE

	// Drop in our cursor
	selector.AddCursor(cursor)

	// Turn the speed down, it defaults to 100
	selector.StepMilli = 200

	return selector
}

func linuxScreen(mainScreen *grill.Selector) *grill.Selector {

	selector := grill.NewSelector()

	title := grill.NewTitle(`
-----------------------------------
             Linux
-----------------------------------
`)
	title.TitleFgColor = grill.COLOR_WHITE
	title.TitleBgColor = grill.COLOR_BLACK
	selector.AddTitle(title)

	selector.NewAddOption("<-OS Selector", mainScreen)
	selector.NewAddOption("Mascot", "Tux")

	// Define a custom cursor here
	// You can have as many steps as you want
	cursor := grill.NewCursor()
	cursor.NewAddStep("<[|]")
	cursor.NewAddStep("<[/]")
	cursor.NewAddStep("<[-]")
	cursor.NewAddStep("<[\\]")
	cursor.CursorFgColor = grill.COLOR_MAGENTA

	// Drop in our cursor
	selector.AddCursor(cursor)

	// Turn the speed down, it defaults to 100
	selector.StepMilli = 200

	return selector
}

func bsdScreen(mainScreen *grill.Selector) *grill.Selector {

	selector := grill.NewSelector()

	title := grill.NewTitle(`
-----------------------------------
                BSD
-----------------------------------
`)
	title.TitleFgColor = grill.COLOR_WHITE
	title.TitleBgColor = grill.COLOR_BLACK
	selector.AddTitle(title)

	selector.NewAddOption("<-OS Selector", mainScreen)
	selector.NewAddOption("Mascot", "Beastie")

	// Define a custom cursor here
	// You can have as many steps as you want
	cursor := grill.NewCursor()
	cursor.NewAddStep("<[|]")
	cursor.NewAddStep("<[/]")
	cursor.NewAddStep("<[-]")
	cursor.NewAddStep("<[\\]")
	cursor.CursorFgColor = grill.COLOR_RED

	// Drop in our cursor
	selector.AddCursor(cursor)

	// Turn the speed down, it defaults to 100
	selector.StepMilli = 200

	return selector
}
