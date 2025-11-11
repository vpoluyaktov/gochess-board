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

// GetBestMove gets the best move for the current position
func (e *UCIEngine) GetBestMove(fen string, moveTime time.Duration) (string, error) {
	// Set position
	if err := e.sendCommand(fmt.Sprintf("position fen %s", fen)); err != nil {
		return "", err
	}
	
	// Start search
	if err := e.sendCommand(fmt.Sprintf("go movetime %d", moveTime.Milliseconds())); err != nil {
		return "", err
	}
	
	// Wait for bestmove
	deadline := time.Now().Add(moveTime + 2*time.Second)
	
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
