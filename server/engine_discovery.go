package server

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// EngineInfo represents information about a discovered chess engine
type EngineInfo struct {
	Name string
	Path string
	ID   string // Unique identifier for the engine
}

// Common UCI engine names to search for
var commonEngineNames = []string{
	"stockfish",
	"lc0",
	"leela",
	"komodo",
	"houdini",
	"rybka",
	"crafty",
	"gnuchess",
	"fruit",
	"toga",
	"glaurung",
	"fairy-stockfish",
	"ethereal",
	"berserk",
	"koivisto",
	"rubichess",
	"slowchess",
	"igel",
	"laser",
	"demolito",
}

// DiscoverEngines searches for installed UCI chess engines
func DiscoverEngines() []EngineInfo {
	engines := make([]EngineInfo, 0)
	seen := make(map[string]bool)

	// Try to run each engine and check if it responds to UCI
	for _, engineName := range commonEngineNames {
		if seen[engineName] {
			continue
		}

		// Try to run the engine and verify it's a UCI engine
		if isUCIEngine(engineName) {
			name := getEngineName(engineName, engineName)
			id := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
			engines = append(engines, EngineInfo{
				Name: name,
				Path: engineName, // Store just the command name
				ID:   id,
			})
			seen[engineName] = true
			log.Printf("Discovered engine: %s (command: %s)", name, engineName)
		}
	}

	return engines
}

// isUCIEngine checks if the given path is a valid UCI chess engine
func isUCIEngine(path string) bool {
	cmd := exec.Command(path)
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false
	}
	
	if err := cmd.Start(); err != nil {
		return false
	}
	
	// Send UCI command
	fmt.Fprintln(stdin, "uci")
	
	// Read response with timeout
	responseChan := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "uciok") {
				responseChan <- true
				return
			}
		}
		responseChan <- false
	}()
	
	// Wait for response or timeout
	select {
	case isUCI := <-responseChan:
		stdin.Close()
		stdout.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return isUCI
	case <-time.After(2 * time.Second):
		stdin.Close()
		stdout.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return false
	}
}

// getEngineName tries to get the engine's name from UCI id command
func getEngineName(path string, fallbackName string) string {
	cmd := exec.Command(path)
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return formatEngineName(fallbackName)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return formatEngineName(fallbackName)
	}
	
	if err := cmd.Start(); err != nil {
		return formatEngineName(fallbackName)
	}
	
	// Send UCI command
	fmt.Fprintln(stdin, "uci")
	
	// Read response with timeout
	nameChan := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		engineName := ""
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "id name ") {
				engineName = strings.TrimPrefix(line, "id name ")
			}
			if strings.HasPrefix(line, "uciok") {
				if engineName != "" {
					nameChan <- engineName
				} else {
					nameChan <- formatEngineName(fallbackName)
				}
				return
			}
		}
		nameChan <- formatEngineName(fallbackName)
	}()
	
	// Wait for response or timeout
	select {
	case name := <-nameChan:
		stdin.Close()
		stdout.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return name
	case <-time.After(2 * time.Second):
		stdin.Close()
		stdout.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return formatEngineName(fallbackName)
	}
}

// formatEngineName formats the engine name for display
func formatEngineName(name string) string {
	// Capitalize first letter of each word
	words := strings.Split(name, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
