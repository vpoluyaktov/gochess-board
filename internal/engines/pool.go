package engines

import (
	"fmt"
	"sync"
	"time"

	"gochess-board/internal/logger"
)

// EnginePool manages persistent engine instances with automatic cleanup
type EnginePool struct {
	mu            sync.Mutex
	engines       map[string]*PooledEngine
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
	idleTimeout   time.Duration
	engineFactory EngineFactory
}

// PooledEngine represents a cached engine instance
type PooledEngine struct {
	Engine     ChessEngine
	GameID     string
	EnginePath string
	EngineType string // "uci", "cecp", or "internal"
	EngineName string
	LastUsed   time.Time
	Options    map[string]string // Cached options for comparison
}

// EngineFactory is a function type for creating new engines
type EngineFactory func(enginePath, engineName, engineType string) (ChessEngine, error)

// DefaultIdleTimeout is the default time after which idle engines are closed
const DefaultIdleTimeout = 10 * time.Minute

// CleanupInterval is how often the pool checks for idle engines
const CleanupInterval = 1 * time.Minute

// NewEnginePool creates a new engine pool with the specified idle timeout
func NewEnginePool(idleTimeout time.Duration, factory EngineFactory) *EnginePool {
	if idleTimeout <= 0 {
		idleTimeout = DefaultIdleTimeout
	}

	pool := &EnginePool{
		engines:       make(map[string]*PooledEngine),
		idleTimeout:   idleTimeout,
		stopCleanup:   make(chan struct{}),
		engineFactory: factory,
	}

	// Start background cleanup goroutine
	pool.cleanupTicker = time.NewTicker(CleanupInterval)
	go pool.cleanupLoop()

	logger.Info("ENGINE_POOL", "Engine pool started with %v idle timeout", idleTimeout)
	return pool
}

// cleanupLoop runs in the background and removes idle engines
func (p *EnginePool) cleanupLoop() {
	for {
		select {
		case <-p.cleanupTicker.C:
			p.cleanupIdleEngines()
		case <-p.stopCleanup:
			return
		}
	}
}

// cleanupIdleEngines removes engines that have been idle for too long
func (p *EnginePool) cleanupIdleEngines() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for key, pooled := range p.engines {
		if now.Sub(pooled.LastUsed) > p.idleTimeout {
			logger.Info("ENGINE_POOL", "Closing idle engine: %s (game: %s, idle for %v)",
				pooled.EngineName, pooled.GameID, now.Sub(pooled.LastUsed).Round(time.Second))
			pooled.Engine.Close()
			delete(p.engines, key)
		}
	}
}

// makeKey creates a unique key for engine lookup
// Key is based on gameID + enginePath to allow different engines per game
func makeKey(gameID, enginePath string) string {
	return fmt.Sprintf("%s:%s", gameID, enginePath)
}

// GetOrCreateEngine returns an existing engine for the game or creates a new one
func (p *EnginePool) GetOrCreateEngine(gameID, enginePath, engineName, engineType string, options map[string]string) (ChessEngine, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := makeKey(gameID, enginePath)

	// Check if we have a cached engine for this game
	if pooled, exists := p.engines[key]; exists {
		// Check if options have changed (e.g., ELO setting)
		if optionsMatch(pooled.Options, options) {
			pooled.LastUsed = time.Now()
			logger.Debug("ENGINE_POOL", "Reusing cached engine: %s (game: %s)", engineName, gameID)
			return pooled.Engine, nil
		}
		// Options changed, close old engine and create new one
		logger.Info("ENGINE_POOL", "Options changed, recreating engine: %s (game: %s)", engineName, gameID)
		pooled.Engine.Close()
		delete(p.engines, key)
	}

	// Create new engine
	logger.Info("ENGINE_POOL", "Creating new engine: %s (game: %s, type: %s)", engineName, gameID, engineType)
	engine, err := p.engineFactory(enginePath, engineName, engineType)
	if err != nil {
		return nil, err
	}

	// Apply options
	for optName, optValue := range options {
		if err := engine.SetOption(optName, optValue); err != nil {
			logger.Warn("ENGINE_POOL", "Failed to set option %s=%s: %v", optName, optValue, err)
		}
	}

	// Cache the engine
	p.engines[key] = &PooledEngine{
		Engine:     engine,
		GameID:     gameID,
		EnginePath: enginePath,
		EngineType: engineType,
		EngineName: engineName,
		LastUsed:   time.Now(),
		Options:    copyOptions(options),
	}

	return engine, nil
}

// ReleaseEngine marks an engine as available (updates last used time)
// In the pool model, we don't actually release - we just update the timestamp
func (p *EnginePool) ReleaseEngine(gameID, enginePath string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := makeKey(gameID, enginePath)
	if pooled, exists := p.engines[key]; exists {
		pooled.LastUsed = time.Now()
	}
}

// CloseEngine explicitly closes and removes an engine from the pool
// Useful when a game ends or user switches engines
func (p *EnginePool) CloseEngine(gameID, enginePath string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := makeKey(gameID, enginePath)
	if pooled, exists := p.engines[key]; exists {
		logger.Info("ENGINE_POOL", "Explicitly closing engine: %s (game: %s)", pooled.EngineName, gameID)
		pooled.Engine.Close()
		delete(p.engines, key)
	}
}

// CloseAllEnginesForGame closes all engines associated with a game
func (p *EnginePool) CloseAllEnginesForGame(gameID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, pooled := range p.engines {
		if pooled.GameID == gameID {
			logger.Info("ENGINE_POOL", "Closing engine for ended game: %s (game: %s)", pooled.EngineName, gameID)
			pooled.Engine.Close()
			delete(p.engines, key)
		}
	}
}

// Close shuts down the engine pool and all cached engines
func (p *EnginePool) Close() {
	// Stop cleanup goroutine
	p.cleanupTicker.Stop()
	close(p.stopCleanup)

	// Close all engines
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, pooled := range p.engines {
		logger.Info("ENGINE_POOL", "Shutting down engine: %s (game: %s)", pooled.EngineName, pooled.GameID)
		pooled.Engine.Close()
		delete(p.engines, key)
	}

	logger.Info("ENGINE_POOL", "Engine pool shut down")
}

// Stats returns statistics about the pool
func (p *EnginePool) Stats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := map[string]interface{}{
		"active_engines": len(p.engines),
		"idle_timeout":   p.idleTimeout.String(),
	}

	engineList := make([]map[string]interface{}, 0, len(p.engines))
	for _, pooled := range p.engines {
		engineList = append(engineList, map[string]interface{}{
			"name":      pooled.EngineName,
			"game_id":   pooled.GameID,
			"type":      pooled.EngineType,
			"idle_time": fmt.Sprintf("%ds", int(time.Since(pooled.LastUsed).Seconds())),
		})
	}
	stats["engines"] = engineList

	return stats
}

// optionsMatch checks if two option maps are equivalent
func optionsMatch(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// copyOptions creates a copy of the options map
func copyOptions(opts map[string]string) map[string]string {
	if opts == nil {
		return nil
	}
	copy := make(map[string]string, len(opts))
	for k, v := range opts {
		copy[k] = v
	}
	return copy
}

// GlobalEnginePool is the singleton engine pool instance (nil if not using persistent engines)
var GlobalEnginePool *EnginePool
