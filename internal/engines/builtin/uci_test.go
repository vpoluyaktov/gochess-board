package builtin

import (
	"strings"
	"testing"
	"time"
)

func TestParseGoCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		fen     string
		minTime time.Duration
		maxTime time.Duration
	}{
		{
			name:    "movetime",
			command: "go movetime 5000",
			fen:     "",
			minTime: 4500 * time.Millisecond,
			maxTime: 5000 * time.Millisecond,
		},
		{
			name:    "infinite",
			command: "go infinite",
			fen:     "",
			minTime: 30 * time.Minute,
			maxTime: 2 * time.Hour,
		},
		{
			name:    "wtime only - white to move",
			command: "go wtime 60000",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			minTime: 500 * time.Millisecond,
			maxTime: 15 * time.Second,
		},
		{
			name:    "btime only - black to move",
			command: "go btime 60000",
			fen:     "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1",
			minTime: 500 * time.Millisecond,
			maxTime: 15 * time.Second,
		},
		{
			name:    "wtime with increment - white to move",
			command: "go wtime 60000 winc 1000",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			minTime: 1 * time.Second,
			maxTime: 15 * time.Second,
		},
		{
			name:    "movestogo",
			command: "go wtime 60000 movestogo 10",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			minTime: 5 * time.Second,
			maxTime: 15 * time.Second,
		},
		{
			name:    "low time emergency",
			command: "go wtime 3000",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			minTime: 100 * time.Millisecond,
			maxTime: 500 * time.Millisecond,
		},
		{
			name:    "full time control",
			command: "go wtime 300000 btime 300000 winc 5000 binc 5000",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			minTime: 5 * time.Second,
			maxTime: 75 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Fields(tt.command)
			thinkTime := parseGoCommandWithFEN(parts, tt.fen)

			if thinkTime < tt.minTime {
				t.Errorf("think time %v is less than minimum %v", thinkTime, tt.minTime)
			}
			if thinkTime > tt.maxTime {
				t.Errorf("think time %v is greater than maximum %v", thinkTime, tt.maxTime)
			}

			t.Logf("%s: %v", tt.name, thinkTime)
		})
	}
}

func TestTimeControlStruct(t *testing.T) {
	// Test that TimeControl correctly identifies whose turn it is
	tc := TimeControl{
		WhiteTime:   60 * time.Second,
		BlackTime:   55 * time.Second,
		WhiteInc:    2 * time.Second,
		BlackInc:    2 * time.Second,
		IsWhiteTurn: true,
	}

	thinkTime := calculateThinkTime(tc)
	t.Logf("White to move with 60s+2s: %v", thinkTime)

	// Should use white's time
	if thinkTime < 1*time.Second || thinkTime > 20*time.Second {
		t.Errorf("unexpected think time for white: %v", thinkTime)
	}

	// Now test black's turn
	tc.IsWhiteTurn = false
	thinkTime = calculateThinkTime(tc)
	t.Logf("Black to move with 55s+2s: %v", thinkTime)

	// Should use black's time
	if thinkTime < 1*time.Second || thinkTime > 20*time.Second {
		t.Errorf("unexpected think time for black: %v", thinkTime)
	}
}

func TestEmergencyTimeManagement(t *testing.T) {
	// Test that engine doesn't use too much time when low on time
	tc := TimeControl{
		WhiteTime:   2 * time.Second,
		WhiteInc:    0,
		IsWhiteTurn: true,
	}

	thinkTime := calculateThinkTime(tc)
	t.Logf("Emergency (2s remaining): %v", thinkTime)

	// Should be very fast
	if thinkTime > 500*time.Millisecond {
		t.Errorf("think time %v is too long for emergency situation", thinkTime)
	}
	if thinkTime < 100*time.Millisecond {
		t.Errorf("think time %v is too short", thinkTime)
	}
}
