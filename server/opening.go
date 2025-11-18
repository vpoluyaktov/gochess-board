package server

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/notnil/chess"
)

// OpeningInfo contains information about a chess opening
type OpeningInfo struct {
	ECO  string
	Name string
	PGN  string
}

// OpeningNode represents a node in the opening trie
type OpeningNode struct {
	Move     string                  // The move that leads to this node (in SAN notation)
	Children map[string]*OpeningNode // Map of SAN move -> child node
	Opening  *OpeningInfo            // Non-nil if this position corresponds to a named opening
}

// OpeningBook manages the opening database
type OpeningBook struct {
	root *OpeningNode
	mu   sync.RWMutex
}

// NewOpeningBook creates a new opening book
func NewOpeningBook() *OpeningBook {
	return &OpeningBook{
		root: &OpeningNode{
			Children: make(map[string]*OpeningNode),
		},
	}
}

// LoadFromEmbedded loads all TSV files from an embedded filesystem
func (ob *OpeningBook) LoadFromEmbedded(embedFS embed.FS, dir string) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	// Read directory entries
	entries, err := fs.ReadDir(embedFS, dir)
	if err != nil {
		return fmt.Errorf("failed to read embedded directory %s: %w", dir, err)
	}

	tsvFiles := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tsv") {
			tsvFiles = append(tsvFiles, filepath.Join(dir, entry.Name()))
		}
	}

	if len(tsvFiles) == 0 {
		return fmt.Errorf("no TSV files found in embedded %s", dir)
	}

	Info("Opening", "Loading opening book from %d embedded files", len(tsvFiles))

	for _, file := range tsvFiles {
		if err := ob.loadEmbeddedTSVFile(embedFS, file); err != nil {
			return fmt.Errorf("failed to load embedded %s: %w", file, err)
		}
	}

	Info("Opening", "Opening book loaded successfully")
	return nil
}

// LoadFromDirectory loads all TSV files from the specified directory (filesystem)
func (ob *OpeningBook) LoadFromDirectory(dir string) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	// Find all .tsv files in the directory
	files, err := filepath.Glob(filepath.Join(dir, "*.tsv"))
	if err != nil {
		return fmt.Errorf("failed to find TSV files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no TSV files found in %s", dir)
	}

	Info("Opening", "Loading opening book from %d files", len(files))

	for _, file := range files {
		if err := ob.loadTSVFile(file); err != nil {
			return fmt.Errorf("failed to load %s: %w", file, err)
		}
	}

	Info("Opening", "Opening book loaded successfully")
	return nil
}

// loadEmbeddedTSVFile loads a single TSV file from embedded filesystem into the trie
func (ob *OpeningBook) loadEmbeddedTSVFile(embedFS embed.FS, filename string) error {
	file, err := embedFS.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return ob.parseTSVReader(file, filename)
}

// loadTSVFile loads a single TSV file into the trie
func (ob *OpeningBook) loadTSVFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return ob.parseTSVReader(file, filename)
}

// parseTSVReader parses TSV content from any reader
func (ob *OpeningBook) parseTSVReader(reader io.Reader, filename string) error {
	scanner := bufio.NewScanner(reader)
	lineNum := 0

	// Skip header line
	if scanner.Scan() {
		lineNum++
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			Warn("Opening", "Skipping malformed line %d in %s", lineNum, filename)
			continue
		}

		eco := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		pgn := strings.TrimSpace(parts[2])

		if err := ob.addOpening(eco, name, pgn); err != nil {
			Warn("Opening", "Failed to add opening at line %d: %v", lineNum, err)
		}
	}

	return scanner.Err()
}

// addOpening adds an opening to the trie
func (ob *OpeningBook) addOpening(eco, name, pgn string) error {
	// Parse the PGN to extract moves
	moves, err := parsePGNMoves(pgn)
	if err != nil {
		return err
	}

	if len(moves) == 0 {
		return fmt.Errorf("no moves in PGN: %s", pgn)
	}

	// Traverse/create the trie path
	current := ob.root
	for _, move := range moves {
		if current.Children == nil {
			current.Children = make(map[string]*OpeningNode)
		}

		if _, exists := current.Children[move]; !exists {
			current.Children[move] = &OpeningNode{
				Move:     move,
				Children: make(map[string]*OpeningNode),
			}
		}
		current = current.Children[move]
	}

	// Store opening info at the terminal node
	// If there's already an opening here, keep the one with the longer name (more specific)
	if current.Opening == nil || len(name) > len(current.Opening.Name) {
		current.Opening = &OpeningInfo{
			ECO:  eco,
			Name: name,
			PGN:  pgn,
		}
	}

	return nil
}

// parsePGNMoves extracts moves from a PGN string
func parsePGNMoves(pgn string) ([]string, error) {
	// Parse the PGN using the chess library
	game := chess.NewGame()

	// Remove move numbers and extra whitespace
	// PGN format: "1. e4 e5 2. Nf3 Nc6"
	parts := strings.Fields(pgn)
	var moves []string

	for _, part := range parts {
		// Skip move numbers (e.g., "1.", "2.")
		if strings.HasSuffix(part, ".") {
			continue
		}

		// Try to parse the move
		if err := game.MoveStr(part); err != nil {
			return nil, fmt.Errorf("invalid move %s in PGN %s: %w", part, pgn, err)
		}

		moves = append(moves, part)
	}

	return moves, nil
}

// Lookup finds the opening for a given sequence of moves
// Returns the deepest matching opening found in the database
// If the move sequence goes beyond the opening book, returns the last known opening
func (ob *OpeningBook) Lookup(moves []string) *OpeningInfo {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	current := ob.root
	var lastOpening *OpeningInfo

	for _, move := range moves {
		if current.Children == nil {
			// We've left the opening book - return the last opening we found
			return lastOpening
		}

		next, exists := current.Children[move]
		if !exists {
			// This move is not in the opening book - return the last opening we found
			return lastOpening
		}

		current = next
		if current.Opening != nil {
			lastOpening = current.Opening
		}
	}

	return lastOpening
}

// LookupByGame finds the opening for a chess game
func (ob *OpeningBook) LookupByGame(game *chess.Game) *OpeningInfo {
	// We need to replay the game to get SAN notation
	// because game.Moves() returns moves in UCI format
	tempGame := chess.NewGame()
	sanMoves := make([]string, 0, len(game.Moves()))

	for _, move := range game.Moves() {
		// Get the SAN notation before making the move
		san := chess.AlgebraicNotation{}.Encode(tempGame.Position(), move)
		sanMoves = append(sanMoves, san)

		// Make the move in the temp game
		tempGame.Move(move)
	}

	return ob.Lookup(sanMoves)
}

// Stats returns statistics about the opening book
func (ob *OpeningBook) Stats() map[string]int {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	stats := map[string]int{
		"total_nodes":    0,
		"total_openings": 0,
		"max_depth":      0,
	}

	var countNodes func(*OpeningNode, int)
	countNodes = func(node *OpeningNode, depth int) {
		if node == nil {
			return
		}

		stats["total_nodes"]++
		if node.Opening != nil {
			stats["total_openings"]++
		}
		if depth > stats["max_depth"] {
			stats["max_depth"] = depth
		}

		for _, child := range node.Children {
			countNodes(child, depth+1)
		}
	}

	countNodes(ob.root, 0)
	return stats
}
