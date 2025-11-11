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
	defaultPort = "8080"
)

func main() {
	// Command line flags
	port := flag.String("port", defaultPort, "Port to run the web server on")
	noBrowser := flag.Bool("no-browser", false, "Don't automatically open browser")
	noTUI := flag.Bool("no-tui", false, "Don't show TUI interface")
	flag.Parse()

	addr := fmt.Sprintf("localhost:%s", *port)
	url := fmt.Sprintf("http://%s", addr)

	// Start the web server in a goroutine
	srv := server.New(addr)
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
		if err := tui.RunTUI(url); err != nil {
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
