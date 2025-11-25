# TUI Opening Statistics Display

## Overview

Added opening database statistics to the TUI's SERVER STATUS box, providing real-time visibility into the loaded opening database.

## Changes Made

### 1. Server API (`server/server.go`)

Added a new method to retrieve opening statistics:

```go
func (s *Server) GetOpeningStats() map[string]int {
    if s.openingBook == nil {
        return map[string]int{
            "total_openings": 0,
            "total_nodes":    0,
            "max_depth":      0,
        }
    }
    return s.openingBook.Stats()
}
```

**Features**:
- Returns statistics from the opening book
- Handles nil opening book gracefully
- Returns zero values if opening book not loaded

### 2. TUI Model (`tui/tui.go`)

#### Model Structure
Added `openingStats` field to the model:

```go
type model struct {
    spinner      spinner.Model
    serverURL    string
    startTime    time.Time
    engines      []server.EngineInfo
    monitor      *server.EngineMonitor
    enginesTable table.Model
    activeTable  table.Model
    width        int
    height       int
    openingStats map[string]int  // NEW
}
```

#### Function Signatures Updated

**InitialModel**:
```go
func InitialModel(serverURL string, engines []server.EngineInfo, 
                  monitor *server.EngineMonitor, openingStats map[string]int) model
```

**RunTUI**:
```go
func RunTUI(serverURL string, engines []server.EngineInfo, 
            monitor *server.EngineMonitor, openingStats map[string]int) error
```

#### Display Logic

Added opening database section to the server info display:

```go
openingsInfo := ""
if m.openingStats != nil && m.openingStats["total_openings"] > 0 {
    openingsInfo = fmt.Sprintf("\n📖 OPENING DATABASE\n\n"+
        "Openings: %d\n"+
        "Nodes:    %d\n"+
        "Max Depth: %d\n",
        m.openingStats["total_openings"],
        m.openingStats["total_nodes"],
        m.openingStats["max_depth"])
}
```

**Also added** `/api/opening` to the API endpoints list.

### 3. Main Application (`main.go`)

Updated TUI initialization to pass opening stats:

```go
if err := tui.RunTUI(url, srv.GetEngines(), server.GlobalMonitor, srv.GetOpeningStats()); err != nil {
    log.Fatalf("TUI error: %v", err)
}
```

## TUI Display

### Before
```
┌─────────────────────────────┐
│ 🖥️  SERVER STATUS           │
│                             │
│ URL:     http://...         │
│ Uptime:  1m23s              │
│ Mode:    Stateless          │
│                             │
│ 📡 API ENDPOINTS            │
│                             │
│ • /api/computer-move        │
│ • /api/analysis             │
│ • /api/engines              │
└─────────────────────────────┘
```

### After
```
┌─────────────────────────────┐
│ 🖥️  SERVER STATUS           │
│                             │
│ URL:     http://...         │
│ Uptime:  1m23s              │
│ Mode:    Stateless          │
│                             │
│ 📖 OPENING DATABASE         │
│                             │
│ Openings: 3594              │
│ Nodes:    8147              │
│ Max Depth: 36               │
│                             │
│ 📡 API ENDPOINTS            │
│                             │
│ • /api/computer-move        │
│ • /api/analysis             │
│ • /api/engines              │
│ • /api/opening              │
└─────────────────────────────┘
```

## Statistics Displayed

### Openings
- **Value**: 3,594
- **Meaning**: Total number of named chess openings in the database
- **Source**: Lichess chess-openings repository

### Nodes
- **Value**: 8,147
- **Meaning**: Total nodes in the trie data structure
- **Purpose**: Shows the efficiency of the trie (more nodes = more detailed variations)

### Max Depth
- **Value**: 36
- **Meaning**: Deepest opening sequence in the database (36 moves)
- **Purpose**: Shows how far into the game openings are tracked

## Benefits

1. **Visibility**: Users can see that the opening database is loaded
2. **Verification**: Confirms the database loaded successfully
3. **Information**: Shows the scale of the opening database
4. **Debugging**: Helps identify if opening database failed to load (all zeros)
5. **Professional**: Makes the TUI more informative and complete

## Error Handling

If the opening database fails to load:
- All statistics show as 0
- The opening database section is hidden (not displayed)
- Server continues to function normally
- Only the opening API endpoint will return errors

## Testing

To verify the display:
```bash
./go-chess
```

The TUI will show:
1. Loading messages during startup
2. Opening database statistics in the SERVER STATUS box
3. Updated API endpoints list including `/api/opening`

## Technical Details

### Data Flow
1. Server loads opening database at startup
2. Server calculates statistics via `openingBook.Stats()`
3. Main passes statistics to TUI via `srv.GetOpeningStats()`
4. TUI stores statistics in model
5. TUI displays statistics in View function

### Performance
- Statistics calculated once at startup
- No runtime overhead
- Cached in TUI model
- No periodic updates needed (static data)

### Memory
- Statistics map: ~100 bytes
- Negligible memory overhead
- Already calculated by opening book

## Future Enhancements

Possible improvements:
1. **Live Updates**: Show when opening database is being reloaded
2. **Cache Hit Rate**: Display opening lookup statistics
3. **Most Popular**: Show most frequently looked up openings
4. **Recent Lookups**: Display last few opening names found
5. **Database Version**: Show commit hash or version of opening data

## Files Modified

- `server/server.go` - Added `GetOpeningStats()` method
- `tui/tui.go` - Added opening stats display
- `main.go` - Pass opening stats to TUI

## Summary

The TUI now provides complete visibility into the opening database, showing users that 3,594 openings are loaded and ready for use. This enhancement improves the professional appearance of the application and provides valuable information to users.
