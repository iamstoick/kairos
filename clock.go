// Author Name: Gerald Z. Villorente
// Author email: geraldvillorente@gmail.com
// @2025-2026
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
		'A': {"     ", " ██  ", "█  █ ", "████ ", "█  █ "},
		'M': {"     ", "█ █ █", "█████", "█ █ █", "█   █"},
		'P': {"     ", "████ ", "█  █ ", "████ ", "█    "},
		' ': {"     ", "     ", "     ", "     ", "     "},
	}
	timezones = []struct {
		name     string
		location string
	}{
		{"UTC", "UTC"},
		{"PST/DST", "America/Los_Angeles"},
		{"GMT", "Etc/GMT"},
		{"PHL Time", "Asia/Manila"},
		{"CST", "America/Chicago"},
		{"MST", "America/Denver"},
		{"EST", "America/New_York"},
	}
)

func main() {
	// Initialize the GUI
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	// Ensures that the GUI resources are properly released when the program exits.
	defer g.Close()

	// Load timezones into memory for quick access during updates.
	locations = make(map[string]*time.Location)
	for _, tz := range timezones {
		// Loads the timezone location from the IANA Time Zone database.
		loc, err := time.LoadLocation(tz.location)
		if err != nil {
			log.Fatalf("Failed to load location for %s: %v", tz.name, err)
		}
		// Stores the loaded location in the locations map with the timezone name as the key.
		locations[tz.name] = loc
	}

	// Set the layout function that will be called to draw the UI.
	g.SetManagerFunc(layout)
	// Set up keybindings for user interactions (swapping timezones and quitting the application).
	if err := KeyBindings(g); err != nil {
		log.Panicln("Failed to create keybindings: ", err)
	}

	// Update the UI every second to reflect the current time.
	go func() {
		// Creates a ticker that sends a value on a channel every second.
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			// Calls the Update method of the GUI to trigger a redraw of the UI.
			g.Update(func(g *gocui.Gui) error { return nil })
		}
	}()

	// Start the main event loop for the GUI. This loop listens for user
	// input and updates the UI accordingly.
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

/**
 * This function is responsible for setting up the layout of the terminal UI using the gocui library.
 * It divides the screen into a top section for the primary timezone and a grid of smaller sections for additional timezones.
 * Each section displays the current time, date, and business hours status for its respective timezone.
 *
 * The function also includes a help footer at the bottom of the screen that provides instructions for user interactions.
 *
 * @param g - The gocui.Gui object representing the terminal UI.
 * @returns An error if any issues occur during view creation or layout setup.
 */
