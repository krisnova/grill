package main

import (
	"github.com/kris-nova/grill"
	"fmt"
)

func main() {
	selector := grill.NewSelector()
	selector.NewAddOption("First Choice", 1)
	selector.NewAddOption("2nd Choice", 2)
	selector.NewAddOption("Third Choice", 3)
	selector.Render()
	option, err := selector.GetSelectedOption()
	if err != nil {
		panic(err)
	}
	i := option.GetValInterface()
	fmt.Println(option.Label, i)
}
