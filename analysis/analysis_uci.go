package analysis

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go-chess/logger"
)

// AnalysisEngine manages a UCI engine for analysis
type AnalysisEngine struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Scanner
	mu     sync.Mutex
	active bool
}

// NewAnalysisEngine creates a new UCI analysis engine
func NewAnalysisEngine(enginePath string) (*AnalysisEngine, error) {
	cmd := exec.Command(enginePath)

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

	engine := &AnalysisEngine{
		cmd:    cmd,
		stdin:  bufio.NewWriter(stdin),
		stdout: bufio.NewScanner(stdout),
		active: true,
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
				info := parseUCIAnalysisInfo(line)
				if info.BestMove != "" {
					select {
					case analysisChannel <- info:
					default:
						// Channel full, skip this update
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

// parseUCIAnalysisInfo parses a UCI info line
func parseUCIAnalysisInfo(line string) AnalysisInfo {
	info := AnalysisInfo{}
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
			return info // PV is always last, return immediately
		}
	}

	return info
}