func layout(g *gocui.Gui) error {
	// Retrieves the current width (maxX) and height (maxY) of your terminal window.
	maxX, maxY := g.Size()
	// Reserves the bottom two lines of the terminal so the "Help Footer" doesn't overlap with the clocks.
	gridMaxY := maxY - 2
	// Divides the available height into three equal horizontal sections: one for the top clock and two for the bottom grid rows.
	rowHeight := gridMaxY / 3

	// Top View (Index 0)
	// The top section is reserved for the primary timezone (UTC). It spans the entire width of the terminal and occupies the first third of the height.
	if v, err := g.SetView("top", 0, 0, maxX-1, rowHeight-1); err != nil && err != gocui.ErrUnknownView {
		return err
	} else {
		// Gets the current time for the primary timezone (UTC) and sets the title of the top view
		// to include the timezone name, a day/night icon, and the business hours indicator.
		now := time.Now().In(locations[timezones[0].name])
		// The title format is: " UTC 🌞 🟢" (for example), where the icon and business hours indicator change based on the current time.
		icon := getDayNightIcon(now)
		// The business hours indicator is determined by the getBusinessHoursIndicator function,
		// which checks if the current time falls within standard working hours.
		biz := getBusinessHoursIndicator(now)
		// Sets the title of the top view to display the timezone name, day/night icon, and business hours indicator.
		v.Title = fmt.Sprintf(" %s %s %s", timezones[0].name, icon, biz)
		// Updates the content of the top view to display the current time and date in the primary timezone.
		UpdateViewTime(v, locations[timezones[0].name])
	}

	// Bottom Grid (Indices 1-6)
	// The bottom section is divided into a grid of smaller views for the additional timezones.
	// The grid is designed to fit up to 6 timezones in a 3-column layout, with each row containing up to 3 timezones.
	itemsPerRow := 3
	// Calculates the width of each column in the grid by dividing the total width by the number of items per row.
	colWidth := maxX / itemsPerRow
	for i := 1; i < len(timezones); i++ {
		// Calculates the row and column indices for the current timezone in the grid.
		rowNum := (i - 1) / itemsPerRow
		// The column index is calculated using modulo arithmetic to ensure it wraps around after reaching the number of items per row.
		colNum := (i - 1) % itemsPerRow

		// Determines the coordinates for the current view based on its row and column position in the grid.
		// The x-coordinates (x0 and x1) are calculated based on the column index and column width,
		// while the y-coordinates (y0 and y1) are calculated based on the row index and row height.
		x0, y0 := colNum*colWidth, (rowNum+1)*rowHeight
		// Adjusts the x1 coordinate to ensure the last column in the row spans the remaining width of the screen.
		// Similarly, adjusts the y1 coordinate to ensure the last row in the grid spans the remaining height of the screen.
		x1, y1 := x0+colWidth-1, y0+rowHeight-1
		// This logic ensures that the grid layout remains consistent and fills the available space appropriately,
		// even if the number of timezones is less than the maximum capacity of the grid.
		if colNum == itemsPerRow-1 {
			// Adjusts the x1 coordinate to span the remaining width of the screen.
			x1 = maxX - 1
		}
		// If the current row is the last row in the grid, adjusts the y1 coordinate to span the
		// remaining height of the screen.
		if rowNum == 1 {
			// Adjusts the y1 coordinate to span the remaining height of the screen.
			y1 = gridMaxY - 1
		}

		// Creates a new view for the current timezone and sets its title and content.
		viewName := fmt.Sprintf("bottom%d", i)
		// If the view already exists, it is reused; otherwise, a new view is created.
		if v, err := g.SetView(viewName, x0, y0, x1, y1); err != nil && err != gocui.ErrUnknownView {
			return err
		} else {
			now := time.Now().In(locations[timezones[i].name])
			// The title is formatted to include the timezone name, the current time, and an indicator for day/night and business hours.
			v.Title = fmt.Sprintf(" [%d] %s %s %s", i, timezones[i].name, getDayNightIcon(now), getBusinessHoursIndicator(now))
			// Updates the content of the view to display the current time and date for the respective timezone.
			UpdateViewTime(v, locations[timezones[i].name])
		}
	}

	// Help footer
	// Creates a new view for the help footer at the bottom of the screen.
	// This view spans the entire width of the terminal and is positioned just above the bottom edge.
	if v, err := g.SetView("help", 0, maxY-2, maxX-1, maxY-1); err != nil {
		// If the view already exists, it is reused; otherwise, a new view is created.
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.FgColor = gocui.ColorCyan
		fmt.Fprintln(v, CenterDate("Keys [1-6]: Swap with Top | Ctrl+C: Quit", maxX))
	}
	return nil
}

/**
 * This function updates the time displayed in a specific view.
 * It takes into account the timezone associated with that view to ensure accurate time representation.
 *
 * It handles the time calculation, the blinking animation, adaptive layout for different screen sizes, and the progress bar placement.
 * The function is designed to be called every second to keep the displayed time up-to-date.
 *
 * @param v - The gocui view to update.
 * @param loc - The time.Location object representing the timezone for that view.
 */
