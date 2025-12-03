package analysis

import (
	"gochess-board/engines/builtin"
	"gochess-board/logger"
)

// BuiltinAnalysisEngine manages the built-in engine for analysis
type BuiltinAnalysisEngine struct {
	engine        *builtin.InternalEngine
	currentStopCh chan bool // Stop channel for current analysis
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
				multiPV := make([]PVLine, len(info.MultiPV))
				for i, pv := range info.MultiPV {
					multiPV[i] = PVLine{
						Score:     pv.Score,
						ScoreType: pv.ScoreType,
						Moves:     pv.Moves,
					}
				}

				analysisInfo := AnalysisInfo{
					Depth:     info.Depth,
					Score:     info.Score,
					BestMove:  info.BestMove,
					PV:        info.PV, // Real PV from engine
					Nodes:     int64(info.Nodes),
					NPS:       info.NPS,
					Time:      info.Time,
					ScoreType: info.ScoreType,
					MultiPV:   multiPV,
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
