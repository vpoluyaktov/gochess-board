package engines

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gochess-board/logger"
)

// UCIEngine represents a UCI chess engine
type UCIEngine struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader
	mu     sync.Mutex
	name   string // Engine name for logging
}

// NewUCIEngine creates and initializes a new UCI engine
func NewUCIEngine(enginePath string, engineName string) (*UCIEngine, error) {
	cmd := exec.Command(enginePath)

	// Set working directory to temp to avoid cluttering the repository with log files
	// os.TempDir() returns the appropriate temp directory for the OS:
	// - Linux/macOS: /tmp
	// - Windows: %TEMP% (e.g., C:\Users\<user>\AppData\Local\Temp)
	tempDir := filepath.Join(os.TempDir(), "gochess-board-engines")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.Warn("ENGINE", "Failed to create temp directory for engine: %v", err)
	} else {
		cmd.Dir = tempDir
		logger.Debug("ENGINE", "Engine working directory set to: %s", tempDir)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start engine: %w", err)
	}

	// Use provided name, or fall back to filename if empty
	if engineName == "" {
		engineName = filepath.Base(enginePath)
	}

	engine := &UCIEngine{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewReader(stdout),
		name:   engineName,
	}

	logger.Debug("ENGINE", "[%s] Initializing UCI engine at path: %s", engineName, enginePath)

	// Initialize UCI
	if err := engine.sendCommand("uci"); err != nil {
		logger.Error("ENGINE", "[%s] Failed to send 'uci' command: %v", engineName, err)
		return nil, err
	}

	// Wait for uciok
	logger.Debug("ENGINE", "[%s] Waiting for 'uciok' response...", engineName)
	if err := engine.waitForResponse("uciok", 5*time.Second); err != nil {
		logger.Error("ENGINE", "[%s] Failed to get 'uciok': %v", engineName, err)
		return nil, err
	}

	// Set ready
	if err := engine.sendCommand("isready"); err != nil {
		logger.Error("ENGINE", "[%s] Failed to send 'isready' command: %v", engineName, err)
		return nil, err
	}

	logger.Debug("ENGINE", "[%s] Waiting for 'readyok' response...", engineName)
	if err := engine.waitForResponse("readyok", 5*time.Second); err != nil {
		logger.Error("ENGINE", "[%s] Failed to get 'readyok': %v", engineName, err)
		return nil, err
	}

	logger.Debug("ENGINE", "[%s] Successfully initialized", engineName)

	return engine, nil
}

// sendCommand sends a command to the engine
func (e *UCIEngine) sendCommand(cmd string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	logger.Debug("ENGINE", "[%s] >>> %s", e.name, cmd)
	_, err := fmt.Fprintf(e.stdin, "%s\n", cmd)
	return err
}

// sendCommandInfo sends a command to the engine and logs it at INFO level
func (e *UCIEngine) sendCommandInfo(cmd string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	logger.Info("ENGINE", "[%s] >>> %s", e.name, cmd)
	_, err := fmt.Fprintf(e.stdin, "%s\n", cmd)
	return err
}

// readLine reads a line from the engine
func (e *UCIEngine) readLine() (string, error) {
	line, err := e.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	logger.Debug("ENGINE", "[%s] <<< %s", e.name, line)
	return line, nil
}

// waitForResponse waits for a specific response from the engine
func (e *UCIEngine) waitForResponse(expected string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		line, err := e.readLine()
		if err != nil {
			return err
		}

		if strings.HasPrefix(line, expected) {
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for: %s", expected)
}

// SetOption sets a UCI option on the engine
func (e *UCIEngine) SetOption(name, value string) error {
	cmd := fmt.Sprintf("setoption name %s value %s", name, value)
	if err := e.sendCommand(cmd); err != nil {
		return err
	}

	// Wait for engine to be ready after setting option
	if err := e.sendCommand("isready"); err != nil {
		return err
	}

	return e.waitForResponse("readyok", 2*time.Second)
}

// GetBestMove gets the best move for the current position using fixed time
func (e *UCIEngine) GetBestMove(fen string, moveTime time.Duration) (string, error) {
	// Set position
	if err := e.sendCommandInfo(fmt.Sprintf("position fen %s", fen)); err != nil {
		return "", err
	}

	// Start search
	if err := e.sendCommandInfo(fmt.Sprintf("go movetime %d", moveTime.Milliseconds())); err != nil {
		return "", err
	}

	// Wait for the movetime to elapse, then send stop
	time.Sleep(moveTime)
	if err := e.sendCommand("stop"); err != nil {
		return "", err
	}

	// Wait for bestmove response
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		line, err := e.readLine()
		if err != nil {
			logger.Error("ENGINE", "[%s] Error reading line: %v", e.name, err)
			return "", err
		}

		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				logger.Info("ENGINE", "[%s] <<< bestmove %s", e.name, parts[1])
				return parts[1], nil
			}
		}
	}

	logger.Error("ENGINE", "[%s] Timeout waiting for bestmove", e.name)
	return "", fmt.Errorf("timeout waiting for bestmove")
}

// GetBestMoveWithClock gets the best move using chess clock time management
// The engine will manage its own time based on remaining time and increment
func (e *UCIEngine) GetBestMoveWithClock(fen string, moveHistory []string, whiteTime, blackTime, whiteInc, blackInc time.Duration) (string, error) {
	// Set position using FEN (most reliable way to ensure correct position)
	posCmd := fmt.Sprintf("position fen %s", fen)

	if err := e.sendCommandInfo(posCmd); err != nil {
		return "", err
	}

	// Build go command with time controls
	// Format: go wtime <ms> btime <ms> winc <ms> binc <ms>
	goCmd := fmt.Sprintf("go wtime %d btime %d winc %d binc %d",
		whiteTime.Milliseconds(),
		blackTime.Milliseconds(),
		whiteInc.Milliseconds(),
		blackInc.Milliseconds())

	if err := e.sendCommandInfo(goCmd); err != nil {
		return "", err
	}

	// Wait for bestmove response (engine manages its own time)
	// Give it a generous timeout (max of remaining time + 5 seconds)
	maxTime := whiteTime
	if blackTime > maxTime {
		maxTime = blackTime
	}
	deadline := time.Now().Add(maxTime + 5*time.Second)

	for time.Now().Before(deadline) {
		line, err := e.readLine()
		if err != nil {
			logger.Error("ENGINE", "[%s] Error reading line: %v", e.name, err)
			return "", err
		}

		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				logger.Info("ENGINE", "[%s] <<< bestmove %s", e.name, parts[1])
				return parts[1], nil
			}
		}
	}

	logger.Error("ENGINE", "[%s] Timeout waiting for bestmove", e.name)
	return "", fmt.Errorf("timeout waiting for bestmove")
}

// Close closes the engine with a timeout to prevent hanging
func (e *UCIEngine) Close() error {
	e.sendCommand("quit")
	e.stdin.Close()
	e.stdout.Close()

	// Wait for engine to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- e.cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		// Engine didn't exit gracefully, force kill
		logger.Warn("ENGINE", "[%s] Engine did not exit gracefully, killing process", e.name)
		if e.cmd.Process != nil {
			e.cmd.Process.Kill()
		}
		return fmt.Errorf("engine did not exit gracefully, killed")
	}
}
