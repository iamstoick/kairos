// Author Name: Gerald Z. Villorente
// Author email: geraldvillorente@gmail.com
// @2025
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	runewidth "github.com/mattn/go-runewidth"
)

var (
	locations map[string]*time.Location
	digits    = map[rune][]string{
		'0': {"█████", "█   █", "█   █", "█   █", "█████"},
		'1': {"  █  ", " ██  ", "  █  ", "  █  ", "█████"},
		'2': {"█████", "    █", "█████", "█    ", "█████"},
		'3': {"█████", "    █", "█████", "    █", "█████"},
		'4': {"█   █", "█   █", "█████", "    █", "    █"},
		'5': {"█████", "█    ", "█████", "    █", "█████"},
		'6': {"█████", "█    ", "█████", "█   █", "█████"},
		'7': {"█████", "    █", "    █", "    █", "    █"},
		'8': {"█████", "█   █", "█████", "█   █", "█████"},
		'9': {"█████", "█   █", "█████", "    █", "█████"},
		':': {"     ", "  █  ", "     ", "  █  ", "     "},
		// AM and PM digit mappings...
		'A': {"     ", " ██  ", "█  █ ", "████ ", "█  █ "},
		'M': {"     ", "█ █ █", "█████", "█ █ █", "█   █"},
		'P': {"     ", "████ ", "█  █ ", "████ ", "█    "},
		// Ensure ' ' (space) is also defined if used
		' ': {"     ", "     ", "     ", "     ", "     "},
	}
	timezones = []struct {
		name     string
		location string
	}{
		{"UTC", "UTC"},                     // Index 0 (Top)
		{"PST/DST", "America/Los_Angeles"}, // Index 1
		{"GMT", "Etc/GMT"},                 // Index 2
		{"Philippine Time", "Asia/Manila"}, // Index 3
		{"CST", "America/Chicago"},         // Index 4
		{"MST", "America/Denver"},          // Index 5
		{"EST", "America/New_York"},        // Index 6
	}
)

// main initializes the application's GUI and its components, sets up timezone locations,
// and runs the main event loop for handling user input and screen updates.
func main() {
	// Declare a variable to capture errors.
	var err error
	// Create a new GUI with normal output mode.
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err) // Log a panic error if the GUI cannot be created.
	}
	defer g.Close() // Ensure that g.Close() is called when the function exits to clean up resources.

	// Initialize location objects for storing timezone information.
	locations = make(map[string]*time.Location) // Create a map to hold location data.
	for _, tz := range timezones {              // Loop through predefined timezones.
		// Load each location by timezone identifier.
		locations[tz.name], err = time.LoadLocation(tz.location)
		if err != nil {
			// Log a fatal error and exit if a location cannot be loaded.
			log.Fatalf("Failed to load location for %s: %v", tz.name, err)
		}
	}
	// Set the function that defines the layout of the GUI.
	g.SetManagerFunc(layout)
	// Set keybindings for the GUI; this configures how user inputs are handled.
	if err := KeyBindings(g); err != nil {
		log.Panicln("Failed to create keybindings: ", err)
	}

	// Start a goroutine to update the time every second.
	go UpdateTimeEverySecond(g)

	// Start the main GUI loop which handles drawing and events.
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

// UpdateTimeEveryMinute is responsible for updating the time every minute.
func UpdateTimeEverySecond(g *gocui.Gui) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		g.Update(func(*gocui.Gui) error {
			if err := layout(g); err != nil {
				log.Println("Failed to layout:", err)
			}
			return nil
		})
	}
}

