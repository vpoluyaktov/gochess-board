package analysis

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gochess-board/logger"
)

// AnalysisEngine manages a UCI engine for analysis
type AnalysisEngine struct {
	cmd       *exec.Cmd
	stdin     *bufio.Writer
	stdout    *bufio.Scanner
	mu        sync.Mutex
	active    bool
	multiPVs  map[int]AnalysisInfo // Store multiple PV lines indexed by multipv number
	multiPVMu sync.Mutex           // Mutex for multiPVs map
}

// NewAnalysisEngine creates a new UCI analysis engine
func NewAnalysisEngine(enginePath string) (*AnalysisEngine, error) {
	cmd := exec.Command(enginePath)

	// Set working directory to temp to avoid cluttering the repository with log files
	// os.TempDir() returns the appropriate temp directory for the OS:
	// - Linux/macOS: /tmp
	// - Windows: %TEMP% (e.g., C:\Users\<user>\AppData\Local\Temp)
	tempDir := filepath.Join(os.TempDir(), "gochess-board-engines")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.Warn("ANALYSIS", "Failed to create temp directory for engine: %v", err)
	} else {
		cmd.Dir = tempDir
		logger.Debug("ANALYSIS", "Engine working directory set to: %s", tempDir)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Create scanner with larger buffer to handle long PV lines
	scanner := bufio.NewScanner(stdout)
	// Increase buffer size to 1MB to handle very long PV lines from deep analysis
	const maxScanTokenSize = 1024 * 1024 // 1MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)
	
	engine := &AnalysisEngine{
		cmd:      cmd,
		stdin:    bufio.NewWriter(stdin),
		stdout:   scanner,
		active:   true,
		multiPVs: make(map[int]AnalysisInfo),
	}

	logger.Info("ANALYSIS", "Initializing UCI analysis engine: %s", enginePath)

	// Initialize UCI
	engine.sendCommand("uci")

	// Wait for uciok
	logger.Info("ANALYSIS", "Waiting for uciok...")
	for engine.stdout.Scan() {
		line := engine.stdout.Text()
		logger.Debug("ANALYSIS", "<<< %s", line)
		if strings.HasPrefix(line, "uciok") {
			logger.Info("ANALYSIS", "Got uciok")
			break
		}
	}

	// Set MultiPV option to 3 for getting 3 best moves
	engine.sendCommand("setoption name MultiPV value 3")

	engine.sendCommand("isready")
	logger.Info("ANALYSIS", "Waiting for readyok...")
	for engine.stdout.Scan() {
		line := engine.stdout.Text()
		logger.Debug("ANALYSIS", "<<< %s", line)
		if strings.HasPrefix(line, "readyok") {
			logger.Info("ANALYSIS", "Got readyok, engine ready")
			break
		}
	}

	return engine, nil
}

func (e *AnalysisEngine) sendCommand(cmd string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	logger.Debug("ANALYSIS", ">>> %s", cmd)
	_, err := e.stdin.WriteString(cmd + "\n")
	if err != nil {
		return err
	}
	return e.stdin.Flush()
}

// StartAnalysis starts analyzing a position
func (e *AnalysisEngine) StartAnalysis(fen string, analysisChannel chan<- AnalysisInfo) error {
	e.sendCommand("ucinewgame")
	e.sendCommand("position fen " + fen)
	e.sendCommand("go infinite")

	go func() {
		for e.active && e.stdout.Scan() {
			line := e.stdout.Text()

			if strings.HasPrefix(line, "info") {
				info, multiPVNum := parseUCIAnalysisInfo(line)
				if info.BestMove != "" {
					// Store this PV line
					e.multiPVMu.Lock()
					e.multiPVs[multiPVNum] = info

					// If we have all 3 PVs (or at least PV 1), send combined info
					if multiPVNum == 1 || len(e.multiPVs) >= 3 {
						combinedInfo := e.buildCombinedInfo()
						e.multiPVMu.Unlock()

						select {
						case analysisChannel <- combinedInfo:
						default:
							// Channel full, skip this update
						}
					} else {
						e.multiPVMu.Unlock()
					}
				}
			}
		}
	}()

	return nil
}

// StopAnalysis stops the current analysis
func (e *AnalysisEngine) StopAnalysis() {
	e.sendCommand("stop")
	time.Sleep(50 * time.Millisecond) // Give engine time to stop
}

// Close closes the engine
func (e *AnalysisEngine) Close() {
	e.active = false
	e.sendCommand("quit")
	e.cmd.Wait()
}

// buildCombinedInfo combines multiple PV lines into a single AnalysisInfo
// Must be called with multiPVMu locked
func (e *AnalysisEngine) buildCombinedInfo() AnalysisInfo {
	// Use PV 1 as the base (best move)
	baseInfo, ok := e.multiPVs[1]
	if !ok {
		return AnalysisInfo{}
	}

	// Build MultiPV array with all available PV lines
	multiPV := make([]PVLine, 0, 3)
	for i := 1; i <= 3; i++ {
		if pvInfo, exists := e.multiPVs[i]; exists {
			multiPV = append(multiPV, PVLine{
				Score:     pvInfo.Score,
				ScoreType: pvInfo.ScoreType,
				Moves:     pvInfo.PV,
			})
		}
	}

	baseInfo.MultiPV = multiPV
	return baseInfo
}

// parseUCIAnalysisInfo parses a UCI info line and returns the info and multipv number
func parseUCIAnalysisInfo(line string) (AnalysisInfo, int) {
	info := AnalysisInfo{}
	multiPVNum := 1 // Default to 1 if not specified
	parts := strings.Fields(line)

	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "depth":
			if i+1 < len(parts) {
				fmt.Sscanf(parts[i+1], "%d", &info.Depth)
			}
		case "score":
			if i+2 < len(parts) {
				info.ScoreType = parts[i+1]
				switch parts[i+1] {
				case "cp", "mate":
					fmt.Sscanf(parts[i+2], "%d", &info.Score)
				}
				i += 2
			}
		case "nodes":
			if i+1 < len(parts) {
				fmt.Sscanf(parts[i+1], "%d", &info.Nodes)
			}
		case "nps":
			if i+1 < len(parts) {
				fmt.Sscanf(parts[i+1], "%d", &info.NPS)
			}
		case "time":
			if i+1 < len(parts) {
				fmt.Sscanf(parts[i+1], "%d", &info.Time)
			}
		case "pv":
			// Everything after "pv" is the principal variation
			if i+1 < len(parts) {
				info.PV = parts[i+1:]
				if len(info.PV) > 0 {
					info.BestMove = info.PV[0]
				}
			}
			return info, multiPVNum // PV is always last, return immediately
		case "multipv":
			if i+1 < len(parts) {
				fmt.Sscanf(parts[i+1], "%d", &multiPVNum)
			}
		}
	}

	return info, multiPVNum
}
