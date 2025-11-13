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
	Name              string            `json:"name"`
	Path              string            `json:"path"`
	ID                string            `json:"id"`
	SupportsLimitStrength bool          `json:"supportsLimitStrength"`
	MinElo            int               `json:"minElo,omitempty"`
	MaxElo            int               `json:"maxElo,omitempty"`
	DefaultElo        int               `json:"defaultElo,omitempty"`
	Options           map[string]string `json:"options,omitempty"` // UCI options
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
			if info, ok := getEngineInfo(name); ok {
				resultChan <- result{
					info: info,
					ok:   true,
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
			
			// Log engine capabilities
			eloInfo := ""
			if res.info.SupportsLimitStrength {
				eloInfo = fmt.Sprintf(" [ELO: %d-%d, default: %d]", 
					res.info.MinElo, res.info.MaxElo, res.info.DefaultElo)
			}
			log.Printf("Discovered engine: %s (command: %s)%s", 
				res.info.Name, res.info.Path, eloInfo)
		}
	}

	return engines
}

// getEngineInfo probes a UCI engine and returns comprehensive information
func getEngineInfo(path string) (EngineInfo, bool) {
	cmd := exec.Command(path)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return EngineInfo{}, false
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return EngineInfo{}, false
	}

	if err := cmd.Start(); err != nil {
		return EngineInfo{}, false
	}

	// Send UCI command
	fmt.Fprintln(stdin, "uci")

	// Read response with timeout
	resultChan := make(chan struct {
		info EngineInfo
		ok   bool
	}, 1)

	go func() {
		scanner := bufio.NewScanner(stdout)
		info := EngineInfo{
			Path:    path,
			Options: make(map[string]string),
		}

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "id name ") {
				info.Name = strings.TrimPrefix(line, "id name ")
			}

			// Parse UCI options
			if strings.HasPrefix(line, "option name ") {
				parseUCIOption(line, &info)
			}

			if strings.HasPrefix(line, "uciok") {
				if info.Name == "" {
					info.Name = formatEngineName(path)
				}
				info.ID = strings.ToLower(strings.ReplaceAll(info.Name, " ", "_"))
				resultChan <- struct {
					info EngineInfo
					ok   bool
				}{info, true}
				return
			}
		}
		resultChan <- struct {
			info EngineInfo
			ok   bool
		}{EngineInfo{}, false}
	}()

	// Wait for response or timeout
	select {
	case result := <-resultChan:
		stdin.Close()
		stdout.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return result.info, result.ok
	case <-time.After(2 * time.Second):
		stdin.Close()
		stdout.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return EngineInfo{}, false
	}
}

// parseUCIOption parses a UCI option line and updates EngineInfo
func parseUCIOption(line string, info *EngineInfo) {
	// Example: "option name UCI_LimitStrength type check default false"
	// Example: "option name UCI_Elo type spin default 1350 min 1350 max 2850"

	if strings.Contains(line, "UCI_LimitStrength") {
		info.SupportsLimitStrength = true
	}

	if strings.Contains(line, "UCI_Elo") {
		parts := strings.Fields(line)
		for i, part := range parts {
			switch part {
			case "default":
				if i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &info.DefaultElo)
				}
			case "min":
				if i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &info.MinElo)
				}
			case "max":
				if i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &info.MaxElo)
				}
			}
		}
	}

	// Store the full option for future use
	if strings.HasPrefix(line, "option name ") {
		optionStr := strings.TrimPrefix(line, "option name ")
		parts := strings.SplitN(optionStr, " type ", 2)
		if len(parts) == 2 {
			optionName := parts[0]
			info.Options[optionName] = line
		}
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
