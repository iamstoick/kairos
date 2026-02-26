// Author Name: Gerald Z. Villorente
// Author email: geraldvillorente@gmail.com
// @2025-2026
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/shirou/gopsutil/v3/cpu"
)

// TimezoneConfig defines the structure for saved timezones.
// Fields must be capitalized to be exported for JSON encoding.
type TimezoneConfig struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

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

	timezones []TimezoneConfig

	currentCPU        string
	currentMEM        string
	notification      string
	notificationTimer *time.Timer
)

func main() {
	// Load the configuration file first to populate the
	// timezones variable with any saved settings from previous runs.
	loadConfig()

	// Check for command-line arguments to add or remove timezones before starting the GUI.
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "help":
			printHelp()
			return
		case "list":
			printList()
			return
		case "add":
			if len(os.Args) != 4 {
				fmt.Println("Usage: kairos add \"Name\" \"Location/City\"")
				return
			}
			// Add to slice using the named TimezoneConfig type and save
			timezones = append(timezones, TimezoneConfig{
				Name:     os.Args[2],
				Location: os.Args[3],
			})
			saveConfig()
			fmt.Printf("Added %s successfully!\n", os.Args[2])
			return

		case "remove":
			if len(os.Args) != 3 {
				fmt.Println("Usage: kairos remove \"Name\"")
				return
			}

			// Create a new slice of the SAME type to store remaining zones
			var newList []TimezoneConfig
			found := false
			for _, tz := range timezones {
				if tz.Name != os.Args[2] {
					newList = append(newList, tz)
				} else {
					found = true
				}
			}

			if !found {
				fmt.Printf("Timezone '%s' not found.\n", os.Args[2])
				return
			}

			timezones = newList
			saveConfig()
			fmt.Printf("Removed %s successfully!\n", os.Args[2])
			return
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Type 'kairos help' for usage instructions.")
			return
		}
	}

	// If no command-line arguments are provided, it proceeds to run the terminal-based GUI application.
	runGUI()
}

/**
 * This function initializes and runs the terminal-based GUI application using the gocui library.
 * It sets up the GUI, loads timezone locations, defines the layout, keybindings, and starts the main event loop.
 */
