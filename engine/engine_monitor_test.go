package engine

import (
	"testing"
	"time"
)

func TestEngineMonitor_RegisterEngine(t *testing.T) {
	monitor := &EngineMonitor{
		engines: make(map[string]*ActiveEngine),
	}

	engine := &ActiveEngine{
		Name:      "Stockfish",
		Path:      "/usr/bin/stockfish",
		ELO:       2000,
		SessionID: "session-1",
		StartTime: time.Now(),
	}

	monitor.RegisterEngine("session-1", engine)

	if len(monitor.engines) != 1 {
		t.Errorf("Expected 1 engine, got %d", len(monitor.engines))
	}

	registered, exists := monitor.engines["session-1"]
	if !exists {
		t.Error("Engine was not registered")
	}
	if registered.Name != "Stockfish" {
		t.Errorf("Expected engine name 'Stockfish', got %q", registered.Name)
	}
}

func TestEngineMonitor_UnregisterEngine(t *testing.T) {
	monitor := &EngineMonitor{
		engines: make(map[string]*ActiveEngine),
	}

	engine := &ActiveEngine{
		Name:      "Stockfish",
		SessionID: "session-1",
		StartTime: time.Now(),
	}

	monitor.RegisterEngine("session-1", engine)
	monitor.UnregisterEngine("session-1")

	if len(monitor.engines) != 0 {
		t.Errorf("Expected 0 engines after unregister, got %d", len(monitor.engines))
	}
}

func TestEngineMonitor_UnregisterNonExistent(t *testing.T) {
	monitor := &EngineMonitor{
		engines: make(map[string]*ActiveEngine),
	}

	// Should not panic when unregistering non-existent engine
	monitor.UnregisterEngine("non-existent")

	if len(monitor.engines) != 0 {
		t.Errorf("Expected 0 engines, got %d", len(monitor.engines))
	}
}

func TestEngineMonitor_GetActiveEngines(t *testing.T) {
	monitor := &EngineMonitor{
		engines: make(map[string]*ActiveEngine),
	}

	engine1 := &ActiveEngine{
		Name:      "Stockfish",
		SessionID: "session-1",
		StartTime: time.Now(),
	}
	engine2 := &ActiveEngine{
		Name:      "Fruit",
		SessionID: "session-2",
		StartTime: time.Now(),
	}

	monitor.RegisterEngine("session-1", engine1)
	monitor.RegisterEngine("session-2", engine2)

	active := monitor.GetActiveEngines()

	if len(active) != 2 {
		t.Errorf("Expected 2 active engines, got %d", len(active))
	}

	// Verify engines are in the list
	names := make(map[string]bool)
	for _, e := range active {
		names[e.Name] = true
	}

	if !names["Stockfish"] {
		t.Error("Expected Stockfish in active engines")
	}
	if !names["Fruit"] {
		t.Error("Expected Fruit in active engines")
	}
}

func TestEngineMonitor_GetActiveEnginesEmpty(t *testing.T) {
	monitor := &EngineMonitor{
		engines: make(map[string]*ActiveEngine),
	}

	active := monitor.GetActiveEngines()

	if len(active) != 0 {
		t.Errorf("Expected 0 active engines, got %d", len(active))
	}
	if active == nil {
		t.Error("Expected empty slice, got nil")
	}
}

func TestEngineMonitor_CleanupStaleEngines(t *testing.T) {
	monitor := &EngineMonitor{
		engines: make(map[string]*ActiveEngine),
	}

	now := time.Now()

	// Add old engine (started 2 hours ago)
	oldEngine := &ActiveEngine{
		Name:      "Old Engine",
		SessionID: "old-session",
		StartTime: now.Add(-2 * time.Hour),
	}

	// Add recent engine (started 5 minutes ago)
	recentEngine := &ActiveEngine{
		Name:      "Recent Engine",
		SessionID: "recent-session",
		StartTime: now.Add(-5 * time.Minute),
	}

	monitor.RegisterEngine("old-session", oldEngine)
	monitor.RegisterEngine("recent-session", recentEngine)

	// Cleanup engines older than 1 hour
	monitor.CleanupStaleEngines(1 * time.Hour)

	active := monitor.GetActiveEngines()

	if len(active) != 1 {
		t.Errorf("Expected 1 engine after cleanup, got %d", len(active))
	}

	if active[0].Name != "Recent Engine" {
		t.Errorf("Expected 'Recent Engine' to remain, got %q", active[0].Name)
	}
}

