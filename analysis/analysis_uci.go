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
	cmd             *exec.Cmd
	stdin           *bufio.Writer
	stdout          *bufio.Scanner
	mu              sync.Mutex
	active          bool
	multiPVs        map[int]AnalysisInfo // Store multiple PV lines indexed by multipv number
	multiPVMu       sync.Mutex           // Mutex for multiPVs map
	blackToMove     bool                 // Track whose turn it is for proper MultiPV sorting
	currentFEN      string               // Current position being analyzed
	analysisChannel chan<- AnalysisInfo  // Current channel to send analysis results
	channelMu       sync.Mutex           // Mutex for analysisChannel
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

	logger.Debug("ANALYSIS", "Initializing UCI analysis engine: %s", enginePath)

	// Initialize UCI
	engine.sendCommand("uci")

	// Wait for uciok
	logger.Debug("ANALYSIS", "Waiting for uciok...")
	for engine.stdout.Scan() {
		line := engine.stdout.Text()
		logger.Debug("ANALYSIS", "<<< %s", line)
		if strings.HasPrefix(line, "uciok") {
			logger.Debug("ANALYSIS", "Got uciok")
			break
		}
	}

	// Set MultiPV option to 3 for getting 3 best moves
	engine.sendCommand("setoption name MultiPV value 3")

	engine.sendCommand("isready")
	logger.Debug("ANALYSIS", "Waiting for readyok...")
	for engine.stdout.Scan() {
		line := engine.stdout.Text()
		logger.Debug("ANALYSIS", "<<< %s", line)
		if strings.HasPrefix(line, "readyok") {
			logger.Debug("ANALYSIS", "UCI analysis engine ready")
			break
		}
	}

	// Start a single reader goroutine that will run for the lifetime of the engine
	go engine.readLoop()

	return engine, nil
}

// readLoop continuously reads from the engine stdout and processes analysis info
func (e *AnalysisEngine) readLoop() {
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

					// Send to current analysis channel (if set)
					e.channelMu.Lock()
					ch := e.analysisChannel
					e.channelMu.Unlock()

					if ch != nil {
						select {
						case ch <- combinedInfo:
						default:
							// Channel full, skip this update
						}
					}
				} else {
					e.multiPVMu.Unlock()
				}
			}
		}
	}
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
	// Update the channel reference (the readLoop goroutine will use this)
	e.channelMu.Lock()
	e.analysisChannel = analysisChannel
	e.channelMu.Unlock()

	// Store the FEN being analyzed
	e.currentFEN = fen

	// Parse FEN to determine whose turn it is
	parts := strings.Fields(fen)
	if len(parts) >= 2 {
		e.blackToMove = parts[1] == "b"
	} else {
		e.blackToMove = false
	}

	logger.Debug("ANALYSIS", "UCI StartAnalysis: FEN=%s, blackToMove=%v", fen, e.blackToMove)

	// Clear old MultiPV data from previous position
	e.multiPVMu.Lock()
	e.multiPVs = make(map[int]AnalysisInfo)
	e.multiPVMu.Unlock()

	e.sendCommand("ucinewgame")
	e.sendCommand("position fen " + fen)
	e.sendCommand("go infinite")

	return nil
}

// StopAnalysis stops the current analysis
func (e *AnalysisEngine) StopAnalysis() {
	e.sendCommand("stop")
	time.Sleep(50 * time.Millisecond) // Give engine time to stop
}

// Close closes the engine with a timeout to prevent hanging
func (e *AnalysisEngine) Close() {
	e.active = false
	e.sendCommand("quit")

	// Wait for engine to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- e.cmd.Wait()
	}()

	select {
	case <-done:
		return
	case <-time.After(2 * time.Second):
		// Engine didn't exit gracefully, force kill
		logger.Warn("ANALYSIS", "Engine did not exit gracefully, killing process")
		if e.cmd.Process != nil {
			e.cmd.Process.Kill()
		}
	}
}

// buildCombinedInfo combines multiple PV lines into a single AnalysisInfo
// Must be called with multiPVMu locked
func (e *AnalysisEngine) buildCombinedInfo() AnalysisInfo {
	// Use PV 1 as the base (best move)
	baseInfo, ok := e.multiPVs[1]
	if !ok {
		return AnalysisInfo{}
	}

	// UCI engines report scores from the side-to-move's perspective.
	// A positive score means the position is good for the side to move.
	// A negative score means the position is bad for the side to move.
	//
	// To normalize to White's perspective (positive = White winning):
	// - When White is to move: keep the score as-is (positive = good for White)
	// - When Black is to move: negate the score (positive from Black's view = negative from White's view)
	//
	// Example: After 1.e4, Black to move, Stockfish says +30 (good for Black)
	//          We negate to -30 (bad for White = good for Black) ✓
	//
	// Example: After 1.e4, Black to move, Stockfish says -10 (bad for Black)
	//          We negate to +10 (good for White) ✓
	if e.blackToMove {
		logger.Debug("ANALYSIS", "UCI: Black to move, negating score %d -> %d", baseInfo.Score, -baseInfo.Score)
		baseInfo.Score = -baseInfo.Score
	}

	// Build MultiPV array with all available PV lines
	multiPV := make([]PVLine, 0, 3)
	for i := 1; i <= 3; i++ {
		if pvInfo, exists := e.multiPVs[i]; exists {
			score := pvInfo.Score
			// Normalize score to White's perspective
			if e.blackToMove {
				score = -score
			}
			logger.Debug("ANALYSIS", "UCI MultiPV[%d]: score=%d, scoreType=%s, moves=%v", i, score, pvInfo.ScoreType, pvInfo.PV)
			multiPV = append(multiPV, PVLine{
				Score:     score,
				ScoreType: pvInfo.ScoreType,
				Moves:     pvInfo.PV,
			})
		}
	}

	// MultiPV is now normalized to White's perspective (positive = good for White).
	// The array is already sorted by the engine with best move first.
	// No need to reverse for Black's turn since we've negated the scores.

	baseInfo.MultiPV = multiPV
	baseInfo.FEN = e.currentFEN // Include FEN for frontend sync verification
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
