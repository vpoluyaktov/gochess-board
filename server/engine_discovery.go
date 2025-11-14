package server

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// EngineInfo represents information about a discovered chess engine
type EngineInfo struct {
	Name                  string            `json:"name"`
	Path                  string            `json:"path"`
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

// Common UCI engine names to search for
var uciEngineNames = []string{
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

// Common CECP/XBoard engine names to search for
var cecpEngineNames = []string{
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
		log.Printf("[ENGINE_DISCOVERY] Polyglot wrapper found")
	} else {
		log.Printf("[ENGINE_DISCOVERY] Polyglot wrapper not found, CECP engines and book support will be unavailable")
	}

	// Log book file status
	if bookFile != "" {
		log.Printf("[ENGINE_DISCOVERY] Opening book file specified: %s", bookFile)
	} else {
		log.Printf("[ENGINE_DISCOVERY] No opening book file specified")
	}

	// Discover UCI engines
	log.Printf("[ENGINE_DISCOVERY] Discovering UCI engines...")
	uciEngines := discoverEngineList(uciEngineNames, getEngineInfo)
	for _, engine := range uciEngines {
		if !seen[engine.Path] {
			engines = append(engines, engine)
			seen[engine.Path] = true

			eloInfo := ""
			if engine.SupportsLimitStrength {
				eloInfo = fmt.Sprintf(" [ELO: %d-%d, default: %d]",
					engine.MinElo, engine.MaxElo, engine.DefaultElo)
			}
			log.Printf("[ENGINE_DISCOVERY] Discovered UCI engine: %s (command: %s)%s",
				engine.Name, engine.Path, eloInfo)
		}
	}

	// Discover CECP engines only if polyglot is available
	// Store them separately - they will only be added as polyglot-wrapped variants
	var cecpEngines []EngineInfo
	if hasPolyglot {
		log.Printf("[ENGINE_DISCOVERY] Discovering CECP engines...")
		cecpEngines = discoverEngineList(cecpEngineNames, getCECPEngineInfo)
		for _, engine := range cecpEngines {
			log.Printf("[ENGINE_DISCOVERY] Discovered CECP engine: %s (command: %s)",
				engine.Name, engine.Path)
		}
	}

	// Create polyglot-wrapped variants if polyglot is available
	if hasPolyglot {
		// Only create UCI+Book variants if a book file is specified
		if bookFile != "" {
			log.Printf("[ENGINE_DISCOVERY] Creating UCI + Book variants with opening book...")
			polyglotEngines := createPolyglotVariantsWithBook(engines, bookFile)
			engines = append(engines, polyglotEngines...)
		} else {
			log.Printf("[ENGINE_DISCOVERY] No book file specified, skipping UCI + Book variants")
		}

		// Always create variants for CECP engines (required - only way to use them)
		if len(cecpEngines) > 0 {
			log.Printf("[ENGINE_DISCOVERY] Creating CECP engine variants via Polyglot...")
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
				info.Type = "uci"
				resultChan <- struct {
					info EngineInfo
					ok   bool
				}{info, true}
				return
			}
		}
		log.Printf("[ENGINE_DISCOVERY] Engine %s is not UCI compatible (no 'uciok' response)", path)
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
		log.Printf("[ENGINE_DISCOVERY] Engine %s is not UCI compatible (timeout)", path)
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
					info.Name = formatEngineName(path)
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
				info.Name = formatEngineName(path)
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
				log.Printf("[ENGINE_DISCOVERY] Failed to create polyglot config for %s: %v", engine.Name, err)
				continue
			}

			variant := EngineInfo{
				Name:                  engine.Name + " + Book",
				Path:                  configFile,
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
			log.Printf("[ENGINE_DISCOVERY] Created polyglot variant: %s", variant.Name)
		}

		// For CECP engines, create a polyglot variant (required to use them)
		if engine.Type == "cecp" {
			configFile, err := createPolyglotConfig(engine.Path, bookFile)
			if err != nil {
				log.Printf("[ENGINE_DISCOVERY] Failed to create polyglot config for %s: %v", engine.Name, err)
				continue
			}

			variant := EngineInfo{
				Name:             engine.Name + " (via Polyglot)",
				Path:             configFile,
				ID:               engine.ID + "_polyglot",
				Type:             "polyglot-cecp",
				SupportsBook:     bookFile != "",
				BookPath:         bookFile,
				UnderlyingEngine: engine.Path,
				Options:          make(map[string]string),
			}
			variants = append(variants, variant)
			log.Printf("[ENGINE_DISCOVERY] Created polyglot variant for CECP engine: %s", variant.Name)
		}
	}

	return variants
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

	// Generate polyglot INI content
	config := fmt.Sprintf(`[PolyGlot]
EngineDir = %s
EngineCommand = %s
Book = %s
BookFile = %s
LogFile = %s
BookDepth = 255
`, filepath.Dir(enginePath), filepath.Base(enginePath), useBook, bookPath, logFile)

	// Write config file
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return "", fmt.Errorf("failed to write polyglot config: %w", err)
	}

	log.Printf("[ENGINE_DISCOVERY] Created polyglot config: %s", configFile)
	return configFile, nil
}
