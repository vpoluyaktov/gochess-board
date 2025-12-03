package engines

import (
	"sync"
	"testing"
	"time"
)

// mockEngine is a simple mock implementation of ChessEngine for testing
type mockEngine struct {
	closed bool
	mu     sync.Mutex
}

func (m *mockEngine) GetBestMove(fen string, moveTime time.Duration) (string, error) {
	return "e2e4", nil
}

func (m *mockEngine) GetBestMoveWithClock(fen string, moveHistory []string, whiteTime, blackTime, whiteInc, blackInc time.Duration) (string, error) {
	return "e2e4", nil
}

func (m *mockEngine) SetOption(name, value string) error {
	return nil
}

func (m *mockEngine) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockEngine) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// mockEngineFactory creates mock engines for testing
func mockEngineFactory(enginePath, engineName, engineType string) (ChessEngine, error) {
	return &mockEngine{}, nil
}

func TestNewEnginePool(t *testing.T) {
	pool := NewEnginePool(5*time.Minute, mockEngineFactory)
	defer pool.Close()

	if pool == nil {
		t.Fatal("Expected non-nil pool")
	}

	stats := pool.Stats()
	if stats["active_engines"].(int) != 0 {
		t.Errorf("Expected 0 active engines, got %d", stats["active_engines"].(int))
	}
}

func TestEnginePool_GetOrCreateEngine(t *testing.T) {
	pool := NewEnginePool(5*time.Minute, mockEngineFactory)
	defer pool.Close()

	// Create first engine
	engine1, err := pool.GetOrCreateEngine("game1", "/usr/bin/stockfish", "Stockfish", "uci", nil)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	if engine1 == nil {
		t.Fatal("Expected non-nil engine")
	}

	stats := pool.Stats()
	if stats["active_engines"].(int) != 1 {
		t.Errorf("Expected 1 active engine, got %d", stats["active_engines"].(int))
	}

	// Get same engine again (should reuse)
	engine2, err := pool.GetOrCreateEngine("game1", "/usr/bin/stockfish", "Stockfish", "uci", nil)
	if err != nil {
		t.Fatalf("Failed to get engine: %v", err)
	}
	if engine1 != engine2 {
		t.Error("Expected same engine instance to be reused")
	}

	stats = pool.Stats()
	if stats["active_engines"].(int) != 1 {
		t.Errorf("Expected 1 active engine after reuse, got %d", stats["active_engines"].(int))
	}
}

func TestEnginePool_DifferentGames(t *testing.T) {
	pool := NewEnginePool(5*time.Minute, mockEngineFactory)
	defer pool.Close()

	// Create engine for game1
	engine1, err := pool.GetOrCreateEngine("game1", "/usr/bin/stockfish", "Stockfish", "uci", nil)
	if err != nil {
		t.Fatalf("Failed to create engine for game1: %v", err)
	}

	// Create engine for game2 (should be different instance)
	engine2, err := pool.GetOrCreateEngine("game2", "/usr/bin/stockfish", "Stockfish", "uci", nil)
	if err != nil {
		t.Fatalf("Failed to create engine for game2: %v", err)
	}

	if engine1 == engine2 {
		t.Error("Expected different engine instances for different games")
	}

	stats := pool.Stats()
	if stats["active_engines"].(int) != 2 {
		t.Errorf("Expected 2 active engines, got %d", stats["active_engines"].(int))
	}
}

func TestEnginePool_OptionsChange(t *testing.T) {
	pool := NewEnginePool(5*time.Minute, mockEngineFactory)
	defer pool.Close()

	// Create engine with options
	options1 := map[string]string{"UCI_Elo": "1500"}
	engine1, err := pool.GetOrCreateEngine("game1", "/usr/bin/stockfish", "Stockfish", "uci", options1)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Get engine with different options (should create new one)
	options2 := map[string]string{"UCI_Elo": "2000"}
	engine2, err := pool.GetOrCreateEngine("game1", "/usr/bin/stockfish", "Stockfish", "uci", options2)
	if err != nil {
		t.Fatalf("Failed to create engine with new options: %v", err)
	}

	if engine1 == engine2 {
		t.Error("Expected new engine instance when options change")
	}

	// Old engine should be closed
	if !engine1.(*mockEngine).IsClosed() {
		t.Error("Expected old engine to be closed when options change")
	}
}

func TestEnginePool_CloseEngine(t *testing.T) {
	pool := NewEnginePool(5*time.Minute, mockEngineFactory)
	defer pool.Close()

	// Create engine
	engine, err := pool.GetOrCreateEngine("game1", "/usr/bin/stockfish", "Stockfish", "uci", nil)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	stats := pool.Stats()
	if stats["active_engines"].(int) != 1 {
		t.Errorf("Expected 1 active engine, got %d", stats["active_engines"].(int))
	}

	// Close engine explicitly
	pool.CloseEngine("game1", "/usr/bin/stockfish")

	stats = pool.Stats()
	if stats["active_engines"].(int) != 0 {
		t.Errorf("Expected 0 active engines after close, got %d", stats["active_engines"].(int))
	}

	if !engine.(*mockEngine).IsClosed() {
		t.Error("Expected engine to be closed")
	}
}

