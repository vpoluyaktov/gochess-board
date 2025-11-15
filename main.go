package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"go-chess/server"
	"go-chess/tui"
)

const (
	defaultPort = "35256"
)

func main() {
	// Customize usage to show double dashes (GNU-style)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "go-chess")
		fmt.Fprintf(flag.CommandLine.Output(), "  --port string\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Port to run the web server on (default %q)\n", defaultPort)
		fmt.Fprintf(flag.CommandLine.Output(), "  --no-browser\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Don't automatically open browser\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  --no-tui\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Don't show TUI interface\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  --book-file string\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Path to opening book file for polyglot (optional)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  --log-level string\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Log level: DEBUG, INFO, WARN, ERROR (default \"INFO\")\n")
	}

	// Command line flags
	port := flag.String("port", defaultPort, "Port to run the web server on")
	noBrowser := flag.Bool("no-browser", false, "Don't automatically open browser")
	noTUI := flag.Bool("no-tui", false, "Don't show TUI interface")
	bookFile := flag.String("book-file", "", "Path to opening book file for polyglot (optional)")
	logLevel := flag.String("log-level", "INFO", "Log level: DEBUG, INFO, WARN, ERROR")
	flag.Parse()

	// Initialize debug logging to file
	if err := server.InitDebugLogging("chess-debug.log"); err != nil {
		fmt.Printf("Warning: Failed to initialize debug logging: %v\n", err)
	}

	// Set log level from command line
	server.SetLogLevel(*logLevel)

	addr := fmt.Sprintf(":%s", *port)
	url := fmt.Sprintf("http://localhost:%s", *port)

	// Display startup message
	fmt.Println("Discovering chess engines...")
	fmt.Println("Loading opening database...")

	// Start the web server in a goroutine
	srv := server.New(addr, *bookFile)
	fmt.Println("Server initialized successfully!")
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(500 * time.Millisecond)

	// Open browser automatically unless disabled
	if !*noBrowser {
		fmt.Printf("Opening %s in your browser...\n", url)
		if err := openBrowser(url); err != nil {
			log.Printf("Failed to open browser: %v", err)
			fmt.Printf("Please open %s manually\n", url)
		}
	}

	// Run TUI or simple output
	if !*noTUI {
		// Give browser time to open before starting TUI
		time.Sleep(1 * time.Second)

		// Run TUI (blocks until quit)
		if err := tui.RunTUI(url, srv.GetEngines(), server.GlobalMonitor, srv.GetOpeningStats()); err != nil {
			log.Fatalf("TUI error: %v", err)
		}
	} else {
		fmt.Printf("Chess board server running at %s\n", url)
		fmt.Println("Press Ctrl+C to stop")

		// Block forever
		select {}
	}
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
