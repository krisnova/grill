package main

import (
	"github.com/kris-nova/grill"
	"fmt"
)

type Choice struct {
	Value string
}

func main() {
	selector := grill.NewSelector()

	// Here we demonstrate that the implementation can use whatever they want for the value
	// So feel free to define types, and use them in the choices to make your life easier!
	selector.NewAddOption("First Choice", &Choice{Value: "The First choice wins"})
	selector.NewAddOption("2nd Choice", &Choice{Value: "The 2nd choice wins"})
	selector.NewAddOption("Third Choice", &Choice{Value: "The last choice wins"})

	// Add a title to the program
	selector.Title = `
-----------------------------------
Please make a choice! If you dare!
-----------------------------------
`

	// Define a custom cursor here
	// You can have as many steps as you want
	cursor := &grill.Cursor{}
	cursor.Steps = append(cursor.Steps, &grill.CursorStep{Value: "<-- *" })
	cursor.Steps = append(cursor.Steps, &grill.CursorStep{Value: "<--  " })
	cursor.Steps = append(cursor.Steps, &grill.CursorStep{Value: "<-- ." })
	cursor.Steps = append(cursor.Steps, &grill.CursorStep{Value: "<--  " })

	// Drop in our cursor
	selector.Cursor = cursor

	// Turn the speed down, it defaults to 100
	selector.StepMilli = 200

	// Render the menu
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