func runGUI() {
	if len(timezones) == 0 {
		fmt.Println("No timezones configured. Use: kairos add \"Name\" \"Location\"")
		fmt.Println("Example: kairos add \"PHL\" \"Asia/Manila\"")
		return
	}

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
		loc, err := time.LoadLocation(tz.Location)
		if err != nil {
			continue // Skip invalid ones from config
		}
		// Stores the loaded location in the locations map with the timezone name as the key.
		locations[tz.Name] = loc
	}

	// Set the layout function that will be called to draw the UI.
	g.SetManagerFunc(layout)
	// Set up keybindings for user interactions (swapping timezones and quitting the application).
	if err := KeyBindings(g); err != nil {
		log.Panicln("Failed to create keybindings: ", err)
	}

	// Start the stats worker to update CPU and memory usage.
	startStatsWorker()

	// Update the UI every second to reflect the current time.
	go func() {
		// Creates a ticker that sends a value on a channel every second.
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			// Calls the Update method of the GUI to trigger a redraw of the UI.
			g.Update(func(g *gocui.Gui) error { return nil })
		}
	}()

	// Start the main event loop for the GUI.
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
	// Reserves the bottom lines of the terminal so the "Help Footer" doesn't overlap.
	gridMaxY := maxY - 3
	// Divides the available height into horizontal sections.
	rowHeight := gridMaxY / 3

	// Top View (Index 0)
	if v, err := g.SetView("top", 0, 0, maxX-1, rowHeight-1); err != nil && err != gocui.ErrUnknownView {
		return err
	} else {
		// Gets the current time for the primary timezone and sets the title.
		loc, ok := locations[timezones[0].Name]
		if ok {
			// Gets the current time for the primary timezone (UTC) and sets the title of the top view
			// to include the timezone name, a day/night icon, and the business hours indicator.
			now := time.Now().In(locations[timezones[0].Name])
			// The title format is: " UTC 🌞 🟢" (for example), where the icon and business hours indicator change based on the current time.
			icon := getDayNightIcon(now)
			// The business hours indicator is determined by the getBusinessHoursIndicator function,
			// which checks if the current time falls within standard working hours.
			biz := getBusinessHoursIndicator(now)
			// Sets the title of the top view to display the timezone name, day/night icon, and business hours indicator.
			v.Title = fmt.Sprintf(" %s %s %s", timezones[0].Name, icon, biz)
			// Updates the content of the top view to display the current time and date in the primary timezone.
			UpdateViewTime(v, loc)
		}
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
			loc, ok := locations[timezones[i].Name]
			if ok {
				now := time.Now().In(loc)
				// The title is formatted to include the timezone name, the current time, and an indicator for day/night and business hours.
				v.Title = fmt.Sprintf(" [%d] %s %s %s", i, timezones[i].Name, getDayNightIcon(now), getBusinessHoursIndicator(now))
				// Updates the content of the view to display the current time and date for the respective timezone.
				UpdateViewTime(v, loc)
			}
		}
	}

	// Help footer
	// Creates a new view for the help footer at the bottom of the screen.
	// This view spans the entire width of the terminal and is positioned just above the bottom edge.
	if v, err := g.SetView("help", -1, maxY-3, maxX, maxY-1); err != nil {
		// If the view already exists, it is reused; otherwise, a new view is created.
		if err != gocui.ErrUnknownView {
			return err
		}
		// Sets the frame and colors for the help footer view.
		v.Frame = false
		v.FgColor = gocui.ColorCyan
		v.BgColor = gocui.ColorDefault
	}
	// Updates the content of the help footer to display instructions for user interactions and the last update time.
	if v, err := g.View("help"); err == nil {
		v.Clear()
		v.SetCursor(0, 0)

		// Get the current time for the heartbeat display in the footer.
		heartbeat := time.Now().Format("15:04:05")
		statusPart := fmt.Sprintf("%s | %s", currentCPU, currentMEM)

		// If there is a notification, it is displayed in yellow and bold.
		if notification != "" {
			statusPart = fmt.Sprintf("\x1b[33m\x1b[1m %s \x1b[0m", notification)
		}

		// The footer text includes instructions for swapping timezones, quitting the application, and displays the current CPU and memory usage along with a heartbeat timestamp.
		footerText := fmt.Sprintf("Keys [1-6] to swap timezones | Ctrl+C to quit | %s %s", statusPart, heartbeat)

		// Use Fprint instead of Fprintln to avoid an extra newline
		// that might trigger a scroll-down in a 1-line view.
		fmt.Fprint(v, CenterDate(footerText, maxX))
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
 *
 * @param {time.Time} now - The current time in the timezone to check.
 * @return {string} - A visual indicator (🟢 for business hours, ⚫ for non-business hours).
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
 * This function determines if a specific timezone is currently within standard
 * working hours (9:00 AM to 5:00 PM, Monday through Friday) and returns a visual status indicator.
 *
 * @param {time.Time} now - The current time in the timezone to check.
 * @param {int} width - The width of the terminal window. This is used to calculate the size of the progress bar.
 * @return {string} - A visual indicator (🟢 for business hours, ⚫ for non-business hours).
 */
func getDayProgressBar(now time.Time, width int) string {
	// 1. Calculate elapsed and remaining time
	// This converts the current time into total seconds passed since midnight.
	// Since there are exactly $86,400$ seconds in a day, dividing by this number gives a decimal percentage ($0.0$ to $1.0$).
	secondsElapsed := float64(now.Hour()*3600 + now.Minute()*60 + now.Second())
	totalSeconds := 86400.0
	percent := secondsElapsed / totalSeconds

	// Calculate remaining time in hours and minutes for the time remaining display.
	remainingSecs := int(totalSeconds - secondsElapsed)
	timeRemaining := fmt.Sprintf(" %dh %dm left", remainingSecs/3600, (remainingSecs%3600)/60)

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

	// 4. Construct the final string.
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
 * This function displays a notification message for 3 seconds.
 * @param msg - The message to display.
 */
func showNotification(msg string) {
	notification = msg
	if notificationTimer != nil {
		notificationTimer.Stop()
	}
	// Set a timer to clear the notification after 3 seconds.
	notificationTimer = time.AfterFunc(3*time.Second, func() {
		notification = ""
	})
}

/**
 * This function starts a worker goroutine that periodically updates the CPU and memory usage statistics.
 * The worker runs every 2 seconds and updates the global variables `currentCPU` and `currentMEM` with the latest statistics.
 */
func startStatsWorker() {
	// Start a goroutine to update CPU and memory usage every 2 seconds
	go func() {
		// Initialize CPU usage to avoid showing "0.0%" on the first run
		currentCPU = "CPU: Calculating..."
		currentMEM = "MEM: Calculating..."
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			percentages, _ := cpu.Percent(0, false)
			if len(percentages) > 0 {
				usage := percentages[0]
				// Set the color to green by default.
				color := "\x1b[32m"
				// If CPU usage exceeds 50%, change the color to yellow to indicate moderate usage.
				if usage > 50 {
					color = "\x1b[33m"
				}
				// If CPU usage exceeds 80%, change the color to red to indicate high usage.
				if usage > 80 {
					color = "\x1b[31m"
				}
				currentCPU = fmt.Sprintf("CPU: %s%.1f%%\x1b[0m", color, usage)
			}

			// Update memory usage
			var m runtime.MemStats
			// Reads the current memory statistics into the MemStats struct.
			runtime.ReadMemStats(&m)
			// Calculates the percentage of memory used by dividing the allocated
			// memory (Alloc) by the total system memory (Sys) and multiplying by 100.
			usagePercent := float64(m.Alloc) / float64(m.Sys) * 100
			// Set the color to green by default.
			color := "\x1b[32m"
			// If memory usage exceeds 50%, change the color to yellow to indicate moderate usage.
			if usagePercent > 50 {
				color = "\x1b[33m"
			}
			// If memory usage exceeds 80%, change the color to red to indicate high usage.
			currentMEM = fmt.Sprintf("MEM: %s%dMB\x1b[0m", color, m.Alloc/1024/1024)
		}
	}()
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
	repl := strings.NewReplacer("\x1b[1m", "", "\x1b[0m", "", "\x1b[33m", "", "\x1b[32m", "", "\x1b[31m", "")
	clean := repl.Replace(s)
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
			if idx >= len(timezones) {
				return nil
			}
			oldTop := timezones[0].Name
			timezones[0], timezones[idx] = timezones[idx], timezones[0]
			// After swapping, it updates the locations map to reflect the new primary timezone.
			showNotification(fmt.Sprintf("Swapped %s with %s", oldTop, timezones[0].Name))
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

/**
 * Retrieves the path to the configuration file in the user's home directory.
 *
 * @returns The full path to the configuration file.
 */
func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kairos_config.json")
}

/**
 * Saves the current timezones configuration to a JSON file in the user's home directory.
 */
func saveConfig() {
	data, _ := json.Marshal(timezones)
	os.WriteFile(getConfigPath(), data, 0644)
}

/**
 * Loads the timezones configuration from a JSON file in the user's home directory.
 */
func loadConfig() {
	// Attempts to read the configuration file from the user's home directory.
	data, err := os.ReadFile(getConfigPath())
	if err == nil {
		// If the file is successfully read, it unmarshals the JSON data into the timezones slice.
		json.Unmarshal(data, &timezones)
	}
}

/**
 * This function prints the command-line usage instructions for the Kairos application.
 * It guides users on how to add, remove, and launch the timezone dashboard.
 */
func printHelp() {
	fmt.Println("\n\x1b[36m\x1b[1mKAIROS - World Clock Dashboard\x1b[0m")
	fmt.Println("A terminal-based timezone monitor and system health dashboard.")
	fmt.Println("\n\x1b[1mUSAGE:\x1b[0m")
	fmt.Println("  kairos              \x1b[90m# Launches the dashboard\x1b[0m")
	fmt.Println("  kairos help         \x1b[90m# Shows this help menu\x1b[0m")
	fmt.Println("  kairos list         \x1b[90m# Lists all saved timezones\x1b[0m")
	fmt.Println("  kairos add [N] [L]  \x1b[90m# Adds a new timezone\x1b[0m")
	fmt.Println("  kairos remove [N]   \x1b[90m# Removes a timezone\x1b[0m")

	fmt.Println("\n\x1b[1mADD ARGUMENTS:\x1b[0m")
	fmt.Println("  \x1b[33m[N]\x1b[0m : Display Name (e.g., \"Manila\", \"NYC\")")
	fmt.Println("  \x1b[33m[L]\x1b[0m : IANA Location (e.g., \"Asia/Manila\", \"America/New_York\")")

	fmt.Println("\n\x1b[1mEXAMPLES:\x1b[0m")
	fmt.Println("  kairos add \"Tokyo\" \"Asia/Tokyo\"")
	fmt.Println("  kairos remove \"Tokyo\"")

	fmt.Println("\n\x1b[1mCONTROLS (Inside Dashboard):\x1b[0m")
	fmt.Println("  • \x1b[32mKeys 1-6\x1b[0m : Swap secondary timezone with the primary (top) view.")
	fmt.Println("  • \x1b[31mCtrl+C\x1b[0m   : Quit the application.")
	fmt.Println()
}

/**
 * This function displays a list of all currently configured timezones in a table format.
 * It helps users verify their settings before launching the dashboard.
 */
func printList() {
	if len(timezones) == 0 {
		fmt.Println("\x1b[31mNo timezones configured.\x1b[0m Use 'kairos help' to see how to add some.")
		return
	}

	fmt.Println("\n\x1b[36m\x1b[1mCONFIGURED TIMEZONES\x1b[0m")
	fmt.Printf("%-5s %-15s %-25s\n", "ID", "NAME", "IANA LOCATION")
	fmt.Println(strings.Repeat("-", 45))

	for i, tz := range timezones {
		label := fmt.Sprintf(" %d", i)
		// Mark the Primary/Top timezone with a green [P] label for easy identification.
		if i == 0 {
			label = "\x1b[32m[P]  \x1b[0m"
		}
		fmt.Printf("%-5s %-15s %-25s\n", label, tz.Name, tz.Location)
	}
	fmt.Println("\x1b[90m(P) = Primary Timezone (Top View)\x1b[0m")
}