// layout configures the layout of the GUI views within the application window.
// It sets up a top view for displaying time in UTC and three bottom views for
// different time zones using ASCII art for time display.
/*func layout(g *gocui.Gui) error {
	// Obtain the maximum horizontal dimension of the GUI.
	maxX, _ := g.Size()

	// Determine appropriate heights based on the expected size of ASCII art.
	topHeight := 10 // Height sufficient to display the ASCII art time.
	// Set the height for the bottom views, typically twice the height of the top view.
	bottomHeight := topHeight * 2

	// Setting up or updating the top view for UTC.
	v, err := g.SetView("top", 0, 0, maxX-1, topHeight)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	v.Title = " UTC: Coordinated Universal Time "
	UpdateViewTime(v, locations["UTC"]) // Ensure this is called regardless of the error type.

	// Set up the bottom left view to display Pacific Standard Time/Daylight Saving Time.
	SetupBottomView(g, "bottom_left", 0, topHeight+1, maxX/3-1, bottomHeight, " PST/DST: America/Los_Angeles ", locations["PST/DST"])
	// Set up the bottom middle view to display Eastern Standard Time.
	SetupBottomView(g, "bottom_middle", maxX/3, topHeight+1, 2*maxX/3-1, bottomHeight, " EST: America/New_York ", locations["EST"])
	// Set up the bottom right view to display Philippine Time.
	SetupBottomView(g, "bottom_right", 2*maxX/3, topHeight+1, maxX-1, bottomHeight, " Asia/Manila ", locations["Philippine Time"])

	// Return nil to indicate successful layout setup.
	return nil
}*/

/*
Layout until February 2026

	func layout(g *gocui.Gui) error {
		maxX, _ := g.Size()
		topHeight := 10
		bottomHeight := topHeight * 2

		// Set the top view
		topView, err := g.SetView("top", 0, 0, maxX-1, topHeight-1)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		topView.Title = " " + timezones[0].name + " "
		UpdateViewTime(topView, locations[timezones[0].name])

		// Calculate the width of each bottom view based on the number of timezones minus the top view
		bottomViewWidth := maxX / (len(timezones) - 1)

		// Set bottom views.
		for i := 1; i < len(timezones); i++ {
			x0 := (i - 1) * bottomViewWidth
			x1 := x0 + bottomViewWidth - 1
			viewName := fmt.Sprintf("bottom%d", i)
			bottomView, err := g.SetView(viewName, x0, topHeight, x1, bottomHeight)
			if err != nil && err != gocui.ErrUnknownView {
				return err
			}
			bottomView.Title = " " + timezones[i].name + " "
			UpdateViewTime(bottomView, locations[timezones[i].name])
		}

		return nil
	}
*/
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Reserve the last line for the help footer
	gridMaxY := maxY - 2
	rowHeight := gridMaxY / 3

	// Top View (Index 0)
	if v, err := g.SetView("top", 0, 0, maxX-1, rowHeight-1); err != nil && err != gocui.ErrUnknownView {
		return err
	} else {
		v.Title = " " + timezones[0].name + " "
		UpdateViewTime(v, locations[timezones[0].name])
	}

	// Bottom Grid (Indices 1-6)
	itemsPerRow := 3
	columnWidth := maxX / itemsPerRow
	for i := 1; i < len(timezones); i++ {
		rowNum := (i - 1) / itemsPerRow
		colNum := (i - 1) % itemsPerRow

		x0 := colNum * columnWidth
		x1 := x0 + columnWidth - 1
		if colNum == itemsPerRow-1 {
			x1 = maxX - 1
		}

		y0 := (rowNum + 1) * rowHeight
		y1 := y0 + rowHeight - 1

		viewName := fmt.Sprintf("bottom%d", i)
		if v, err := g.SetView(viewName, x0, y0, x1, y1); err != nil && err != gocui.ErrUnknownView {
			return err
		} else {
			v.Title = fmt.Sprintf(" [%d] %s ", i, timezones[i].name) // Show key in title
			UpdateViewTime(v, locations[timezones[i].name])
		}
	}

	// 3. Help Footer View
	if v, err := g.SetView("help", 0, maxY-2, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorCyan
		fmt.Fprintln(v, CenterDate("Keys [1-6]: Swap with Top | Ctrl+C: Quit", maxX))
	}

	return nil
}

