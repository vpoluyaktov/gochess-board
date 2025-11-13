package server

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// AnalysisInfo represents engine analysis data
type AnalysisInfo struct {
	Depth     int      `json:"depth"`
	Score     int      `json:"score"`    // centipawns
	BestMove  string   `json:"bestMove"` // e.g., "e2e4"
	PV        []string `json:"pv"`       // principal variation
	Nodes     int64    `json:"nodes"`
	NPS       int64    `json:"nps"`       // nodes per second
	Time      int      `json:"time"`      // milliseconds
	ScoreType string   `json:"scoreType"` // "cp" or "mate"
}

// AnalysisEngine manages a UCI engine for analysis
type AnalysisEngine struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Scanner
	mu     sync.Mutex
	active bool
}

// NewAnalysisEngine creates a new analysis engine
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

	// Initialize UCI
	engine.sendCommand("uci")

	// Wait for uciok
	for engine.stdout.Scan() {
		line := engine.stdout.Text()
		if strings.HasPrefix(line, "uciok") {
			break
		}
	}

	engine.sendCommand("isready")
	for engine.stdout.Scan() {
		line := engine.stdout.Text()
		if strings.HasPrefix(line, "readyok") {
			break
		}
	}

	return engine, nil
}

func (e *AnalysisEngine) sendCommand(cmd string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

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
				info := parseAnalysisInfo(line)
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

// parseAnalysisInfo parses a UCI info line
func parseAnalysisInfo(line string) AnalysisInfo {
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

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// handleAnalysisWebSocket handles WebSocket connections for live analysis
func (s *Server) handleAnalysisWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Println("Analysis WebSocket connected")

	var engine *AnalysisEngine
	var sessionID string
	analysisChannel := make(chan AnalysisInfo, 10)

	// Handle incoming messages
	for {
		var msg struct {
			Action     string `json:"action"`
			FEN        string `json:"fen"`
			EnginePath string `json:"enginePath"`
		}

		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		switch msg.Action {
		case "start":
			// Stop previous analysis if any
			if engine != nil {
				engine.StopAnalysis()
				engine.Close()
				// Unregister old engine
				if sessionID != "" {
					GlobalMonitor.UnregisterEngine(sessionID)
				}
			}

			// Start new engine
			enginePath := msg.EnginePath
			if enginePath == "" {
				enginePath = "stockfish" // default
			}

			engine, err = NewAnalysisEngine(enginePath)
			if err != nil {
				log.Printf("Failed to start analysis engine: %v", err)
				conn.WriteJSON(map[string]string{"error": err.Error()})
				continue
			}

			// Register analysis engine in monitor
			sessionID = fmt.Sprintf("analysis-%s", time.Now().Format("20060102-150405.000000"))
			engineName := "Unknown"
			for _, e := range s.engines {
				if e.Path == enginePath {
					engineName = e.Name
					break
				}
			}
			
			activeEngine := &ActiveEngine{
				Name:           engineName + " (Analysis)",
				Path:           enginePath,
				ELO:            0, // Analysis engines run at full strength
				WhiteTime:      0,
				BlackTime:      0,
				WhiteIncrement: 0,
				BlackIncrement: 0,
				StartTime:      time.Now(),
				SessionID:      sessionID,
			}
			GlobalMonitor.RegisterEngine(sessionID, activeEngine)

			// Start analysis
			err = engine.StartAnalysis(msg.FEN, analysisChannel)
			if err != nil {
				log.Printf("Failed to start analysis: %v", err)
				conn.WriteJSON(map[string]string{"error": err.Error()})
				GlobalMonitor.UnregisterEngine(sessionID)
				continue
			}

			// Send analysis updates
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond) // Throttle to 10 updates/sec
				defer ticker.Stop()

				var lastInfo AnalysisInfo

				for {
					select {
					case info := <-analysisChannel:
						lastInfo = info
					case <-ticker.C:
						if lastInfo.BestMove != "" {
							err := conn.WriteJSON(lastInfo)
							if err != nil {
								return
							}
						}
					}
				}
			}()

		case "stop":
			if engine != nil {
				engine.StopAnalysis()
				engine.Close()
				engine = nil
			}
			// Unregister analysis engine
			if sessionID != "" {
				GlobalMonitor.UnregisterEngine(sessionID)
				sessionID = ""
			}

		case "update":
			// Update position for analysis
			if engine != nil {
				engine.StopAnalysis()
				time.Sleep(50 * time.Millisecond)
				engine.StartAnalysis(msg.FEN, analysisChannel)
			}
		}
	}

	// Cleanup
	if engine != nil {
		engine.StopAnalysis()
		engine.Close()
	}
	
	// Unregister analysis engine
	if sessionID != "" {
		GlobalMonitor.UnregisterEngine(sessionID)
	}

	log.Println("Analysis WebSocket disconnected")
}