func UpdateViewTime(v *gocui.View, loc *time.Location) {
	// Gets the current time specifically for the timezone associated with that view.
	now := time.Now().In(loc)
	// Wipes the previous frame so the new time can be drawn without leaving "ghost" characters behind.
	v.Clear()
	width, height := v.Size()

	// Blinking colon logic
	// The Modulo Operator: Checks if the current second is even or odd.
	// If it's odd, it replaces the colon with a space (03 04 PM), creating the blinking animation effect.
	format := "03:04 PM"
	if now.Second()%2 != 0 {
		format = "03 04 PM"
	}

	// Adaptive layout logic
	// This is a fail-safe for small windows (like a resized terminal or a tablet).
	// If there isn't enough vertical space for the big ASCII art, it switches to a simple, clean text format.
	if height < 8 {
		fmt.Fprintf(v, "\n%s", CenterDate(now.Format("03:04:05 PM"), width))
		fmt.Fprintf(v, "\n%s", CenterDate(now.Format("Mon, Jan 2"), width))
		// Moves the "drawing pen" to the very last line of the box to place the progress bar.
		v.SetCursor(0, height-1)
		fmt.Fprint(v, getDayProgressBar(now, width))
		return
	}

	// Converts the formatted time string into a slice of strings representing the large block characters.
	// Each line of the ASCII art is then centered horizontally within the view.
	asciiArt := PrintTimeASCII(now.Format(format))
	fmt.Fprint(v, "\n")
	for _, line := range asciiArt {
		fmt.Fprintln(v, CenterTime(line, width))
	}

	// Adds the date below the time.
	// The date is formatted in a more traditional way (Monday, January 2, 2006) and is also centered.
	// The date is bolded using ANSI escape codes.
	dateStr := fmt.Sprintf("\x1b[1m%s\x1b[0m", now.Format("Monday, January 2, 2006"))
	fmt.Fprintln(v, CenterDate(dateStr, width))

	// Adds the business hours indicator.
	fmt.Fprintln(v, CenterDate(getBusinessHoursIndicator(now), width))

	// Moves the "drawing pen" to the very last line of the box to place the progress bar.
	v.SetCursor(0, height-1)
	fmt.Fprint(v, getDayProgressBar(now, width))
}

/**
 * This function determines if a specific timezone is currently within standard
 * working hours (9:00 AM to 5:00 PM, Monday through Friday) and returns a visual status indicator.
 */
func getBusinessHoursIndicator(now time.Time) string {
	// Retrieves the current hour in a 24-hour format (0–23).
	hour := now.Hour()
	// Identifies the day of the week (Sunday through Saturday).
	weekday := now.Weekday()

	// Check if it's a weekday (Mon-Fri) and between 9 AM and 5 PM.
	// Note that hour < 17 means the green light stays on until 4:59:59 PM;
	// once it hits 5:00 PM (hour 17), it switches to "closed".
	if weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour < 17 {
		return "🟢" // Open for business
	}
	return "⚫" // Outside business hours
}

/**
 * This function calculates the percentage of the day that has passed and displays a progress bar.
 * The progress bar is color-coded to indicate different times of the day.
 * @param now - The current time.
 * @param width - The width of the progress bar.
 * @returns The progress bar as a string.
 */
func getDayProgressBar(now time.Time, width int) string {
	// 1. Calculate elapsed and remaining time
	// This converts the current time into total seconds passed since midnight.
	// Since there are exactly $86,400$ seconds in a day, dividing by this number gives a decimal percentage ($0.0$ to $1.0$).
	secondsElapsed := float64(now.Hour()*3600 + now.Minute()*60 + now.Second())
	totalSeconds := 86400.0
	percent := secondsElapsed / totalSeconds

	// Calculate remaining time
	remainingSecs := int(totalSeconds - secondsElapsed)
	remHours := remainingSecs / 3600
	remMins := (remainingSecs % 3600) / 60
	timeRemaining := fmt.Sprintf(" %dh %dm left", remHours, remMins)

	// 2. Adjust bar width to make room for the text
	// We subtract the length of the countdown string from the available width
	// It takes the total available width of the UI box and subtracts 2 to account for the leading and trailing brackets [].
	barWidth := width - 2 - len(timeRemaining)
	if barWidth < 0 {
		barWidth = 0
	}
	// Multiplies the available bar width by the percentage to determine how many "solid" blocks (█) to draw.
	fillWidth := int(float64(barWidth) * percent)

	// 3. Dynamic Color Logic
	// Green: The default color for morning and daytime. Active during standard
	// business hours (9:00 AM to 5:00 PM).
	color := "\x1b[32m"
	// Yellow: Triggered between 5:00 PM and 9:00 PM, signaling the end of the day.
	if now.Hour() >= 17 && now.Hour() < 21 {
		color = "\x1b[33m"
	}
	// Red: Triggered from 9:00 PM until 5:00 AM, indicating late-night hours.
	if now.Hour() >= 21 || now.Hour() < 5 {
		color = "\x1b[31m"
	}

	// 4. Construct the Final String
	bar := "[" + strings.Repeat("█", fillWidth) + strings.Repeat(" ", barWidth-fillWidth) + "]"
	return color + bar + timeRemaining + "\x1b[0m"
}