// setupBottomView configures a bottom view in the GUI for displaying time zone specific time.
// It sets view dimensions, title, and initializes the view with current time for the specified location.
//
// Parameters:
//
//	g    - The GUI object which manages all views.
//	name - The name of the view, which acts as a unique identifier.
//	x0, y0 - The starting coordinates (top-left corner) of the view.
//	x1, y1 - The ending coordinates (bottom-right corner) of the view.
//	title - The title of the view, displayed at the top of the view box.
//	loc   - The time location (timezone) that this view will display.
func SetupBottomView(g *gocui.Gui, name string, x0, y0, x1, y1 int, title string, loc *time.Location) {
	v, err := g.SetView(name, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return
	}
	v.Title = title
	v.Highlight = true
	UpdateViewTime(v, loc) // Ensure this is updating the time correctly
}

// updateViewTime updates the specified view with the current time and date in the given location.
// It formats the time in a 12-hour format with AM/PM and displays the date in bold.
//
// Parameters:
//
//	v   - the gocui view where the time and date will be displayed.
//	loc - the time location (timezone) to display the time for.
func UpdateViewTime(v *gocui.View, loc *time.Location) {
	// Fetch the current time in the specified location.
	now := time.Now().In(loc)
	// Clear the view's content to prepare for new content.
	v.Clear()

	// Setup padding where text should start to provide visual margins within the view.
	topPadding, leftPadding := 1, 2 // Add vertical and horizontal padding.
	// Move the cursor to the start position considering the padding.
	v.SetCursor(leftPadding, topPadding)
	// Fetch the width of the view for alignment purposes, the returned height is ignored.
	width, height := v.Size()

	// If the view is too small for ASCII, show plain text instead
	if height < 7 {
		fmt.Fprintf(v, "\n\n%s", CenterDate(now.Format("03:04:05 PM"), width))
		fmt.Fprintf(v, "\n%s", CenterDate(now.Format("Mon, Jan 2"), width))
		return
	}

	// Format the current time as a string in the 12-hour format including AM/PM.
	timeStr := now.Format("03:04 PM") // This will output, e.g., "09:55 PM"
	// Convert the time string to ASCII art representation for visual enhancement.
	asciiArt := PrintTimeASCII(timeStr)
	// Fetch the maximum width of the view again to ensure accurate width during alignment.
	maxWidth, _ := v.Size() // Get the maximum width of the view.
	fmt.Fprint(v, "\n\n")
	// Print each line of the ASCII art centered within the view.
	for _, line := range asciiArt {
		fmt.Fprintln(v, CenterTime(line, maxWidth))
	}
	// Format the current date in bold for emphasis and clarity.
	date := fmt.Sprintf("\x1b[1m%s\x1b[0m", now.Format("Monday, January 2, 2006")) // Bold the date.
	// Print the centered bold date at the bottom of the time display.
	fmt.Fprint(v, "\n")
	fmt.Fprintln(v, CenterDate(date, width))
}

// centerTime centers a given string within a specified width using spaces for padding.
// Parameters:
//
//	s - the string to be centered.
//	width - the total width within which the string should be centered.
func CenterTime(s string, width int) string {
	// Calculate the number of spaces needed on each side of the string to center it.
	// runewidth.StringWidth(s) computes the visual width of the string considering wide characters.
	padSize := (width - runewidth.StringWidth(s)) / 2
	// If the calculated padding size is greater than zero, pad the string with spaces.
	if padSize > 0 {
		// strings.Repeat(" ", padSize) creates a string consisting of 'padSize' spaces.
		// The string 's' is sandwiched between two such strings to center it within the 'width'.
		return strings.Repeat(" ", padSize) + s + strings.Repeat(" ", padSize)
	}
	// If no padding is needed, or if 'padSize' is zero or negative, return the original string.
	return s
}

