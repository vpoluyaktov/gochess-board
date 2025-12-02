package analysis

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/notnil/chess"

	"gochess-board/logger"
)

// CECPAnalysisEngine manages a CECP engine for analysis
type CECPAnalysisEngine struct {
	cmd      *exec.Cmd
	stdin    *bufio.Writer
	stdout   *bufio.Scanner
	mu       sync.Mutex
	active   bool
	position *chess.Position // Track current position for SAN to UCI conversion
}

// NewCECPAnalysisEngine creates a new CECP analysis engine
func NewCECPAnalysisEngine(enginePath string) (*CECPAnalysisEngine, error) {
	// Some CECP engines need special flags to enter xboard mode
	var cmd *exec.Cmd
	if strings.Contains(enginePath, "gnuchess") {
		cmd = exec.Command(enginePath, "--xboard")
	} else {
		cmd = exec.Command(enginePath)
	}

	// Set working directory to temp to avoid cluttering the repository with log files
	// os.TempDir() returns the appropriate temp directory for the OS:
	// - Linux/macOS: /tmp
	// - Windows: %TEMP% (e.g., C:\\Users\\<user>\\AppData\\Local\\Temp)
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

	engine := &CECPAnalysisEngine{
		cmd:    cmd,
		stdin:  bufio.NewWriter(stdin),
		stdout: bufio.NewScanner(stdout),
		active: true,
	}

	logger.Info("ANALYSIS", "Initializing CECP analysis engine: %s", enginePath)

	// Initialize CECP protocol
	engine.sendCommand("xboard")
	engine.sendCommand("protover 2")

	// Wait for feature done=1
	logger.Info("ANALYSIS", "Waiting for feature done=1...")
	for engine.stdout.Scan() {
		line := engine.stdout.Text()
		logger.Debug("ANALYSIS", "<<< %s", line)
		if strings.Contains(line, "feature") && strings.Contains(line, "done=1") {
			logger.Info("ANALYSIS", "Got feature done=1")
			break
		}
	}

	// Enable post mode to get thinking output
	engine.sendCommand("post")

	logger.Info("ANALYSIS", "CECP analysis engine ready")

	return engine, nil
}

func (e *CECPAnalysisEngine) sendCommand(cmd string) error {
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
func (e *CECPAnalysisEngine) StartAnalysis(fen string, analysisChannel chan<- AnalysisInfo) error {
	// Parse and store the position for SAN to UCI conversion
	fenObj, err := chess.FEN(fen)
	if err != nil {
		return fmt.Errorf("invalid FEN: %w", err)
	}
	game := chess.NewGame(fenObj)
	e.position = game.Position()

	e.sendCommand("setboard " + fen)
	e.sendCommand("analyze")

	// Read analysis output in a goroutine
	go func() {
		logger.Info("ANALYSIS", "CECP analysis goroutine started")
		for e.stdout.Scan() {
			if !e.active {
				logger.Info("ANALYSIS", "CECP analysis stopped (active=false)")
				break
			}

			line := e.stdout.Text()
			logger.Debug("ANALYSIS", "<<< %s", line)

			// Parse CECP analysis output
			// Format: depth score time nodes PV
			info := parseCECPAnalysis(line, e.position)
			if info.BestMove != "" {
				logger.Debug("ANALYSIS", "Sending analysis: depth=%d, score=%d, move=%s", info.Depth, info.Score, info.BestMove)
				analysisChannel <- info
			}

			// Check for analyze complete
			if strings.Contains(line, "analyze complete") {
				logger.Info("ANALYSIS", "CECP analysis complete")
				break
			}
		}
		logger.Info("ANALYSIS", "CECP analysis goroutine exited")
	}()

	return nil
}

// StopAnalysis stops the current analysis
func (e *CECPAnalysisEngine) StopAnalysis() {
	e.active = false
	e.sendCommand("exit") // Exit analyze mode
}

// Close closes the engine
func (e *CECPAnalysisEngine) Close() {
	e.StopAnalysis()
	e.sendCommand("quit")
	e.cmd.Wait()
}

// parseCECPAnalysis parses CECP analysis output and converts SAN moves to UCI
// Format: depth score time nodes PV
// Example: "13 +22 245 6028471 Nf3 Nc6 Nc3 Nf6 e3 d5"
func parseCECPAnalysis(line string, position *chess.Position) AnalysisInfo {
	info := AnalysisInfo{}

	// Skip non-analysis lines (feature lines, invalid move, etc.)
	if strings.Contains(line, "feature") || strings.Contains(line, "Invalid") || strings.Contains(line, "Analyze") {
		return info
	}

	fields := strings.Fields(line)
	if len(fields) < 5 {
		return info
	}

	// Parse depth
	if depth, err := strconv.Atoi(fields[0]); err == nil {
		info.Depth = depth
	} else {
		// Not a valid analysis line
		return info
	}

	// Parse score (in centipawns, may have +/- prefix)
	scoreStr := fields[1]
	if score, err := strconv.Atoi(scoreStr); err == nil {
		info.Score = score
		info.ScoreType = "cp"
	} else {
		return info
	}

	// Parse time (in centiseconds, convert to milliseconds)
	if timeCs, err := strconv.Atoi(fields[2]); err == nil {
		info.Time = timeCs * 10
	}

	// Parse nodes
	if nodes, err := strconv.ParseInt(fields[3], 10, 64); err == nil {
		info.Nodes = nodes
		if info.Time > 0 {
			info.NPS = nodes * 1000 / int64(info.Time)
		}
	}

	// Parse PV (principal variation) and convert SAN to UCI
	// GNU Chess format: just moves without numbers: "Nc3 Nc6 Nf3 Nf6"
	pvStart := 4
	pv := []string{}

	// Create a temp position starting from the current position
	tempPos := position

	for i := pvStart; i < len(fields); i++ {
		moveStr := fields[i]
		// Skip move numbers if present (e.g., "1.", "2.")
		if strings.HasSuffix(moveStr, ".") {
			continue
		}

		// Try to parse the SAN move
		move, err := chess.AlgebraicNotation{}.Decode(tempPos, moveStr)
		if err != nil {
			// If we can't parse this move, stop processing PV
			break
		}

		// Convert to UCI notation
		uciMove := chess.UCINotation{}.Encode(tempPos, move)
		pv = append(pv, uciMove)

		// Apply the move to get next position
		tempPos = tempPos.Update(move)
	}

	if len(pv) > 0 {
		info.PV = pv
		info.BestMove = pv[0]

		// CECP engines don't support multi-PV natively
		// Create a single-entry MultiPV for consistency with UCI engines
		info.MultiPV = []PVLine{
			{
				Score:     info.Score,
				ScoreType: info.ScoreType,
				Moves:     pv,
			},
		}
	}

	return info
}
