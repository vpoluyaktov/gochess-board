# Chess Opening Recognition Feature

## Overview
Implemented an in-memory trie-based opening book that provides fast lookup of chess opening names based on move sequences.

## Architecture

### Data Structure
- **Trie (Prefix Tree)**: Each node represents a move in an opening sequence
- **Storage**: In-memory structure built at server startup
- **Source Data**: 5 TSV files from lichess-org/chess-openings repository
  - `a.tsv`, `b.tsv`, `c.tsv`, `d.tsv`, `e.tsv`
  - Total: ~360KB, 3,594 openings, 8,147 nodes
  - Max depth: 36 moves

### Components

#### 1. Opening Book (`server/opening.go`)
- **OpeningBook**: Main structure with thread-safe trie
- **OpeningNode**: Trie node with move, children, and opening info
- **OpeningInfo**: ECO code, name, and PGN notation

#### 2. Key Functions
- `LoadFromDirectory()`: Loads all TSV files from a directory
- `Lookup(moves []string)`: Finds opening by SAN move sequence
- `LookupByGame(game *chess.Game)`: Finds opening from chess.Game object
- `Stats()`: Returns statistics about the loaded opening book

#### 3. API Endpoint (`/api/opening`)
- **Method**: POST
- **Request**: `{"moves": ["e4", "e5", "Nf3", "Nc6"]}`
- **Response**: `{"eco": "C44", "name": "King's Knight Opening", "pgn": "1. e4 e5 2. Nf3 Nc6"}`

## Performance

### Loading Time
- ~7-8 seconds to load all 5 TSV files at startup
- One-time cost, happens during server initialization

### Lookup Time
- O(m) where m = number of moves
- Typically microseconds for any lookup
- No disk I/O after initial load

### Memory Usage
- ~1-2 MB for the entire trie structure
- Negligible compared to typical server memory

## Implementation Details

### Move Notation
- **Storage**: SAN (Standard Algebraic Notation) - e.g., "e4", "Nf3", "Bc4"
- **Conversion**: UCI moves from chess.Game are converted to SAN for lookup
- **Reason**: TSV files use SAN notation, and it's more human-readable

### Longest Prefix Matching
- Returns the deepest matching opening in the trie
- If exact sequence not found, returns the longest matching prefix
- Example: "e4 e5 Nf3 Nc6 Bc4 Bc5 d3" might match "Italian Game" at move 5

### Thread Safety
- Uses `sync.RWMutex` for concurrent access
- Multiple goroutines can safely lookup simultaneously
- Write lock only during initialization

## Testing

### Test Coverage
- `TestOpeningBook`: Tests common openings (Italian, Sicilian, French, Queen's Gambit)
- `TestOpeningBookByGame`: Tests lookup from chess.Game object
- `TestOpeningBookNoMatch`: Tests behavior with non-standard openings

### Test Results
```
Opening book stats: map[max_depth:36 total_nodes:8147 total_openings:3594]
✓ Italian Game (C50)
✓ Sicilian Defense (B20)
✓ French Defense (C00)
✓ Queen's Gambit (D06)
✓ King's Knight Opening: Normal Variation (C44)
```

## Files Modified/Created

### New Files
- `server/opening.go` - Opening book implementation
- `server/opening_test.go` - Test suite
- `server/assets/openings/*.tsv` - Opening database (5 files)
- `OPENING_FEATURE.md` - This documentation

### Modified Files
- `server/server.go` - Added opening book initialization and API endpoint
  - Added `openingBook *OpeningBook` field to Server struct
  - Added `/api/opening` endpoint
  - Added `handleGetOpening()` method

## Usage

### Server Side
```go
// Opening book is automatically loaded at server startup
srv := server.New(addr, bookFile)
```

### Client Side (JavaScript)
```javascript
// Get opening name for current position
const response = await fetch('/api/opening', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        moves: ['e4', 'e5', 'Nf3', 'Nc6', 'Bc4']
    })
});

const opening = await response.json();
console.log(`${opening.name} (${opening.eco})`);
// Output: "Italian Game (C50)"
```

## Next Steps (UI Integration)

To display the opening name in the web UI:

1. **Track Move History**: Maintain SAN move list in JavaScript
2. **Call API After Each Move**: Query `/api/opening` endpoint
3. **Display Opening Info**: Show ECO code and name in the UI
4. **Update on Position Change**: Refresh when moves are made/undone

### Suggested UI Placement
- Below the chess board
- In a dedicated "Opening" panel
- As a tooltip on hover
- In the move history sidebar

## Data Source

Opening data from: https://github.com/lichess-org/chess-openings
- Commit: 3379fcd843d3aa24425a0aef51dc00cfe28a8071
- License: Public Domain (as per lichess)
- Maintained by the Lichess community

### Updating the Opening Database

To update the opening database to a newer version:

```bash
# Update to latest commit (default)
./update-openings.sh

# Update to specific commit
./update-openings.sh <commit-hash>

# Show help
./update-openings.sh --help
```

The script will:
1. Download all 5 TSV files from the repository
2. Save them to `server/assets/openings/`
3. Show statistics about the downloaded files
4. Suggest running tests to verify the update

After updating, rebuild the application:
```bash
go build -o go-chess .
```

## Benefits

1. **Fast**: Microsecond lookups, no database queries
2. **Offline**: No external API calls required
3. **Comprehensive**: 3,594 openings covering all major variations
4. **Accurate**: Data from Lichess, a trusted source
5. **Scalable**: Can handle thousands of concurrent lookups
6. **Memory Efficient**: Only ~1-2 MB for entire database
