package main

import (
	"github.com/kris-nova/grill"
	"fmt"
)

type Choice struct {
	Value string
}

func main() {
	// Start a new selector
	selector := grill.NewSelector()

	// Add a title to the program
	titleStr := "Please make a selection from our wonderful options below.."
	title := grill.NewTitle(titleStr)
	title.TitleFgColor = grill.COLOR_WHITE
	title.TitleBgColor = grill.COLOR_BLACK
	selector.AddTitle(title)

	// Here we demonstrate that the implementation can use whatever they want for the value
	// So feel free to define types, and use them in the choices to make your life easier!
	// You can use *anything* for the second argument here, and get it back later
	selector.NewAddOption("First Choice", &Choice{Value: "The First choice wins"})
	selector.NewAddOption("2nd Choice", &Choice{Value: "The 2nd choice wins"})
	selector.NewAddOption("Third Choice", &Choice{Value: "The last choice wins"})

	// Demonstrate custom options with a default option
	defaultOption := grill.NewOption("I don't know", -1)
	defaultOption.OptionFgColor = grill.COLOR_RED
	selector.AddOption(defaultOption)



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

	// Configure the step speed. Every step, is when the buffer will re-write itself
	// so we can define any arbitrary number we want here. These are configured in
	// milliseconds, and the default value is 100 milliseconds. In this case, let's
	// use 200 to slow it down a bit.
	selector.StepMilli = 200

	// Render the menu on the user's screen
	err := selector.Render()
	if err != nil {
		panic(err)
	}

	// Get the option the user selected
	option, err := selector.GetSelectedOption()
	if err != nil {
		panic(err)
	}
	i := option.GetValInterface()
	fmt.Println(option.Label, i)
}
