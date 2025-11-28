package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gochess-board/engines"
	"gochess-board/engines/builtin"
	"gochess-board/logger"
	"gochess-board/server"
	"gochess-board/tui"
	"gochess-board/utils"
)

const (
	defaultPort = "35256"
)

func main() {
	// Customize usage to show double dashes (GNU-style)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "gochess-board")
		fmt.Fprintf(flag.CommandLine.Output(), "  --port string\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Port to run the web server on (default %q)\n", defaultPort)
		fmt.Fprintf(flag.CommandLine.Output(), "  --no-browser\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Don't automatically open browser\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  --no-tui\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Don't show TUI interface\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  --restart\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Kill any existing gochess-board process before starting\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  --book-file string\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Path to opening book file for polyglot (optional)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  --log-level string\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Log level: DEBUG, INFO, WARN, ERROR (default \"INFO\")\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  --engine-only\n")
		fmt.Fprintf(flag.CommandLine.Output(), "        Run built-in engine in UCI protocol mode\n")
	}

	// Command line flags
	port := flag.String("port", defaultPort, "Port to run the web server on")
	noBrowser := flag.Bool("no-browser", false, "Don't automatically open browser")
	noTUI := flag.Bool("no-tui", false, "Don't show TUI interface")
	restart := flag.Bool("restart", false, "Kill any existing gochess-board process before starting")
	bookFile := flag.String("book-file", "", "Path to opening book file for polyglot (optional)")
	logLevel := flag.String("log-level", "INFO", "Log level: DEBUG, INFO, WARN, ERROR")
	engineOnly := flag.Bool("engine-only", false, "Run built-in engine in UCI protocol mode")
	flag.Parse()

	// Initialize debug logging to file
	if err := server.InitDebugLogging("gochess.log"); err != nil {
		fmt.Printf("Warning: Failed to initialize debug logging: %v\n", err)
	}

	// Set log level from command line
	logger.SetLogLevel(*logLevel)

	// Handle engine-only mode (UCI protocol)
	if *engineOnly {
		builtin.RunUCI()
		return
	}

	// Handle restart flag - kill process using our port
	if *restart {
		err := utils.KillProcessOnPort(*port)
		if err == nil {
			// Successfully killed a process
			fmt.Printf("Killed process using port %s\n", *port)
			// Give process time to clean up
			time.Sleep(500 * time.Millisecond)
		} else if !strings.Contains(err.Error(), "no process found") {
			// Real error (not just "no process found")
			fmt.Printf("Warning: Failed to kill process on port %s: %v\n", *port, err)
		}
		// If no process found, silently continue
	}

	addr := fmt.Sprintf(":%s", *port)
	url := fmt.Sprintf("http://localhost:%s", *port)

	// Display startup message
	fmt.Println("Discovering chess engines...")
	fmt.Println("Loading opening database...")

	// Start the web server in a goroutine
	srv := server.New(addr, *bookFile)

	// Display book information if loaded
	if bookLoaded, entryCount := srv.GetPolyglotBookInfo(); bookLoaded {
		fmt.Printf("Polyglot opening book loaded: %d entries\n", entryCount)
	} else if *bookFile != "" {
		fmt.Println("Warning: Book file specified but failed to load")
	}

	fmt.Println("Server initialized successfully!")

	// Channel to receive server startup errors
	errChan := make(chan error, 1)

	go func() {
		if err := srv.Start(); err != nil {
			// Print user-friendly error to stderr
			fmt.Fprintf(os.Stderr, "\nError: Failed to start server on %s\n", url)
			fmt.Fprintf(os.Stderr, "Reason: %v\n", err)
			fmt.Fprintf(os.Stderr, "\nThis usually means the port is already in use.\n")
			fmt.Fprintf(os.Stderr, "Try using a different port with: go run . --port <port_number>\n")
			errChan <- err
		}
	}()

	// Give the server a moment to start and check for errors
	time.Sleep(500 * time.Millisecond)

	// Check if server failed to start
	select {
	case <-errChan:
		os.Exit(1)
	default:
		// Server started successfully
	}

	// Open browser automatically unless disabled
	if !*noBrowser {
		fmt.Printf("Opening %s in your browser...\n", url)
		if err := utils.OpenBrowser(url); err != nil {
			log.Printf("Failed to open browser: %v", err)
			fmt.Printf("Please open %s manually\n", url)
		}
	}

	// Run TUI or simple output
	if !*noTUI {
		// Give browser time to open before starting TUI
		time.Sleep(1 * time.Second)

		// Get book info for TUI
		bookLoaded, bookEntries := srv.GetPolyglotBookInfo()

		// Run TUI (blocks until quit)
		if err := tui.RunTUI(url, srv.GetEngines(), engines.GlobalMonitor, srv.GetOpeningStats(), bookLoaded, bookEntries); err != nil {
			log.Fatalf("TUI error: %v", err)
		}
	} else {
		fmt.Printf("Chess board server running at %s\n", url)
		fmt.Println("Press Ctrl+C to stop")

		// Block forever
		select {}
	}
}