// centerDate centers a string within a given width, ensuring that ANSI escape codes do not affect the alignment.
// Parameters:
//
//	s - the string to be centered, potentially containing ANSI escape codes.
//	width - the total width within which the string should be centered.
func CenterDate(s string, width int) string {
	// Remove the ANSI escape codes for bold from the string to calculate the visual width correctly.
	cleanString := strings.Replace(s, "\x1b[1m", "", -1)          // Remove the ANSI start bold.
	cleanString = strings.Replace(cleanString, "\x1b[0m", "", -1) // Remove the ANSI end bold.
	// Calculate the actual visual width of the string using runewidth.StringWidth,
	// which accounts for wide characters and correctly computes length ignoring ANSI codes.
	lineWidth := runewidth.StringWidth(cleanString) // Calculate the visual width of the string.
	// Calculate how many spaces are needed on each side to center the text within the specified width.
	padSize := (width - lineWidth) / 2 // Determine the number of spaces needed to pad the string on both sides.

	if padSize > 0 {
		padding := strings.Repeat(" ", padSize)
		return padding + s + padding // Pad the original string to maintain ANSI codes.
	}
	return s // Return as is if no padding is needed.
}

/*
Until Feb 2026

	func KeyBindings(g *gocui.Gui) error {
		if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			return gocui.ErrQuit
		}); err != nil {
			return err
		}

		// Bind Ctrl+A to swapTimezones function.
		errSwapToptoBottomLeft := g.SetKeybinding("", gocui.KeyCtrlA, gocui.ModNone, SwapToptoBottomLeft)
		if errSwapToptoBottomLeft != nil {
			return errSwapToptoBottomLeft
		}

		// Bind Ctrl+S to swapTimezones function.
		errSwapToptoBottomMiddle := g.SetKeybinding("", gocui.KeyCtrlS, gocui.ModNone, SwapToptoBottomMiddle)
		if errSwapToptoBottomMiddle != nil {
			return errSwapToptoBottomMiddle
		}

		// Bind Ctrl+D to swapTimezones function.
		errSwapToptoBottomRight := g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, SwapToptoBottomRight)
		if errSwapToptoBottomRight != nil {
			return errSwapToptoBottomRight
		}

		return nil
	}
*/
func KeyBindings(g *gocui.Gui) error {
	// Standard Quit
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	})

	// Map keys '1' through '6' to swap timezones[0] with timezones[1..6]
	keys := []rune{'1', '2', '3', '4', '5', '6'}
	for i, key := range keys {
		targetIdx := i + 1
		g.SetKeybinding("", key, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			return SwapTimezones(g, 0, targetIdx)
		})
	}
	return nil
}

func SwapTimezones(g *gocui.Gui, idxA, idxB int) error {
	if idxB < len(timezones) {
		timezones[idxA], timezones[idxB] = timezones[idxB], timezones[idxA]
	}
	return nil // The 1-second ticker will refresh the display automatically
}

func SwapToptoBottomRight(g *gocui.Gui, v *gocui.View) error {
	// Swap UTC and Philippine Time in the `timezones` slice
	// Assuming they are at specific indices, you might need to adjust these based on your slice setup
	// Assuming UTC is index 0 and Philippine Time is index 3
	// Swap the timezones. This is a more generic approach that works for any slice length.
	timezones[0], timezones[3] = timezones[3], timezones[0]

	g.Update(func(gui *gocui.Gui) error {
		// Directly handle the error inside this function.
		err := layout(gui)
		if err != nil {
			log.Println("Error updating GUI:", err)
		}
		return nil
	})

	return nil
}

func SwapToptoBottomLeft(g *gocui.Gui, v *gocui.View) error {
	// Swap UTC and PST in the `timezones` slice
	// Assuming they are at specific indices, you might need to adjust these based on your slice setup
	// Assuming UTC is index 0 and PST is index 1
	// Swap the timezones. This is a more generic approach that works for any slice length.
	timezones[0], timezones[1] = timezones[1], timezones[0]

	g.Update(func(gui *gocui.Gui) error {
		// Directly handle the error inside this function.
		err := layout(gui)
		if err != nil {
			log.Println("Error updating GUI:", err)
		}
		return nil
	})

	return nil
}

