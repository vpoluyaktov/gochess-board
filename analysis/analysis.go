package analysis

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

// AnalysisEngineInterface is an interface for analysis engines (UCI or CECP)
type AnalysisEngineInterface interface {
	StartAnalysis(fen string, analysisChannel chan<- AnalysisInfo) error
	StopAnalysis()
	Close()
}