func TestEngineMonitor_CleanupStaleEnginesNone(t *testing.T) {
	monitor := &EngineMonitor{
		engines: make(map[string]*ActiveEngine),
	}

	now := time.Now()

	// Add recent engines
	engine1 := &ActiveEngine{
		Name:      "Engine 1",
		SessionID: "session-1",
		StartTime: now.Add(-5 * time.Minute),
	}
	engine2 := &ActiveEngine{
		Name:      "Engine 2",
		SessionID: "session-2",
		StartTime: now.Add(-10 * time.Minute),
	}

	monitor.RegisterEngine("session-1", engine1)
	monitor.RegisterEngine("session-2", engine2)

	// Cleanup engines older than 1 hour (none should be removed)
	monitor.CleanupStaleEngines(1 * time.Hour)

	active := monitor.GetActiveEngines()

	if len(active) != 2 {
		t.Errorf("Expected 2 engines after cleanup, got %d", len(active))
	}
}

func TestEngineMonitor_CleanupStaleEnginesAll(t *testing.T) {
	monitor := &EngineMonitor{
		engines: make(map[string]*ActiveEngine),
	}

	now := time.Now()

	// Add old engines
	engine1 := &ActiveEngine{
		Name:      "Engine 1",
		SessionID: "session-1",
		StartTime: now.Add(-2 * time.Hour),
	}
	engine2 := &ActiveEngine{
		Name:      "Engine 2",
		SessionID: "session-2",
		StartTime: now.Add(-3 * time.Hour),
	}

	monitor.RegisterEngine("session-1", engine1)
	monitor.RegisterEngine("session-2", engine2)

	// Cleanup engines older than 1 hour (all should be removed)
	monitor.CleanupStaleEngines(1 * time.Hour)

	active := monitor.GetActiveEngines()

	if len(active) != 0 {
		t.Errorf("Expected 0 engines after cleanup, got %d", len(active))
	}
}

func TestEngineMonitor_ConcurrentAccess(t *testing.T) {
	monitor := &EngineMonitor{
		engines: make(map[string]*ActiveEngine),
	}

	// Test concurrent registration and retrieval
	done := make(chan bool)

	// Goroutine 1: Register engines
	go func() {
		for i := 0; i < 100; i++ {
			engine := &ActiveEngine{
				Name:      "Engine",
				SessionID: string(rune(i)),
				StartTime: time.Now(),
			}
			monitor.RegisterEngine(string(rune(i)), engine)
		}
		done <- true
	}()

	// Goroutine 2: Get active engines
	go func() {
		for i := 0; i < 100; i++ {
			_ = monitor.GetActiveEngines()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Should not panic and should have some engines
	active := monitor.GetActiveEngines()
	if len(active) == 0 {
		t.Error("Expected some engines after concurrent operations")
	}
}

func TestGlobalMonitor(t *testing.T) {
	// Test that GlobalMonitor is initialized
	if GlobalMonitor == nil {
		t.Error("GlobalMonitor should not be nil")
	}

	if GlobalMonitor.engines == nil {
		t.Error("GlobalMonitor.engines should not be nil")
	}

	// Test backward compatibility alias
	if globalMonitor != GlobalMonitor {
		t.Error("globalMonitor should be an alias for GlobalMonitor")
	}
}

func TestActiveEngine_Fields(t *testing.T) {
	now := time.Now()
	engine := &ActiveEngine{
		Name:           "Stockfish",
		Path:           "/usr/bin/stockfish",
		ELO:            2850,
		WhiteTime:      300000,
		BlackTime:      300000,
		WhiteIncrement: 5000,
		BlackIncrement: 5000,
		StartTime:      now,
		SessionID:      "test-session",
	}

	if engine.Name != "Stockfish" {
		t.Errorf("Expected Name 'Stockfish', got %q", engine.Name)
	}
	if engine.ELO != 2850 {
		t.Errorf("Expected ELO 2850, got %d", engine.ELO)
	}
	if engine.SessionID != "test-session" {
		t.Errorf("Expected SessionID 'test-session', got %q", engine.SessionID)
	}
	if !engine.StartTime.Equal(now) {
		t.Error("StartTime does not match")
	}
}
