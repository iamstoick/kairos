# 🕒 Kairos - CLI Multi-Timezone Clock

A highly customizable, interactive Command Line Interface (CLI) clock built in Go. It features a 1-3-3 grid layout that displays a primary focus timezone and six secondary timezones with real-time ASCII art rendering.

## ✨ Features

- **1-3-3 Dynamic Layout**: Displays one large "Focus" clock on top and two rows of three secondary clocks below.
- **Real-time ASCII Art**: Renders time using block characters for high visibility.
- **Interactive Swapping**: Use keys 1 through 6 to instantly swap any secondary timezone into the primary top position.
- **Adaptive UI**: Automatically switches to plain text if the terminal window is too small to display ASCII art.
- **Global Standard**: Built-in support for UTC, GMT, PST, and Philippine Time.

## ⌨️ Keybindings

| Key        | Action                                                    |
| ---        | ---                                                       |
| 1 - 3      | Swap Top clock with Middle Row (Left, Center, Right)      |
| 4 - 6      | Swap Top clock with Bottom Row (Left, Center, Right)      |
| Ctrl + C   | Quit Application                                          |          

## 🚀 Installation
Ensure you have Go installed on your machine.

1. Clone the repository:
```
git clone git@github.com:iamstoick/kairos.git
cd kairos
```
2. Install dependencies:
```
go get github.com/jroimartin/gocui
go get github.com/mattn/go-runewidth
```
3. Run the application:
```
go run clock.go
```

## 🛠️ Configuration
You can easily modify the displayed timezones by editing the `timezones` slice in `clock.go`:
```
timezones = []struct {
    name     string
    location string
}{
    {"UTC", "UTC"},
    {"PST/DST", "America/Los_Angeles"},
    {"GMT", "Etc/GMT"},
    {"Philippine Time", "Asia/Manila"},
    // Add more here...
}
```

