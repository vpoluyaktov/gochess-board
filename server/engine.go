package server

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// UCIEngine represents a UCI chess engine
type UCIEngine struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader
	mu     sync.Mutex
}

// NewUCIEngine creates and initializes a new UCI engine
func NewUCIEngine(enginePath string) (*UCIEngine, error) {
	cmd := exec.Command(enginePath)
	
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
	
	engine := &UCIEngine{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewReader(stdout),
	}
	
	// Initialize UCI
	if err := engine.sendCommand("uci"); err != nil {
		return nil, err
	}
	
	// Wait for uciok
	if err := engine.waitForResponse("uciok", 5*time.Second); err != nil {
		return nil, err
	}
	
	// Set ready
	if err := engine.sendCommand("isready"); err != nil {
		return nil, err
	}
	
	if err := engine.waitForResponse("readyok", 5*time.Second); err != nil {
		return nil, err
	}
	
	return engine, nil
}

// sendCommand sends a command to the engine
func (e *UCIEngine) sendCommand(cmd string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	_, err := fmt.Fprintf(e.stdin, "%s\n", cmd)
	return err
}

// readLine reads a line from the engine
func (e *UCIEngine) readLine() (string, error) {
	line, err := e.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
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
	if err := e.sendCommand(fmt.Sprintf("position fen %s", fen)); err != nil {
		return "", err
	}
	
	// Start search
	if err := e.sendCommand(fmt.Sprintf("go movetime %d", moveTime.Milliseconds())); err != nil {
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
			return "", err
		}
		
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	
	return "", fmt.Errorf("timeout waiting for bestmove")
}

// GetBestMoveWithClock gets the best move using chess clock time management
// The engine will manage its own time based on remaining time and increment
func (e *UCIEngine) GetBestMoveWithClock(fen string, moveHistory []string, whiteTime, blackTime, whiteInc, blackInc time.Duration) (string, error) {
	// Set position using FEN (most reliable way to ensure correct position)
	posCmd := fmt.Sprintf("position fen %s", fen)
	
	if err := e.sendCommand(posCmd); err != nil {
		return "", err
	}
	
	// Build go command with time controls
	// Format: go wtime <ms> btime <ms> winc <ms> binc <ms>
	goCmd := fmt.Sprintf("go wtime %d btime %d winc %d binc %d",
		whiteTime.Milliseconds(),
		blackTime.Milliseconds(),
		whiteInc.Milliseconds(),
		blackInc.Milliseconds())
	
	if err := e.sendCommand(goCmd); err != nil {
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
			return "", err
		}
		
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	
	return "", fmt.Errorf("timeout waiting for bestmove")
}

// Close closes the engine
func (e *UCIEngine) Close() error {
	e.sendCommand("quit")
	e.stdin.Close()
	e.stdout.Close()
	return e.cmd.Wait()
}
