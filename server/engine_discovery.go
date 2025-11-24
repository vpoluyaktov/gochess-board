package server

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go-chess/utils"
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
	Type                  string            `json:"type"`                       // "uci", "cecp", "polyglot-uci", "polyglot-cecp"
	SupportsBook          bool              `json:"supportsBook"`               // true for polyglot variants
	BookPath              string            `json:"bookPath,omitempty"`         // optional opening book path
	UnderlyingEngine      string            `json:"underlyingEngine,omitempty"` // for polyglot: the actual engine being wrapped
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
func DiscoverEngines(bookFile string) []EngineInfo {
	engines := make([]EngineInfo, 0)
	seen := make(map[string]bool)

	// Check if polyglot is available first
	hasPolyglot := isPolyglotInstalled()
	if hasPolyglot {
		Info("ENGINE_DISCOVERY", "Polyglot wrapper found")
	} else {
		Info("ENGINE_DISCOVERY", "Polyglot wrapper not found, CECP engines and book support will be unavailable")
	}

	// Log book file status
	if bookFile != "" {
		Info("ENGINE_DISCOVERY", "Opening book file specified: %s", bookFile)
	} else {
		Info("ENGINE_DISCOVERY", "No opening book file specified")
	}

	// Discover UCI engines
	Info("ENGINE_DISCOVERY", "Discovering UCI engines...")
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
			Info("ENGINE_DISCOVERY", "Discovered UCI engine: %s (command: %s)%s",
				engine.Name, engine.Path, eloInfo)
		}
	}

	// Discover CECP engines only if polyglot is available
	// Store them separately - they will only be added as polyglot-wrapped variants
	var cecpEngines []EngineInfo
	if hasPolyglot {
		Info("ENGINE_DISCOVERY", "Discovering CECP engines...")
		cecpEngines = discoverEngineList(getEngineNames(cecpEngineBaseNames), getCECPEngineInfo)
		for _, engine := range cecpEngines {
			Info("ENGINE_DISCOVERY", "Discovered CECP engine: %s (command: %s)",
				engine.Name, engine.Path)
		}
	}

	// Create polyglot-wrapped variants if polyglot is available
	if hasPolyglot {
		// Only create UCI+Book variants if a book file is specified
		if bookFile != "" {
			Info("ENGINE_DISCOVERY", "Creating UCI + Book variants with opening book...")
			polyglotEngines := createPolyglotVariantsWithBook(engines, bookFile)
			engines = append(engines, polyglotEngines...)
		} else {
			Info("ENGINE_DISCOVERY", "No book file specified, skipping UCI + Book variants")
		}

		// Always create variants for CECP engines (required - only way to use them)
		if len(cecpEngines) > 0 {
			Info("ENGINE_DISCOVERY", "Creating CECP engine variants via Polyglot...")
			cecpPolyglotEngines := createPolyglotVariantsWithBook(cecpEngines, bookFile)
			engines = append(engines, cecpPolyglotEngines...)
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
		Info("ENGINE_DISCOVERY", "Engine %s is not UCI compatible (no 'uciok' response)", path)
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
		Info("ENGINE_DISCOVERY", "Engine %s is not UCI compatible (timeout)", path)
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

// isPolyglotInstalled checks if polyglot is available in the system
func isPolyglotInstalled() bool {
	_, err := exec.LookPath("polyglot")
	return err == nil
}

// createPolyglotVariantsWithBook creates polyglot-wrapped variants of discovered engines
func createPolyglotVariantsWithBook(engines []EngineInfo, bookFile string) []EngineInfo {
	variants := make([]EngineInfo, 0)

	for _, engine := range engines {
		// For UCI engines, create a polyglot variant with book support
		if engine.Type == "uci" {
			configFile, err := createPolyglotConfig(engine.Path, bookFile)
			if err != nil {
				Info("ENGINE_DISCOVERY", "Failed to create polyglot config for %s: %v", engine.Name, err)
				continue
			}

			// Create wrapper script that runs polyglot with the config file
			wrapperScript, err := createPolyglotWrapper(configFile)
			if err != nil {
				Info("ENGINE_DISCOVERY", "Failed to create polyglot wrapper for %s: %v", engine.Name, err)
				continue
			}

			variant := EngineInfo{
				Name:                  engine.Name + " + Book",
				Path:                  wrapperScript,
				Version:               engine.Version,
				ID:                    engine.ID + "_polyglot",
				Type:                  "polyglot-uci",
				SupportsBook:          true,
				BookPath:              bookFile,
				UnderlyingEngine:      engine.Path,
				SupportsLimitStrength: engine.SupportsLimitStrength,
				MinElo:                engine.MinElo,
				MaxElo:                engine.MaxElo,
				DefaultElo:            engine.DefaultElo,
				Options:               engine.Options,
			}
			variants = append(variants, variant)
			Info("ENGINE_DISCOVERY", "Created polyglot variant: %s", variant.Name)
		}

		// For CECP engines, create a polyglot variant (required to use them)
		if engine.Type == "cecp" {
			configFile, err := createPolyglotConfig(engine.Path, bookFile)
			if err != nil {
				Info("ENGINE_DISCOVERY", "Failed to create polyglot config for %s: %v", engine.Name, err)
				continue
			}

			// Create wrapper script that runs polyglot with the config file
			wrapperScript, err := createPolyglotWrapper(configFile)
			if err != nil {
				Info("ENGINE_DISCOVERY", "Failed to create polyglot wrapper for %s: %v", engine.Name, err)
				continue
			}

			variant := EngineInfo{
				Name:             engine.Name + " (via Polyglot)",
				Path:             wrapperScript,
				Version:          engine.Version,
				ID:               engine.ID + "_polyglot",
				Type:             "polyglot-cecp",
				SupportsBook:     bookFile != "",
				BookPath:         bookFile,
				UnderlyingEngine: engine.Path,
				Options:          make(map[string]string),
			}
			variants = append(variants, variant)
			Info("ENGINE_DISCOVERY", "Created polyglot variant for CECP engine: %s", variant.Name)
		}
	}

	return variants
}

// createPolyglotWrapper creates an executable wrapper script that runs polyglot with a config file
// On Unix systems, creates a shell script (.sh). On Windows, creates a batch file (.bat)
func createPolyglotWrapper(configFile string) (string, error) {
	// Create temp directory for polyglot wrappers
	tempDir := filepath.Join(os.TempDir(), "go-chess-polyglot")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create polyglot wrapper directory: %w", err)
	}

	// Create unique wrapper filename based on config file hash
	hash := fmt.Sprintf("%x", md5.Sum([]byte(configFile)))[:8]

	var wrapperFile string
	var script string

	if runtime.GOOS == "windows" {
		// Windows batch file
		wrapperFile = filepath.Join(tempDir, fmt.Sprintf("polyglot-wrapper-%s.bat", hash))
		script = fmt.Sprintf(`@echo off
polyglot "%s"
`, configFile)
	} else {
		// Unix shell script
		wrapperFile = filepath.Join(tempDir, fmt.Sprintf("polyglot-wrapper-%s.sh", hash))
		script = fmt.Sprintf(`#!/bin/sh
exec polyglot "%s"
`, configFile)
	}

	// Write wrapper script
	if err := os.WriteFile(wrapperFile, []byte(script), 0755); err != nil {
		return "", fmt.Errorf("failed to write polyglot wrapper: %w", err)
	}

	return wrapperFile, nil
}

// createPolyglotConfig creates a polyglot INI configuration file for an engine
func createPolyglotConfig(enginePath string, bookPath string) (string, error) {
	// Create temp directory for polyglot configs
	tempDir := filepath.Join(os.TempDir(), "go-chess-polyglot")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create polyglot config directory: %w", err)
	}

	// Create unique config filename based on engine path hash
	hash := fmt.Sprintf("%x", md5.Sum([]byte(enginePath)))[:8]
	configFile := filepath.Join(tempDir, fmt.Sprintf("polyglot-%s.ini", hash))

	// Prepare book settings
	useBook := "false"
	if bookPath != "" {
		useBook = "true"
	}

	// Create log file path for this engine
	logFile := filepath.Join(tempDir, fmt.Sprintf("polyglot-%s.log", hash))

	// Get absolute path to engine
	absEnginePath, err := exec.LookPath(enginePath)
	if err != nil {
		// If LookPath fails, use the enginePath as-is
		absEnginePath = enginePath
	}

	engineDir := filepath.Dir(absEnginePath)
	engineCmd := filepath.Base(absEnginePath)

	// Generate polyglot INI content
	config := fmt.Sprintf(`[PolyGlot]
EngineDir = %s
EngineCommand = %s
Book = %s
BookFile = %s
LogFile = %s
BookDepth = 255
`, engineDir, engineCmd, useBook, bookPath, logFile)

	// Write config file
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return "", fmt.Errorf("failed to write polyglot config: %w", err)
	}

	Info("ENGINE_DISCOVERY", "Created polyglot config: %s", configFile)
	return configFile, nil
}
