<img width="1807" height="940" alt="image" src="https://github.com/user-attachments/assets/d3f3a6cc-01c6-4153-b99b-e2dd943c71dd" />


# 🕒 Kairos - CLI Multi-Timezone Clock
**High-Performance Terminal World Clock & System Monitor**

A highly customizable, interactive Command Line Interface (CLI) clock built in Go. It features a 1-3-3 grid layout that displays a primary focus timezone and six secondary timezones with real-time ASCII art rendering.

Kairos is a specialized CLI dashboard designed for developers and remote teams. It combines high-fidelity ASCII clocks, real-time system metrics (CPU/MEM), and an interactive timezone-swapping grid.

## ✨ Features
- **Dynamic 1-3-3 Layout**: One primary focus view and a grid for secondary timezones.
- **Interactive Swapping**: Instantly swap any secondary timezone into the primary view using keys `1-6`.
- **System Awareness**: Integrated background workers monitor CPU and Memory usage with color-coded alerts.
- **Persistence**: Save your favorite timezones locally; no need to re-configure on every launch.
- **Smart Indicators**: Visual icons for Day/Night and Business Hours (🟢/⚫).

## ⌨️ Keybindings

| Key        | Action                                                    |
| ---        | ---                                                       |
| 1 - 3      | Swap Top clock with Middle Row (Left, Center, Right)      |
| 4 - 6      | Swap Top clock with Bottom Row (Left, Center, Right)      |
| Ctrl + C   | Quit Application                                          |          

## 🚀 Installation

### Using Go CLI
Ensure you have Go installed on your machine.

```
go install [github.com/geraldvillorente/kairos@latest](https://github.com/geraldvillorente/kairos@latest)
```

### Using Source Code
1. Clone the repository:
```
git clone git@github.com:iamstoick/kairos.git
cd kairos
```
2. Install dependencies:
```
go get github.com/jroimartin/gocui
go get github.com/mattn/go-runewidth
go get github.com/shirou/gopsutil/v3/cpu
```
3. Run the application:
```
go run clock.go
```
4. Optional: Build the binary:
```
go build -o kairos clock.go     
```
Then run the binary:
```
./kairos    
```
Or move the binary to a directory in your PATH:
```
mv kairos /usr/local/bin/
```
Then run the binary:
```
kairos  
```

### Using the binary release
See the latest release here: [Releases](https://github.com/iamstoick/kairos/releases)

## 🛠️ Usage
Kairos operates as a full CLI utility. Use the following commands to manage your dashboard:

| Command	                    |    Description                                                    |
| ---	                        | ---                                                               |
| kairos	                    | Launch the interactive TUI dashboard.                             |
| kairos add "Name" "Location"	| Add a new timezone (e.g., kairos add "NYC" "America/New_York").   |
| kairos remove "Name"	        | Remove a timezone from your configuration.                        |
| kairos list	                | List all configured timezones and their IDs.                      |
| kairos help	                | Show the help menu.                                               |

## ⌨️ Dashboard Controls
- `1` - `6`: Swap the timezone at that index with the top (primary) view.
- `Ctrl + C`: Gracefully exit the application.

## 📄 License
© 2025-2026 Gerald Z. Villorente. Licensed under the MIT License.