/**
 * This function returns a sun or moon icon based on the current time.
 * @param now - The current time.
 * @returns The sun or moon icon as a string.
 */
func getDayNightIcon(now time.Time) string {
	if now.Hour() >= 6 && now.Hour() < 18 {
		return "🌞"
	}
	return "🌙"
}

/**
 * This function centers a given string within a specified width by adding leading spaces.
 * If the string is shorter than the width, it calculates the necessary padding and adds spaces to the left.
 * If the string is longer than the width, it returns the original string without modification.
 *
 * @param s - The string to be centered.
 * @param width - The total width within which to center the string.
 * @returns The centered string with leading spaces if necessary.
 */
func CenterTime(s string, width int) string {
	// The runewidth.StringWidth function is used to calculate the display width of the string,
	// accounting for any wide characters (like emojis) that may take up more than one column in the terminal.
	pad := (width - runewidth.StringWidth(s)) / 2
	if pad > 0 {
		return strings.Repeat(" ", pad) + s
	}
	return s
}

/**
 * This function centers a given string within a specified width by adding leading spaces.
 * If the string is shorter than the width, it calculates the necessary padding and adds spaces to the left.
 * If the string is longer than the width, it returns the original string without modification.
 *
 * @param s - The string to be centered.
 * @param width - The total width within which to center the string.
 * @returns The centered string with leading spaces if necessary.
 */
func CenterDate(s string, width int) string {
	// This function is similar to CenterTime but includes a step to remove
	// ANSI escape codes (like bold formatting) from the string before calculating its width.
	clean := strings.NewReplacer("\x1b[1m", "", "\x1b[0m", "").Replace(s)
	// The runewidth.StringWidth function is used to calculate the display width of the string,
	// accounting for any wide characters (like emojis) that may take up more than one column in the terminal.
	pad := (width - runewidth.StringWidth(clean)) / 2
	// If the calculated padding is greater than zero, it adds that many spaces to the left of the string to center it.
	if pad > 0 {
		return strings.Repeat(" ", pad) + s
	}
	return s
}

/**
 * This function sets up keybindings for user interactions within the terminal UI.
 * It allows users to swap the primary timezone with any of the additional timezones by pressing keys 1-6.
 * It also binds Ctrl+C to quit the application gracefully.
 *
 * @param g - The gocui.Gui object representing the terminal UI.
 * @returns An error if any issues occur during keybinding setup.
 */
func KeyBindings(g *gocui.Gui) error {
	// Binds the Ctrl+C key combination to a function that quits the application.
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit })
	for i := 1; i <= 6; i++ {
		idx := i
		// Binds the key combination of the number key (1-6) to a function that swaps the primary timezone with the selected timezone.
		g.SetKeybinding("", rune('0'+i), gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			timezones[0], timezones[idx] = timezones[idx], timezones[0]
			return nil
		})
	}
	return nil
}

/**
 * This function converts a given time string into its ASCII art representation.
 * It iterates over each character in the time string, retrieves the corresponding ASCII art from the digits map,
 * and constructs the final ASCII art lines by combining the lines of each character.
 *
 * @param t - The time string to be converted into ASCII art.
 * @returns A slice of strings, where each string represents a line of the ASCII art.
 */
func PrintTimeASCII(t string) []string {
	// Initializes a slice of strings to hold the lines of the ASCII art.
	// Each line will be built by concatenating the corresponding lines of each character's ASCII art.
	lines := make([]string, 5)
	for _, char := range t {
		// Retrieves the ASCII art for the current character from the digits map.
		// If the character is not found in the map, it skips to the next character.
		art, ok := digits[char]
		if !ok {
			continue
		}
		// Iterates over each line of the ASCII art for the current character and appends it to the corresponding line in the lines slice.
		// Each line of the ASCII art is followed by a space to separate characters.
		for i := 0; i < 5; i++ {
			lines[i] += art[i] + " "
		}
	}
	return lines
}
