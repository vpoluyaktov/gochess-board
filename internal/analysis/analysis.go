package analysis

// PVLine represents a single principal variation line with its score
type PVLine struct {
	Score     int      `json:"score"`     // centipawns
	ScoreType string   `json:"scoreType"` // "cp" or "mate"
	Moves     []string `json:"moves"`     // moves in this variation
}

// AnalysisInfo represents engine analysis data
type AnalysisInfo struct {
	Depth     int      `json:"depth"`
	Score     int      `json:"score"`    // centipawns (for backward compatibility)
	BestMove  string   `json:"bestMove"` // e.g., "e2e4" (for backward compatibility)
	PV        []string `json:"pv"`       // principal variation (for backward compatibility)
	Nodes     int64    `json:"nodes"`
	NPS       int64    `json:"nps"`       // nodes per second
	Time      int      `json:"time"`      // milliseconds
	ScoreType string   `json:"scoreType"` // "cp" or "mate" (for backward compatibility)
	MultiPV   []PVLine `json:"multiPV"`   // multiple PV lines (for 3 best moves feature)
	FEN       string   `json:"fen"`       // position being analyzed (for sync verification)
}

// AnalysisEngineInterface is an interface for analysis engines (UCI or CECP)
type AnalysisEngineInterface interface {
	StartAnalysis(fen string, analysisChannel chan<- AnalysisInfo) error
	StopAnalysis()
	Close()
}
