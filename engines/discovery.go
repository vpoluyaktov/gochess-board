package engines

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gochess-board/logger"
	"gochess-board/utils"
)

// getExecutableName returns the platform-specific executable name
// On Windows, it appends .exe to the name
func getExecutableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// getEngineNames returns platform-specific engine names for discovery
func getEngineNames(baseNames []string) []string {
	names := make([]string, len(baseNames))
	for i, name := range baseNames {
		names[i] = utils.GetExecutableName(name)
	}
	return names
}

// extractVersion extracts version information from engine name
// Examples: "Stockfish 16" -> "16", "Fruit 2.1" -> "2.1", "Toga II 3.0" -> "3.0", "Crafty-23.4" -> "23.4"
func extractVersion(name string) string {
	// First try splitting by hyphen (e.g., "Crafty-23.4")
	if strings.Contains(name, "-") {
		parts := strings.Split(name, "-")
		for i := len(parts) - 1; i >= 0; i-- {
			part := strings.TrimSpace(parts[i])
			if len(part) > 0 && strings.ContainsAny(part, "0123456789") {
				// Remove common prefixes like "v"
				part = strings.TrimPrefix(part, "v")
				part = strings.TrimPrefix(part, "V")
				return part
			}
		}
	}

	// Then try splitting by spaces
	parts := strings.Fields(name)
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		// Check if this looks like a version (contains digits)
		if len(part) > 0 && (strings.ContainsAny(part, "0123456789")) {
			// Remove common prefixes like "v"
			part = strings.TrimPrefix(part, "v")
			part = strings.TrimPrefix(part, "V")
			return part
		}
	}
	return ""
}

// tryGetVersionFromBinary attempts to get version by running the binary with --version flag
func tryGetVersionFromBinary(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	// Parse first line of output
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		// Extract version from output like "GNU Chess 6.2.9"
		return utils.ExtractVersion(firstLine)
	}
	return ""
}

// EngineInfo represents information about a discovered chess engine
type EngineInfo struct {
	Name                  string            `json:"name"`
	Path                  string            `json:"path"`
	Version               string            `json:"version,omitempty"` // engine version
	ID                    string            `json:"id"`
	Type                  string            `json:"type"` // "uci" or "cecp"
	SupportsLimitStrength bool              `json:"supportsLimitStrength"`
	MinElo                int               `json:"minElo,omitempty"`
	MaxElo                int               `json:"maxElo,omitempty"`
	DefaultElo            int               `json:"defaultElo,omitempty"`
	Options               map[string]string `json:"options,omitempty"` // UCI options
}

