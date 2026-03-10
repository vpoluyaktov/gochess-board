package analysis

import (
	"gochess-board/internal/engines/builtin"
	"gochess-board/internal/logger"
	"strings"
)

// BuiltinAnalysisEngine manages the built-in engine for analysis
type BuiltinAnalysisEngine struct {
	engine        *builtin.InternalEngine
	currentStopCh chan bool // Stop channel for current analysis
	blackToMove   bool      // Track whose turn it is for proper MultiPV sorting
	currentFEN    string    // Current position being analyzed (for sync verification)
}

// NewBuiltinAnalysisEngine creates a new built-in analysis engine
func NewBuiltinAnalysisEngine() (*BuiltinAnalysisEngine, error) {
	logger.Debug("ANALYSIS", "Initializing built-in analysis engine")

	return &BuiltinAnalysisEngine{
		engine: builtin.NewEngine(),
	}, nil
}

// StartAnalysis starts analyzing a position
func (e *BuiltinAnalysisEngine) StartAnalysis(fen string, analysisChannel chan<- AnalysisInfo) error {
	logger.Debug("ANALYSIS", "Starting built-in engine analysis for position: %s", fen)

	// Store the FEN being analyzed
	e.currentFEN = fen

	// Parse FEN to determine whose turn it is
	parts := strings.Fields(fen)
	if len(parts) >= 2 {
		e.blackToMove = parts[1] == "b"
	} else {
		e.blackToMove = false
	}

	// Stop any previous analysis
	if e.currentStopCh != nil {
		select {
		case e.currentStopCh <- true:
		default:
		}
	}

	// Create new stop channel for this analysis session
	e.currentStopCh = make(chan bool, 1)
	stopCh := e.currentStopCh

	// Create a channel to receive builtin.AnalysisInfo
	builtinCh := make(chan builtin.AnalysisInfo, 10)

	// Start analysis in a goroutine
	go func() {
		// Analyze up to depth 20 (mimics "go infinite" behavior)
		const maxDepth = 20
		err := e.engine.Analyze(fen, maxDepth, stopCh, builtinCh)
		if err != nil {
			logger.Error("ANALYSIS", "Built-in engine analysis error: %v", err)
		}
		close(builtinCh) // Close channel when analysis completes
	}()

	// Capture values for the goroutine
	blackToMove := e.blackToMove
	currentFEN := e.currentFEN

	// Forward analysis info from builtin channel to analysis channel
	go func() {
		for {
			select {
			case info, ok := <-builtinCh:
				if !ok {
					// Channel closed, stop forwarding
					return
				}

				// Convert builtin.AnalysisInfo to analysis.AnalysisInfo
				// Need to convert builtin.PVLine to analysis.PVLine
				// Also normalize scores to White's perspective:
				// - Positive score = White is winning
				// - Negative score = Black is winning
				multiPV := make([]PVLine, len(info.MultiPV))
				for i, pv := range info.MultiPV {
					score := pv.Score
					// Normalize score to White's perspective
					if blackToMove {
						score = -score
					}
					multiPV[i] = PVLine{
						Score:     score,
						ScoreType: pv.ScoreType,
						Moves:     pv.Moves,
					}
				}

				// Normalize main score to White's perspective
				mainScore := info.Score
				if blackToMove {
					mainScore = -mainScore
				}

				// MultiPV is now normalized to White's perspective.
				// The array is already sorted by the engine with best move first.

				analysisInfo := AnalysisInfo{
					Depth:     info.Depth,
					Score:     mainScore,
					BestMove:  info.BestMove,
					PV:        info.PV, // Real PV from engine
					Nodes:     int64(info.Nodes),
					NPS:       info.NPS,
					Time:      info.Time,
					ScoreType: info.ScoreType,
					MultiPV:   multiPV,
					FEN:       currentFEN, // Include FEN for frontend sync verification
				}

				logger.Debug("ANALYSIS", "Built-in engine: depth=%d, score=%d (%s), move=%s, nodes=%d, nps=%d",
					info.Depth, info.Score, info.ScoreType, info.BestMove, info.Nodes, info.NPS)

				// Send to analysis channel
				select {
				case analysisChannel <- analysisInfo:
				default:
					// Channel full, skip
				}

			case <-stopCh:
				// Stop signal received, drain any remaining messages and exit
				logger.Debug("ANALYSIS", "Forwarding goroutine received stop signal, draining buffer")
				// Drain the buffer to prevent goroutine leak
				go func() {
					for range builtinCh {
						// Discard remaining messages
					}
				}()
				return
			}
		}
	}()

	return nil
}

// StopAnalysis stops the current analysis
func (e *BuiltinAnalysisEngine) StopAnalysis() {
	// Send stop signal to current analysis
	if e.currentStopCh != nil {
		select {
		case e.currentStopCh <- true:
		default:
		}
	}
	logger.Debug("ANALYSIS", "Built-in engine analysis stop requested")
}

// Close closes the engine
func (e *BuiltinAnalysisEngine) Close() {
	e.StopAnalysis()
	logger.Debug("ANALYSIS", "Built-in analysis engine closed")
}
