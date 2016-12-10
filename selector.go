package grill

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
)

// Selector represents a selector to the implementor
// Whoever is using the selector bit of functionality should
// ONLY have to interface with this.
//
// Each Selector has a list of *Option{} 's that are the things
// the user can pick from
type Selector struct {
				     // Internal
	i                  int       // The index the user is currently selecting
	buffer             string    // The next buffer to write as a string
	longestOptionLabel int       // The length of the longest option label at time of Render()

				     // Exported
	Options            []*Option // List of options
	Cursor             *Cursor   // The animated cursor to use
	StepMilli          int       // Time for each animation step in milliseconds
	Title              *Title    // The title of program, this is displayed once on ncurses initialization
}

type Title struct {
	Value            string // The raw string for the title of the selector
	TitleFgColor     int    // Enumerated color for the entire title foreground
	TitleBgColor     int    // Enumerated color for the entire title background
}

// Option is a single option in the selector. Each option can
// take an open ended Value interface{} that can be anything
// the implementation wants, the point is.. if the user selects
// this option, you will get that interface{} back!
//
// Selector{} 's are made up of options. These are what the user actually
// can pick from.
type Option struct {
	Selected      bool
	Label         string
	Value         interface{}
	OptionFgColor int // Enumerated color for the entire option label foreground
	OptionBgColor int // Enumerated color for the entire option label background
}

// A list of string values to use for the *Cursor{}
// Each of the Steps will be displayed on the screen for 1 step before changing
// Steps will repeat forever
type Cursor struct {
	i             int
	Steps         []*CursorStep
	CursorFgColor int
	CursorBgColor int
}

// A single step in the *Cursor{}, usually just a single character
type CursorStep struct {
	Value string
}

// NewSelector will return a new Selector{} for the implementation to interact with.
// All constructor logic should go here for Selector{} 's
func NewSelector() *Selector {
	s := &Selector{}
	s.StepMilli = 100 // Default to 100 milliseconds here, if anyone REALLY needs to override this they can
	return s
}

// Will create a new *Cursor{}
func NewCursor() *Cursor {
	return &Cursor{CursorFgColor: COLOR_BLUE, CursorBgColor: COLOR_BLACK}
}

// NewOption will return a new Option{} for the implementation to interact with.
// All constructor logic shold go here for Option{} 's
func NewOption(label string, value interface{}) *Option {
	o := &Option{Label: label, Value: value, OptionFgColor: COLOR_GREEN, OptionBgColor: COLOR_BLACK}
	return o
}

// Will create a new *CursorStep{} to be used in a *Cursor{}
func NewCursorStep(val string) *CursorStep {
	return &CursorStep{Value: val}
}

// Will create a new *Title{} to be used with a *Selector{}
func NewTitle(val string) *Title {
	if strings.Count(val, "\n") < 1 {
		val = val + "\n"
	}
	return &Title{Value: val, TitleFgColor: COLOR_GREEN, TitleBgColor: COLOR_BLACK}
}

// Basic default cursor, just implements a spinning line for the cursor
func NewDefaultCursor() *Cursor {
	c := &Cursor{}

	// Cute text cursor - thanks beastie ;)
	c.Steps = append(c.Steps, &CursorStep{Value: "<--[|]" })
	c.Steps = append(c.Steps, &CursorStep{Value: "<--[/]" })
	c.Steps = append(c.Steps, &CursorStep{Value: "<--[-]" })
	c.Steps = append(c.Steps, &CursorStep{Value: "<--[\\]" })
	return c
}

// Will add a *CursorStep{} to a *Cursor{}
func (c *Cursor) AddStep(step *CursorStep) {
	c.Steps = append(c.Steps, step)
}

// Will create a new *CursorStep{} from a value, and add it to the *Cursor{}
func (c *Cursor) NewAddStep(val string) {
	step := NewCursorStep(val)
	c.AddStep(step)
}

// Will add a *Cursor{} to a *Selector{}
func (s *Selector) AddCursor(c *Cursor) {
	s.Cursor = c
}

