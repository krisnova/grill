package main

import (
	"fmt"
	"github.com/kris-nova/grill"
)

func main() {

	mainScreen := mainScreen()

	// Get the option the user selected
	option, err := processScreen(mainScreen)
	if err != nil {
		panic(err)
	}
	i := option.GetValInterface()
	fmt.Println(option.Label, i)
}

func processScreen(screen *grill.Selector) (*grill.Option, error) {

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
		option, err = processScreen(nextScreen)
	}

	return option, err

}

func mainScreen() *grill.Selector {

	selector := grill.NewSelector()

	selector.NewAddOption("Linux", linuxScreen(selector))
	selector.NewAddOption("BSD", bsdScreen(selector))
	selector.NewAddOption("Window$", "Blue Screen of Death!")

	// Add a title to the program
	selector.Title = `
-----------------------------------
             OS Selector
-----------------------------------
`

	// Turn the speed down, it defaults to 100
	selector.StepMilli = 200

	return selector
}

func linuxScreen(mainScreen *grill.Selector) *grill.Selector {

	selector := grill.NewSelector()

	selector.NewAddOption("OS Selector", mainScreen)
	selector.NewAddOption("Mascot", "Tux")

	// Add a title to the program
	selector.Title = `
-----------------------------------
             Linux
-----------------------------------
`

	// Turn the speed down, it defaults to 100
	selector.StepMilli = 200

	return selector
}

func bsdScreen(mainScreen *grill.Selector) *grill.Selector {

	selector := grill.NewSelector()

	selector.NewAddOption("OS Selector", mainScreen)
	selector.NewAddOption("Mascot", "Beastie")

	// Add a title to the program
	selector.Title = `
-----------------------------------
                BSD
-----------------------------------
`

	// Turn the speed down, it defaults to 100
	selector.StepMilli = 200

	return selector
}
