package engines

import (
	"sync"
	"time"
)

// ActiveEngine represents a currently running engine instance
type ActiveEngine struct {
	Name           string    `json:"name"`
	Path           string    `json:"path"`
	ELO            int       `json:"elo"`
	WhiteTime      int       `json:"whiteTime"`
	BlackTime      int       `json:"blackTime"`
	WhiteIncrement int       `json:"whiteIncrement"`
	BlackIncrement int       `json:"blackIncrement"`
	StartTime      time.Time `json:"startTime"`
	SessionID      string    `json:"sessionId"`
}

// EngineMonitor tracks active engine instances
type EngineMonitor struct {
	mu      sync.RWMutex
	engines map[string]*ActiveEngine
}

var GlobalMonitor = &EngineMonitor{
	engines: make(map[string]*ActiveEngine),
}

// Alias for backward compatibility
var globalMonitor = GlobalMonitor

// RegisterEngine adds an engine to the active list
func (em *EngineMonitor) RegisterEngine(sessionID string, engine *ActiveEngine) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.engines[sessionID] = engine
}

// UnregisterEngine removes an engine from the active list
func (em *EngineMonitor) UnregisterEngine(sessionID string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	delete(em.engines, sessionID)
}

// GetActiveEngines returns a list of all active engines
func (em *EngineMonitor) GetActiveEngines() []*ActiveEngine {
	em.mu.RLock()
	defer em.mu.RUnlock()

	engines := make([]*ActiveEngine, 0, len(em.engines))
	for _, engine := range em.engines {
		engines = append(engines, engine)
	}
	return engines
}

// CleanupStaleEngines removes engines that have been running too long (likely crashed)
func (em *EngineMonitor) CleanupStaleEngines(maxAge time.Duration) {
	em.mu.Lock()
	defer em.mu.Unlock()

	now := time.Now()
	for id, engine := range em.engines {
		if now.Sub(engine.StartTime) > maxAge {
			delete(em.engines, id)
		}
	}
}
