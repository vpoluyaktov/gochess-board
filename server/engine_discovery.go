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
	type result struct {
		info EngineInfo
		ok   bool
	}
	
	resultChan := make(chan result, len(commonEngineNames))
	
	// Probe all engines in parallel
	for _, engineName := range commonEngineNames {
		go func(name string) {
			if engineName, ok := probeUCIEngine(name); ok {
				id := strings.ToLower(strings.ReplaceAll(engineName, " ", "_"))
				resultChan <- result{
					info: EngineInfo{
						Name: engineName,
						Path: name,
						ID:   id,
					},
					ok: true,
				}
			} else {
				resultChan <- result{ok: false}
			}
		}(engineName)
	}
	
	// Collect results
	engines := make([]EngineInfo, 0)
	seen := make(map[string]bool)
	
	for i := 0; i < len(commonEngineNames); i++ {
		res := <-resultChan
		if res.ok && !seen[res.info.Path] {
			engines = append(engines, res.info)
			seen[res.info.Path] = true
			log.Printf("Discovered engine: %s (command: %s)", res.info.Name, res.info.Path)
		}
	}

	return engines
}

// probeUCIEngine checks if the given path is a valid UCI chess engine and returns its name
func probeUCIEngine(path string) (string, bool) {
	cmd := exec.Command(path)
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", false
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", false
	}
	
	if err := cmd.Start(); err != nil {
		return "", false
	}
	
	// Send UCI command
	fmt.Fprintln(stdin, "uci")
	
	// Read response with timeout
	resultChan := make(chan struct {
		name string
		ok   bool
	}, 1)
	
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
					resultChan <- struct {
						name string
						ok   bool
					}{engineName, true}
				} else {
					resultChan <- struct {
						name string
						ok   bool
					}{formatEngineName(path), true}
				}
				return
			}
		}
		resultChan <- struct {
			name string
			ok   bool
		}{"", false}
	}()
	
	// Wait for response or timeout
	select {
	case result := <-resultChan:
		stdin.Close()
		stdout.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return result.name, result.ok
	case <-time.After(500 * time.Millisecond):
		stdin.Close()
		stdout.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return "", false
	}
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
