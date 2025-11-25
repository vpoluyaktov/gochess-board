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

// CECPEngine represents a CECP/XBoard chess engine
type CECPEngine struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader
	mu     sync.Mutex
	name   string // Engine name for logging
}

// NewCECPEngine creates and initializes a new CECP engine
func NewCECPEngine(enginePath string, engineName string) (*CECPEngine, error) {
	// Some CECP engines need special flags to enter xboard mode
	// GNU Chess needs --xboard flag
	var cmd *exec.Cmd
	if strings.Contains(enginePath, "gnuchess") {
		cmd = exec.Command(enginePath, "--xboard")
	} else {
		cmd = exec.Command(enginePath)
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
		engineName = enginePath
	}

	engine := &CECPEngine{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewReader(stdout),
		name:   engineName,
	}

	Info("ENGINE", "[%s] Initializing CECP engine at path: %s", engineName, enginePath)

	// Initialize CECP protocol
	if err := engine.sendCommand("xboard"); err != nil {
		Error("ENGINE", "[%s] Failed to send 'xboard' command: %v", engineName, err)
		return nil, err
	}

	// Send protover 2 to enable protocol version 2 features
	if err := engine.sendCommand("protover 2"); err != nil {
		Error("ENGINE", "[%s] Failed to send 'protover 2' command: %v", engineName, err)
		return nil, err
	}

	// Wait for feature done=1
	Info("ENGINE", "[%s] Waiting for 'feature done=1' response...", engineName)
	if err := engine.waitForFeatureDone(10 * time.Second); err != nil {
		Error("ENGINE", "[%s] Failed to get 'feature done=1': %v", engineName, err)
		return nil, err
	}

	// Disable pondering (thinking on opponent's time)
	if err := engine.sendCommand("hard"); err != nil {
		Warn("ENGINE", "[%s] Failed to disable pondering: %v", engineName, err)
	}

	// Set post mode to get thinking output
	if err := engine.sendCommand("post"); err != nil {
		Warn("ENGINE", "[%s] Failed to enable post mode: %v", engineName, err)
	}

	Info("ENGINE", "[%s] Successfully initialized", engineName)

	return engine, nil
}

// sendCommand sends a command to the engine
func (e *CECPEngine) sendCommand(cmd string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	Debug("ENGINE", "[%s] >>> %s", e.name, cmd)
	_, err := fmt.Fprintf(e.stdin, "%s\n", cmd)
	return err
}

// readLine reads a line from the engine
func (e *CECPEngine) readLine() (string, error) {
	line, err := e.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	Debug("ENGINE", "[%s] <<< %s", e.name, line)
	return line, nil
}

// waitForFeatureDone waits for the "feature done=1" response
func (e *CECPEngine) waitForFeatureDone(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		line, err := e.readLine()
		if err != nil {
			return err
		}

		if strings.Contains(line, "feature") && strings.Contains(line, "done=1") {
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for: feature done=1")
}

// SetOption sets an option on the engine (CECP uses "option" command)
func (e *CECPEngine) SetOption(name, value string) error {
	cmd := fmt.Sprintf("option %s=%s", name, value)
	return e.sendCommand(cmd)
}

// GetBestMove gets the best move for the current position using fixed time
func (e *CECPEngine) GetBestMove(fen string, moveTime time.Duration) (string, error) {
	// Set position using setboard command
	if err := e.sendCommand(fmt.Sprintf("setboard %s", fen)); err != nil {
		return "", err
	}

	// Set time control - CECP uses centiseconds
	timeCs := int(moveTime.Milliseconds() / 10)
	if err := e.sendCommand(fmt.Sprintf("st %d", timeCs/100)); err != nil {
		return "", err
	}

	// Start search
	if err := e.sendCommand("go"); err != nil {
		return "", err
	}

	// Wait for move response
	deadline := time.Now().Add(moveTime + 5*time.Second)
	Info("ENGINE", "[%s] Waiting for move response...", e.name)

	for time.Now().Before(deadline) {
		line, err := e.readLine()
		if err != nil {
			Error("ENGINE", "[%s] Error reading line: %v", e.name, err)
			return "", err
		}

		// CECP engines respond with "move <move>" or "My move is : <move>"
		if strings.HasPrefix(line, "move ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				move := parts[1]
				Info("ENGINE", "[%s] Got move: %s", e.name, move)
				return move, nil
			}
		} else if strings.Contains(line, "My move is") {
			// GNU Chess format: "My move is : e2e4"
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				move := strings.TrimSpace(parts[1])
				Info("ENGINE", "[%s] Got move: %s", e.name, move)
				return move, nil
			}
		}
	}

	Error("ENGINE", "[%s] Timeout waiting for move", e.name)
	return "", fmt.Errorf("timeout waiting for move")
}

// GetBestMoveWithClock gets the best move using chess clock time management
func (e *CECPEngine) GetBestMoveWithClock(fen string, moveHistory []string, whiteTime, blackTime, whiteInc, blackInc time.Duration) (string, error) {
	// Set position using setboard command
	if err := e.sendCommand(fmt.Sprintf("setboard %s", fen)); err != nil {
		return "", err
	}

	// Set time controls - CECP uses centiseconds
	whiteCs := int(whiteTime.Milliseconds() / 10)
	blackCs := int(blackTime.Milliseconds() / 10)
	// Note: CECP doesn't have a direct increment command like UCI
	// Increments are typically handled by the GUI updating time after each move

	// Send time command
	if err := e.sendCommand(fmt.Sprintf("time %d", whiteCs)); err != nil {
		return "", err
	}
	if err := e.sendCommand(fmt.Sprintf("otim %d", blackCs)); err != nil {
		return "", err
	}

	// Start search
	if err := e.sendCommand("go"); err != nil {
		return "", err
	}

	// Wait for move response
	maxTime := whiteTime
	if blackTime > maxTime {
		maxTime = blackTime
	}
	deadline := time.Now().Add(maxTime + 5*time.Second)
	Info("ENGINE", "[%s] Waiting for move response (timeout: %v)...", e.name, maxTime+5*time.Second)

	for time.Now().Before(deadline) {
		line, err := e.readLine()
		if err != nil {
			Error("ENGINE", "[%s] Error reading line: %v", e.name, err)
			return "", err
		}

		// CECP engines respond with "move <move>" or "My move is : <move>"
		if strings.HasPrefix(line, "move ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				move := parts[1]
				Info("ENGINE", "[%s] Got move: %s", e.name, move)
				return move, nil
			}
		} else if strings.Contains(line, "My move is") {
			// GNU Chess format: "My move is : e2e4"
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				move := strings.TrimSpace(parts[1])
				Info("ENGINE", "[%s] Got move: %s", e.name, move)
				return move, nil
			}
		}
	}

	Error("ENGINE", "[%s] Timeout waiting for move", e.name)
	return "", fmt.Errorf("timeout waiting for move")
}

// Close closes the engine
func (e *CECPEngine) Close() error {
	e.sendCommand("quit")
	e.stdin.Close()
	e.stdout.Close()
	return e.cmd.Wait()
}