// Common UCI engine base names to search for (without .exe extension)
var uciEngineBaseNames = []string{
	"stockfish",
	"lc0",
	"leela",
	"komodo",
	"houdini",
	"rybka",
	"fruit",
	"toga",
	"toga2",
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

// Common CECP/XBoard engine base names to search for (without .exe extension)
var cecpEngineBaseNames = []string{
	"crafty",
	"gnuchess",
	"sjeng",
	"phalanx",
}

// DiscoverEngines searches for installed UCI and CECP chess engines
// Note: Opening book support is now handled natively in Go (see polyglot_book.go)
func DiscoverEngines(bookFile string) []EngineInfo {
	engines := make([]EngineInfo, 0)
	seen := make(map[string]bool)

	// Always add the built-in internal engine first
	internalEngine := EngineInfo{
		Name:                  "GoChess (Built-in)",
		Path:                  "internal",
		Version:               "1.0",
		ID:                    "gochess-basic",
		Type:                  "internal",
		SupportsLimitStrength: false,
		Options:               make(map[string]string),
	}
	engines = append(engines, internalEngine)
	seen["internal"] = true
	logger.Info("ENGINE_DISCOVERY", "Added built-in engine: %s (ELO ~1000-1200)", internalEngine.Name)

	// Log book file status
	if bookFile != "" {
		logger.Info("ENGINE_DISCOVERY", "Opening book file specified: %s", bookFile)
	} else {
		logger.Debug("ENGINE_DISCOVERY", "No opening book file specified")
	}

	// Discover UCI engines
	logger.Debug("ENGINE_DISCOVERY", "Discovering UCI engines...")
	uciEngines := discoverEngineList(getEngineNames(uciEngineBaseNames), getEngineInfo)
	for _, engine := range uciEngines {
		if !seen[engine.Path] {
			engines = append(engines, engine)
			seen[engine.Path] = true

			eloInfo := ""
			if engine.SupportsLimitStrength {
				eloInfo = fmt.Sprintf(" [ELO: %d-%d, default: %d]",
					engine.MinElo, engine.MaxElo, engine.DefaultElo)
			}
			logger.Info("ENGINE_DISCOVERY", "Discovered UCI engine: %s (command: %s)%s",
				engine.Name, engine.Path, eloInfo)
		}
	}

	// Discover CECP engines (native CECP support, no external wrapper needed)
	logger.Debug("ENGINE_DISCOVERY", "Discovering CECP engines...")
	cecpEngines := discoverEngineList(getEngineNames(cecpEngineBaseNames), getCECPEngineInfo)
	for _, engine := range cecpEngines {
		logger.Info("ENGINE_DISCOVERY", "Discovered CECP engine: %s (command: %s)",
			engine.Name, engine.Path)
	}

	// Add CECP engines directly (native CECP support)
	if len(cecpEngines) > 0 {
		logger.Debug("ENGINE_DISCOVERY", "Adding CECP engines with native support...")
		for _, engine := range cecpEngines {
			if !seen[engine.Path] {
				engines = append(engines, engine)
				seen[engine.Path] = true
			}
		}
	}

	return engines
}

// discoverEngineList discovers engines from a list using the provided probe function
func discoverEngineList(engineNames []string, probeFunc func(string) (EngineInfo, bool)) []EngineInfo {
	type result struct {
		info EngineInfo
		ok   bool
	}

	resultChan := make(chan result, len(engineNames))

	// Probe all engines in parallel
	for _, engineName := range engineNames {
		go func(name string) {
			if info, ok := probeFunc(name); ok {
				resultChan <- result{info: info, ok: true}
			} else {
				resultChan <- result{ok: false}
			}
		}(engineName)
	}

	// Collect results
	engines := make([]EngineInfo, 0)
	for i := 0; i < len(engineNames); i++ {
		res := <-resultChan
		if res.ok {
			engines = append(engines, res.info)
		}
	}

	return engines
}

// getEngineInfo probes a UCI engine and returns comprehensive information
func getEngineInfo(path string) (EngineInfo, bool) {
	cmd := exec.Command(path)

	// Set working directory to temp to avoid cluttering the repository with log files
	tempDir := filepath.Join(os.TempDir(), "gochess-board-engines")
	if err := os.MkdirAll(tempDir, 0755); err == nil {
		cmd.Dir = tempDir
	}

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
				info.Version = utils.ExtractVersion(info.Name)
			}

			// Parse UCI options
			if strings.HasPrefix(line, "option name ") {
				parseUCIOption(line, &info)
			}

			if strings.HasPrefix(line, "uciok") {
				if info.Name == "" {
					info.Name = utils.TitleCase(filepath.Base(path))
				}
				info.ID = strings.ToLower(strings.ReplaceAll(info.Name, " ", "_"))
				info.Type = "uci"
				resultChan <- struct {
					info EngineInfo
					ok   bool
				}{info, true}
				return
			}
		}
		logger.Warn("ENGINE_DISCOVERY", "Engine %s is not UCI compatible (no 'uciok' response)", path)
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
		logger.Warn("ENGINE_DISCOVERY", "Engine %s is not UCI compatible (timeout)", path)
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

// getCECPEngineInfo probes a CECP/XBoard engine and returns basic information
func getCECPEngineInfo(path string) (EngineInfo, bool) {
	cmd := exec.Command(path)

	// Set working directory to temp to avoid cluttering the repository with log files
	tempDir := filepath.Join(os.TempDir(), "gochess-board-engines")
	if err := os.MkdirAll(tempDir, 0755); err == nil {
		cmd.Dir = tempDir
	}

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

	// Send CECP/XBoard commands
	fmt.Fprintln(stdin, "xboard")
	fmt.Fprintln(stdin, "protover 2")

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
		foundFeature := false

		for scanner.Scan() {
			line := scanner.Text()

			// CECP engines respond with "feature ..." lines
			if strings.HasPrefix(line, "feature") {
				foundFeature = true
				// Extract engine name if provided
				if strings.Contains(line, "myname=") {
					parts := strings.Split(line, "myname=")
					if len(parts) > 1 {
						name := strings.Trim(strings.Fields(parts[1])[0], "\"")
						info.Name = name
					}
				}
			}

			// CECP engines send "feature done=1" when ready
			if strings.Contains(line, "done=1") {
				if info.Name == "" {
					info.Name = utils.TitleCase(filepath.Base(path))
				}
				info.Version = utils.ExtractVersion(info.Name)
				if info.Version == "" {
					info.Version = tryGetVersionFromBinary(path)
				}
				info.ID = strings.ToLower(strings.ReplaceAll(info.Name, " ", "_"))
				info.Type = "cecp"
				resultChan <- struct {
					info EngineInfo
					ok   bool
				}{info, true}
				return
			}
		}

		// If we found feature lines but no done=1, still consider it CECP
		if foundFeature {
			if info.Name == "" {
				info.Name = utils.TitleCase(filepath.Base(path))
			}
			info.Version = utils.ExtractVersion(info.Name)
			if info.Version == "" {
				info.Version = tryGetVersionFromBinary(path)
			}
			info.ID = strings.ToLower(strings.ReplaceAll(info.Name, " ", "_"))
			info.Type = "cecp"
			resultChan <- struct {
				info EngineInfo
				ok   bool
			}{info, true}
			return
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

// Note: External Polyglot utility functions removed.
// Opening book support is now handled natively in Go via polyglot_book.go
// CECP engine support is handled natively via analysis_cecp.go