func TestEnginePool_CloseAllEnginesForGame(t *testing.T) {
	pool := NewEnginePool(5*time.Minute, mockEngineFactory)
	defer pool.Close()

	// Create multiple engines for same game
	engine1, _ := pool.GetOrCreateEngine("game1", "/usr/bin/stockfish", "Stockfish", "uci", nil)
	engine2, _ := pool.GetOrCreateEngine("game1", "/usr/bin/fruit", "Fruit", "uci", nil)
	pool.GetOrCreateEngine("game2", "/usr/bin/stockfish", "Stockfish", "uci", nil)

	stats := pool.Stats()
	if stats["active_engines"].(int) != 3 {
		t.Errorf("Expected 3 active engines, got %d", stats["active_engines"].(int))
	}

	// Close all engines for game1
	pool.CloseAllEnginesForGame("game1")

	stats = pool.Stats()
	if stats["active_engines"].(int) != 1 {
		t.Errorf("Expected 1 active engine after closing game1, got %d", stats["active_engines"].(int))
	}

	if !engine1.(*mockEngine).IsClosed() {
		t.Error("Expected engine1 to be closed")
	}
	if !engine2.(*mockEngine).IsClosed() {
		t.Error("Expected engine2 to be closed")
	}
}

func TestEnginePool_IdleCleanup(t *testing.T) {
	// Use very short timeout for testing
	pool := NewEnginePool(100*time.Millisecond, mockEngineFactory)
	defer pool.Close()

	// Create engine
	engine, err := pool.GetOrCreateEngine("game1", "/usr/bin/stockfish", "Stockfish", "uci", nil)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	stats := pool.Stats()
	if stats["active_engines"].(int) != 1 {
		t.Errorf("Expected 1 active engine, got %d", stats["active_engines"].(int))
	}

	// Wait for idle timeout
	time.Sleep(150 * time.Millisecond)

	// Trigger cleanup manually
	pool.cleanupIdleEngines()

	stats = pool.Stats()
	if stats["active_engines"].(int) != 0 {
		t.Errorf("Expected 0 active engines after idle cleanup, got %d", stats["active_engines"].(int))
	}

	if !engine.(*mockEngine).IsClosed() {
		t.Error("Expected idle engine to be closed")
	}
}

func TestEnginePool_ReleaseEngine(t *testing.T) {
	pool := NewEnginePool(5*time.Minute, mockEngineFactory)
	defer pool.Close()

	// Create engine
	_, err := pool.GetOrCreateEngine("game1", "/usr/bin/stockfish", "Stockfish", "uci", nil)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Get initial last used time
	pool.mu.Lock()
	key := makeKey("game1", "/usr/bin/stockfish")
	initialTime := pool.engines[key].LastUsed
	pool.mu.Unlock()

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Release engine (updates last used time)
	pool.ReleaseEngine("game1", "/usr/bin/stockfish")

	// Check that last used time was updated
	pool.mu.Lock()
	newTime := pool.engines[key].LastUsed
	pool.mu.Unlock()

	if !newTime.After(initialTime) {
		t.Error("Expected last used time to be updated after release")
	}
}

func TestEnginePool_ConcurrentAccess(t *testing.T) {
	pool := NewEnginePool(5*time.Minute, mockEngineFactory)
	defer pool.Close()

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(gameNum int) {
			defer wg.Done()
			gameID := "game" + string(rune('0'+gameNum))
			for j := 0; j < 10; j++ {
				_, err := pool.GetOrCreateEngine(gameID, "/usr/bin/stockfish", "Stockfish", "uci", nil)
				if err != nil {
					t.Errorf("Failed to get engine: %v", err)
				}
				pool.ReleaseEngine(gameID, "/usr/bin/stockfish")
			}
		}(i)
	}

	wg.Wait()

	stats := pool.Stats()
	activeEngines := stats["active_engines"].(int)
	if activeEngines < 1 || activeEngines > numGoroutines {
		t.Errorf("Expected between 1 and %d active engines, got %d", numGoroutines, activeEngines)
	}
}

func TestOptionsMatch(t *testing.T) {
	tests := []struct {
		name     string
		a        map[string]string
		b        map[string]string
		expected bool
	}{
		{"both nil", nil, nil, true},
		{"both empty", map[string]string{}, map[string]string{}, true},
		{"same options", map[string]string{"a": "1"}, map[string]string{"a": "1"}, true},
		{"different values", map[string]string{"a": "1"}, map[string]string{"a": "2"}, false},
		{"different keys", map[string]string{"a": "1"}, map[string]string{"b": "1"}, false},
		{"different lengths", map[string]string{"a": "1"}, map[string]string{"a": "1", "b": "2"}, false},
		{"one nil", map[string]string{"a": "1"}, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optionsMatch(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("optionsMatch(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMakeKey(t *testing.T) {
	key := makeKey("game123", "/usr/bin/stockfish")
	expected := "game123:/usr/bin/stockfish"
	if key != expected {
		t.Errorf("makeKey() = %q, expected %q", key, expected)
	}
}
