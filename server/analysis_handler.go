package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"gochess-board/analysis"
	"gochess-board/engines"
	"gochess-board/logger"
)

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
		logger.Error("ANALYSIS", "WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	logger.Info("ANALYSIS", "WebSocket connected")

	var analysisEngine analysis.AnalysisEngineInterface
	var sessionID string
	analysisChannel := make(chan analysis.AnalysisInfo, 10)
	stopSending := make(chan bool, 1) // Channel to stop the sending goroutine

	// Handle incoming messages
	for {
		var msg struct {
			Action     string `json:"action"`
			FEN        string `json:"fen"`
			EnginePath string `json:"enginePath"`
		}

		err := conn.ReadJSON(&msg)
		if err != nil {
			logger.Debug("ANALYSIS", "WebSocket closed: %v", err)
			break
		}

		switch msg.Action {
		case "start":
			// Stop previous analysis if any
			if analysisEngine != nil {
				analysisEngine.StopAnalysis()
				analysisEngine.Close()
				// Unregister old engine
				if sessionID != "" {
					engines.GlobalMonitor.UnregisterEngine(sessionID)
				}
			}

			// Start new engine
			enginePath := msg.EnginePath
			if enginePath == "" {
				enginePath = "stockfish" // default
			}

			// Detect engine type
			engineName := "Unknown"
			engineType := "uci" // default
			for _, e := range s.engines {
				if e.Path == enginePath {
					engineName = e.Name
					engineType = e.Type
					break
				}
			}

			// Create appropriate analysis engine based on type
			if engineType == "cecp" {
				analysisEngine, err = analysis.NewCECPAnalysisEngine(enginePath)
			} else {
				analysisEngine, err = analysis.NewAnalysisEngine(enginePath)
			}
			if err != nil {
				logger.Error("ANALYSIS", "Failed to start analysis engine: %v", err)
				conn.WriteJSON(map[string]string{"error": err.Error()})
				continue
			}

			// Register analysis engine in monitor
			sessionID = fmt.Sprintf("analysis-%s", time.Now().Format("20060102-150405.000000"))

			activeEngine := &engines.ActiveEngine{
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
			engines.GlobalMonitor.RegisterEngine(sessionID, activeEngine)

			// Start analysis
			err = analysisEngine.StartAnalysis(msg.FEN, analysisChannel)
			if err != nil {
				logger.Error("ANALYSIS", "Failed to start analysis: %v", err)
				conn.WriteJSON(map[string]string{"error": err.Error()})
				engines.GlobalMonitor.UnregisterEngine(sessionID)
				continue
			}

			// Send analysis updates
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond) // Throttle to 10 updates/sec
				defer ticker.Stop()

				var lastInfo analysis.AnalysisInfo

				for {
					select {
					case info := <-analysisChannel:
						lastInfo = info
						logger.Debug("ANALYSIS", "Received from channel: depth=%d, move=%s", info.Depth, info.BestMove)
					case <-ticker.C:
						if lastInfo.BestMove != "" {
							logger.Debug("ANALYSIS", "Sending to WebSocket: depth=%d, move=%s", lastInfo.Depth, lastInfo.BestMove)
							err := conn.WriteJSON(lastInfo)
							if err != nil {
								logger.Error("ANALYSIS", "WebSocket write error: %v", err)
								return
							}
						}
					case <-stopSending:
						logger.Info("ANALYSIS", "Stopping analysis updates goroutine")
						return
					}
				}
			}()

		case "stop":
			// Signal the sending goroutine to stop
			select {
			case stopSending <- true:
			default:
			}

			if analysisEngine != nil {
				analysisEngine.StopAnalysis()
				analysisEngine.Close()
				analysisEngine = nil
			}
			// Unregister analysis engine
			if sessionID != "" {
				engines.GlobalMonitor.UnregisterEngine(sessionID)
				sessionID = ""
			}

			// Create a new stop channel for the next analysis session
			stopSending = make(chan bool, 1)

		case "update":
			// Update position for analysis
			if analysisEngine != nil {
				analysisEngine.StopAnalysis()
				time.Sleep(50 * time.Millisecond)
				analysisEngine.StartAnalysis(msg.FEN, analysisChannel)
			}
		}
	}

	// Cleanup
	if analysisEngine != nil {
		analysisEngine.StopAnalysis()
		analysisEngine.Close()
	}

	// Unregister analysis engine
	if sessionID != "" {
		engines.GlobalMonitor.UnregisterEngine(sessionID)
	}

	logger.Info("ANALYSIS", "WebSocket disconnected")
}