// Will add a *Title{} to a *Selector{}
func (s *Selector) AddTitle(t *Title) {
	s.Title = t
}

// AddOption will append() an *Option{} to the Selector{}
func (s *Selector) AddOption(o *Option) {
	s.Options = append(s.Options, o)
}

// NewAddOption is a hybrid function, that is used for simplicity. It will create a
// new *Option{} and call AddOption() with the newly created *Option{}
func (s *Selector) NewAddOption(label string, value interface{}) {
	o := NewOption(label, value)
	s.AddOption(o)
}

// Render is the function that will actually take over the user's TTY.
// This function will interact with the user's hardware bus in raw() mode, so even
// control characters like ctl^C and ctl^Z will need to be managed here! We also have
// a buffered and concurrent management of the user's Stdin and Stdout here.. so we
// also need to keep up with that.
//
// This is a very delicate function. If you are developing on this tool, and you break
// your TTY. It is recommended to use the *nix command `reset` to bring your terminal back
// to a working place again.
//
// This is an actual event loop from the hardware bus, that will parse the input
// based on it's decimal value in the ascii table.
//
// When Render() is called, if the user hasn't defined a *Cursor{}, the default *Cursor{} will be chosen
// When Render() is called, we can do some basic algebra to make the buffer calculation cleaner later, so
// we calculate things like longest option label for instance.
//
func (s *Selector) Render() error {
	// Check for default cursor
	if s.Cursor == nil {
		s.Cursor = NewDefaultCursor()
	}

	// Calculate longest option label
	lol := -1 // I couldn't help it..
	for _, opt := range s.Options {
		if len(opt.Label) > lol {
			lol = len(opt.Label)
		}
	}
	s.longestOptionLabel = lol

	// Init ncurses
	InitSelectorCurses(s.StepMilli)

	// Validate we have enough screen real estate
	maxCols := W.MaxX
	if (s.longestOptionLabel - 2) >= maxCols {
		End()
		return fmt.Errorf("Unable to render selector. Terminal not wide enough! %d %d", s.longestOptionLabel - 2, maxCols)
	}
	maxRows := W.MaxY
	if len(s.Options) > maxRows {
		End()
		return fmt.Errorf("Unable to render selector. Terminal not high enough! %d %d", len(s.Options), maxRows)
	}

	// Implement signal handler
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Init the buffer right away
	s.writeScreen()

	// Here we take our existing buffer and use that as our buffer foundation, we want this to appear seamless
	// TODO Kris here
	// This is where we need scr_dump or something..

	for {
		select {
		case <-ch:
			s.Exit()
			break
		default:
			s.refresh()
			if s.handleChar(GetCh()) {
				// If we return true, then we are done with the selector
				End()
				return nil
			}
			break
		}
	}
}

// This takes an ascii representation integer, and manages it for our functionality
// in the selector
func (s *Selector) handleChar(i int) (bool) {
	l := len(s.Options) //  N
	maxIndex := (l - 1) // (N - 1)

	switch i {
	case -1: //
		return false
	case 10 : //[enter]
		s.Options[s.i].Selected = true // Select this option
		return true
	case 66 : //[down]
		if s.i == maxIndex {
			// Do nothing
		} else {
			s.i = s.i + 1 // Reverse logic here
		}
		return false
	case 65 : //[up]
		if s.i == 0 {
			// Do nothing
		} else {
			s.i = s.i - 1 // Reverse logic here
		}
		return false
	default: // Any other char
		return false
	}
	// If we get here we have problems
	return false
}

// Run this every step.. will refresh our screen to make the screen appear alive, and reactive
func (s *Selector) refresh() {
	s.anistep()       // Increment our next animation step
	Clear()           // Erase the screen
	s.writeScreen()    // Calculate the buffer
}

