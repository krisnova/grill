package grill

import (
	"fmt"
	"os"
	"os/signal"
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
	Title              string    // The title of program, this is displayed once on ncurses initialization
}

// Option is a single option in the selector. Each option can
// take an open ended Value interface{} that can be anything
// the implementation wants, the point is.. if the user selects
// this option, you will get that interface{} back!
//
// Selector{} 's are made up of options. These are what the user actually
// can pick from.
type Option struct {
	Selected bool
	Label    string
	Value    interface{}
}

// A list of string values to use for the *Cursor{}
// Each of the Steps will be displayed on the screen for 1 step before changing
// Steps will repeat forever
type Cursor struct {
	i     int
	Steps []*CursorStep
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

// NewOption will return a new Option{} for the implementation to interact with.
// All constructor logic shold go here for Option{} 's
func NewOption(label string, value interface{}) *Option {
	o := &Option{Label: label, Value: value}
	return o
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
func (s *Selector) Render() error {
	if s.Cursor == nil {
		s.Cursor = NewDefaultCursor()
	}
	s.buffer = s.Title

	// Calculate longest option label
	lol := -1 // I couldn't help it..
	for _, opt := range s.Options {
		if len(opt.Label) > lol {
			lol = len(opt.Label)
		}
	}
	s.longestOptionLabel = lol

	InitSelectorCurses(s.StepMilli)
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

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Here we take our existing buffer and use that as our buffer foundation, we want this to appear seamless


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
	s.anistep()
	Clear()           // Erase the screen
	s.calcBuffer()        // Calculate the buffer
	AddStr(s.buffer)  // Write the new buffer out
}

// anistep (or animate step) is the logic that will be ran every step
// This should simply manage setting the index for our cursor step so that
// when the buffer is calculated, we should automatically have the correct
// cursor step.
func (s *Selector) anistep() {
	l := len(s.Cursor.Steps) // Default 4
	maxIndex := (l - 1)      // Default 3
	if s.Cursor.i == maxIndex {
		s.Cursor.i = 0
	} else {
		s.Cursor.i = s.Cursor.i + 1
	}
}

// calcBuffer will calculate the new buffer for the screen to write
// The data calculated here will be stored in the selector, to be written out
// to ncurses later.
//
// This is where all the math-y things happen for the package. Basically we
// have to programmatically process the entire screen for every animated step
// we bring to life.
func (s *Selector) calcBuffer() {
	var buffer string
	buffer = buffer + s.Title // Always add our title
	for id, opt := range s.Options {
		lolDelta := (s.longestOptionLabel - len(opt.Label)) + 2 // Longest plus 2 because space is good
		var lolSpace string
		for i := 0; i <= lolDelta; i++ {
			lolSpace = lolSpace + " "
		}
		if s.i == id {
			// Line with the cursor
			buffer = buffer + opt.Label + lolSpace + s.Cursor.Steps[s.Cursor.i].Value + "\n"
		} else {
			// Just a plain old line
			buffer = buffer + opt.Label + "\n"
		}
	}

	// Clobber the buffer every time
	s.buffer = buffer
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


