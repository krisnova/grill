package grill

// #cgo LDFLAGS: -lncurses
// #include <stdlib.h>
// #include <ncurses.h>
//
// /* Define the curses API for the Go program */
//
// void c_timeout(int i) { timeout(i); }
// int c_scr_dump(const char *c) { return scr_dump(c); }
//
import "C"
import (
	"fmt"
	"unsafe"
)

type Window struct {
	CWindow *C.WINDOW
	MaxX    int
	MaxY    int
}

// Init this on package include
var W *Window = &Window{CWindow: C.stdscr}

// Build our ncurses implementation for the selector
//
// Kris TODO We need to figure out how to "screen dump" or capture the full content
// of the existing TTY before we bring ncurses to the table, ideally we append our selector
// onto an existing buffer or something to make ncurses appear seemles.. maybe some help from
// the open source community here?
func InitSelectorCurses(stepMilli int) {
	W.CWindow = C.initscr()
	//C.newterm(nil, C.stdin, C.stdout)

	C.raw()                       // Pass control chars up
	C.cbreak()                    // Let's read those control characters in like the others
	C.noecho()                    // Don't echo chars while reading
	C.c_timeout(C.int(stepMilli)) // Timeout our Getch() for every 100milli, so we can animate the screen
	C.start_color()               // Turn on colors - because fabulous

	// Build the terminal size
	W.MaxX = int(C.COLS)
	W.MaxY = int(C.LINES)
}

// Clear the screen
func Clear() int {
	return int(C.clear())
}

// Return native Getch() from the kernel
func GetCh() int {
	return int(C.getch())
}

// Please! For all that is sacred! Remember to set our TTY back
func End() {
	C.endwin()
}

func AddStr(str ...interface{}) {
	res := (*C.char)(C.CString(fmt.Sprint(str...)))
	defer C.free(unsafe.Pointer(res))
	C.addstr(res)
}

// Ncurses Color Codes
const (
	COLOR_BLACK = 0
	COLOR_RED = 1
	COLOR_GREEN = 2
	COLOR_YELLOW = 3
	COLOR_BLUE = 4
	COLOR_MAGENTA = 5
	COLOR_CYAN = 6
	COLOR_WHITE = 7
)

// attrn - The identifier of this attribute
// fg - the foreground color from the list above
// bg - the background color from the list above
func InitPair(attrn, fg, bg int) {
	C.init_pair(C.short(attrn), C.short(fg), C.short(bg))
}

// Will turn a color pair attribute on
func ColorPairOn(id int) {
	C.attron(C.COLOR_PAIR(C.int(id)))
}

// Will turn a color pair attribute off
func ColorPairOff(id int) {
	C.attroff(C.COLOR_PAIR(C.int(id)))
}

// Enable attribute
func Attron(attr int) {
	C.attron(C.int(attr))
}

// Disable attribute
func Attroff(attr int) {
	C.attroff(C.int(attr))
}

//func ScrDump() string {
//	var c *C.char
//	C.c_scr_dump(c)
//	fmt.Println(c)
//	return ""
//}

//func GetContents() {
//	var content string
//	C.c_winchnstr(W.CWindow, &content)
//	fmt.Println(content)
//}