// anistep (or animate step) is the logic that will be ran every step
// This should simply manage setting the index for our cursor step so that
// when the buffer is calculated, we should automatically have the correct
// cursor step.
func (s *Selector) anistep() {
	l := len(s.Cursor.Steps) //  N
	maxIndex := (l - 1)      // (N - 1)
	if s.Cursor.i == maxIndex {
		s.Cursor.i = 0
	} else {
		s.Cursor.i = s.Cursor.i + 1
	}
}

// writeScreen will flush a series of buffers out. Basically this function can be called
// on an empty screen, and we can trust that it will write the next screen needed for our
// animated screen.
func (s *Selector) writeScreen() {
	var buffer string
	if s.Title != nil {
		// We have a title, let's build the buffer and flush it
		// Title buffer - use 1024, as we should never have that many options
		// ------------------------------------------------------------------
		InitPair(1, s.Title.TitleFgColor, s.Title.TitleBgColor)
		ColorPairOn(1)
		buffer = buffer + s.Title.Value
		AddStr(buffer)
		ColorPairOff(1)
		buffer = ""
		// ------------------------------------------------------------------
	}

	for id, opt := range s.Options {
		lolDelta := (s.longestOptionLabel - len(opt.Label)) + 2 // Longest plus 2 because space is good
		var lolSpace string
		for i := 0; i <= lolDelta; i++ {
			lolSpace = lolSpace + " "
		}
		if s.i == id {
			// Line with the cursor - we have 2 buffers here
			// 1. The label
			// 2. The cursor
			// ------------------------------------------------------------------
			InitPair(2, opt.OptionFgColor, opt.OptionBgColor)
			buffer = buffer + opt.Label + lolSpace
			ColorPairOn(2)
			AddStr(buffer)
			ColorPairOff(2)
			buffer = ""
			// ------------------------------------------------------------------
			InitPair(3, s.Cursor.CursorFgColor, s.Cursor.CursorBgColor)
			buffer = s.Cursor.Steps[s.Cursor.i].Value // This is the cursor
			ColorPairOn(3)
			AddStr(buffer)
			ColorPairOff(3)
			buffer = ""
			// ------------------------------------------------------------------
		} else {
			// Just a plain old line
			// Write a single buffer for the line
			// ------------------------------------------------------------------
			InitPair(4, opt.OptionFgColor, opt.OptionBgColor)
			buffer = buffer + opt.Label
			ColorPairOn(4)
			AddStr(buffer)
			ColorPairOff(4)
			buffer = ""
			// ------------------------------------------------------------------
		}
		AddStr("\n")
	}
}

// Sig handler during the selector, most people will ctl^C when they freak out.
// So this at least gives them that escape plan
func (s *Selector) Exit() {
	End()
	os.Exit(1)
}

// GetSelectedOption will attempt to return the option that the user selected
// from the list of possible options. The only reason this should fail, is if
// something is horribly wrong with the code base
func (s *Selector) GetSelectedOption() (*Option, error) {
	var so *Option
	for _, o := range s.Options {
		if o.Selected {
			if so != nil {
				return nil, fmt.Errorf("Multiple selected values!")
			}
			so = o
		}
	}
	if so == nil {
		return nil, fmt.Errorf("Unable to find selected option!")
	}
	return so, nil
}

// Wrapper function to get the option label
func (o *Option) GetLabel() (string) {
	return o.Label
}

// Convenience function to attempt to get the value as an int
func (o *Option) GetValInt() (int, error) {
	switch t := o.Value.(type) {
	case int:
		return o.Value.(int), nil
	default:
		return 0, fmt.Errorf("Unable to return int for value %v, type %T", o.Value, t)
	}

}

// Convenience function to attempt to get the value as a string
func (o *Option) GetValString() (string, error) {
	switch t := o.Value.(type) {
	case string:
		return o.Value.(string), nil
	default:
		return "", fmt.Errorf("Unable to return string for value %v, type %T", o.Value, t)
	}
}

// Convenience function to get the value exactly as it was passed in
func (o *Option) GetValInterface() (interface{}) {
	return o.Value
}


