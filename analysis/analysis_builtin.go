package analysis

import (
	"gochess-board/engines/builtin"
	"gochess-board/logger"
)

// BuiltinAnalysisEngine manages the built-in engine for analysis
type BuiltinAnalysisEngine struct {
	engine *builtin.InternalEngine
	stopCh chan bool
}

// NewBuiltinAnalysisEngine creates a new built-in analysis engine
func NewBuiltinAnalysisEngine() (*BuiltinAnalysisEngine, error) {
	logger.Info("ANALYSIS", "Initializing built-in analysis engine")

	return &BuiltinAnalysisEngine{
		engine: builtin.NewEngine(),
		stopCh: make(chan bool, 1),
	}, nil
}

// StartAnalysis starts analyzing a position
func (e *BuiltinAnalysisEngine) StartAnalysis(fen string, analysisChannel chan<- AnalysisInfo) error {
	logger.Info("ANALYSIS", "Starting built-in engine analysis for position: %s", fen)

	// Create a channel to receive builtin.AnalysisInfo
	builtinCh := make(chan builtin.AnalysisInfo, 10)

	// Start analysis in a goroutine
	go func() {
		const maxDepth = 6
		err := e.engine.Analyze(fen, maxDepth, e.stopCh, builtinCh)
		if err != nil {
			logger.Error("ANALYSIS", "Built-in engine analysis error: %v", err)
		}
	}()

	// Forward analysis info from builtin channel to analysis channel
	go func() {
		for info := range builtinCh {
			// Convert builtin.AnalysisInfo to analysis.AnalysisInfo
			analysisInfo := AnalysisInfo{
				Depth:     info.Depth,
				Score:     info.Score,
				BestMove:  info.BestMove,
				PV:        []string{info.BestMove}, // Simple PV for now
				Nodes:     int64(info.Nodes),
				NPS:       info.NPS,
				Time:      info.Time,
				ScoreType: info.ScoreType,
			}

			logger.Debug("ANALYSIS", "Built-in engine: depth=%d, score=%d (%s), move=%s, nodes=%d, nps=%d",
				info.Depth, info.Score, info.ScoreType, info.BestMove, info.Nodes, info.NPS)

			// Send to analysis channel
			select {
			case analysisChannel <- analysisInfo:
			default:
				// Channel full, skip
			}
		}
	}()

	return nil
}

// StopAnalysis stops the current analysis
func (e *BuiltinAnalysisEngine) StopAnalysis() {
	// Send stop signal
	select {
	case e.stopCh <- true:
	default:
	}
	logger.Info("ANALYSIS", "Built-in engine analysis stop requested")
}

// Close closes the engine
func (e *BuiltinAnalysisEngine) Close() {
	e.StopAnalysis()
	logger.Info("ANALYSIS", "Built-in analysis engine closed")
}