func SwapToptoBottomMiddle(g *gocui.Gui, v *gocui.View) error {
	// Swap UTC and PST in the `timezones` slice
	// Assuming they are at specific indices, you might need to adjust these based on your slice setup
	// Assuming UTC is index 0 and PST is index 1
	// Swap the timezones. This is a more generic approach that works for any slice length.
	timezones[0], timezones[2] = timezones[2], timezones[0]

	g.Update(func(gui *gocui.Gui) error {
		// Directly handle the error inside this function.
		err := layout(gui)
		if err != nil {
			log.Println("Error updating GUI:", err)
		}
		return nil
	})

	return nil
}

func SwapUTCtoPST(g *gocui.Gui, v *gocui.View) error {
	// Swap UTC and PST in the `timezones` slice
	// Assuming they are at specific indices, you might need to adjust these based on your slice setup
	// Assuming UTC is index 0 and PST is index 1
	// Swap the timezones. This is a more generic approach that works for any slice length.
	timezones[0], timezones[1] = timezones[1], timezones[0]

	g.Update(func(gui *gocui.Gui) error {
		// Directly handle the error inside this function.
		err := layout(gui)
		if err != nil {
			log.Println("Error updating GUI:", err)
		}
		return nil
	})

	return nil
}

// nextView cycles to the next view in the GUI.
func NextView(g *gocui.Gui, v *gocui.View) error {
	return SwitchView(g, true)
}

// prevView cycles to the previous view in the GUI.
func PrevView(g *gocui.Gui, v *gocui.View) error {
	return SwitchView(g, false)
}

// switchView changes the currently active view in the gocui GUI.
// Parameters:
//
//	g - the GUI object containing all views.
//	next - a boolean that determines the direction of the switch; true for next, false for previous.
func SwitchView(g *gocui.Gui, next bool) error {
	// Retrieve all views managed by the GUI.
	views := g.Views()
	// If there are less than two views, there's no need to switch, return immediately.
	if len(views) < 2 {
		return nil // Not enough views to switch between.
	}
	// Variable to store the index of the current active view.
	var currentIdx int
	// Iterate over all views to find the index of the current view.
	for i, view := range views {
		if g.CurrentView() == view {
			currentIdx = i // Store the index of the current view.
			break          // Stop the loop once the current view is found.
		}
	}
	// Calculate the index of the next or previous view to switch to based on the 'next' parameter.
	if next {
		// Increment the index and wrap around using modulo to cycle through views circularly.
		currentIdx = (currentIdx + 1) % len(views)
	} else {
		// Decrement the index and wrap around using modulo to cycle through views circularly.
		// This ensures that navigating backwards from the first view brings you to the last view.
		currentIdx = (currentIdx - 1 + len(views)) % len(views)
	}
	// Set the current view to the view at the calculated index.
	g.SetCurrentView(views[currentIdx].Name())
	// Return no error upon successful completion of the view switch.
	return nil
}

// printTimeASCII converts a time string into a series of lines that form an ASCII art representation.
func PrintTimeASCII(t string) []string {
	// Create an array to hold 5 strings, as each digit or character in the ASCII art representation
	// is assumed to span 5 lines vertically.
	// Assuming each part of the digits and letters has 5 lines
	lines := make([]string, 5)
	// Iterate over each character in the input string `t`.
	for _, digit := range t {
		// Retrieve the ASCII art corresponding to the current character.
		art, ok := digits[digit]
		// If the ASCII art for the character is not found, output a debug message and skip this character.
		if !ok {
			// Debug: check for missing characters in the ASCII art map.
			fmt.Println("Missing art for:", string(digit))
			// Skip missing characters to prevent panic.
			continue
		}
		// For each line of the ASCII art (assumed to be 5 lines), append it to the corresponding line in the `lines` array.
		// Loop through each line of the ASCII art.
		for i := 0; i < 5; i++ {
			// Add a space after each character for visual separation.
			lines[i] += art[i] + " "
		}
	}
	// Return the completed lines of ASCII art.
	return lines
}
