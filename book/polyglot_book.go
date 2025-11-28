package book

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"sync"

	"github.com/notnil/chess"

	"gochess-board/logger"
)

// PolyglotBook represents a Polyglot opening book (.bin file)
type PolyglotBook struct {
	Entries []PolyglotEntry
	mu      sync.RWMutex
}

// PolyglotEntry represents a single entry in the Polyglot book
type PolyglotEntry struct {
	Key    uint64 // Zobrist hash of the position
	Move   uint16 // Move in Polyglot format
	Weight uint16 // Frequency/weight of this move
	Learn  uint32 // Learning data (unused)
}

// NewPolyglotBook creates a new empty Polyglot book
func NewPolyglotBook() *PolyglotBook {
	return &PolyglotBook{
		Entries: make([]PolyglotEntry, 0),
	}
}

// LoadFromFile loads a Polyglot book from a .bin file
func (pb *PolyglotBook) LoadFromFile(filepath string) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open book file: %w", err)
	}
	defer file.Close()

	// Get file size to estimate number of entries
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat book file: %w", err)
	}

	// Each entry is 16 bytes
	entryCount := stat.Size() / 16
	pb.Entries = make([]PolyglotEntry, 0, entryCount)

	// Read all entries
	for {
		var entry PolyglotEntry

		// Read key (8 bytes, big-endian)
		if err := binary.Read(file, binary.BigEndian, &entry.Key); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read entry key: %w", err)
		}

		// Read move (2 bytes, big-endian)
		if err := binary.Read(file, binary.BigEndian, &entry.Move); err != nil {
			return fmt.Errorf("failed to read entry move: %w", err)
		}

		// Read weight (2 bytes, big-endian)
		if err := binary.Read(file, binary.BigEndian, &entry.Weight); err != nil {
			return fmt.Errorf("failed to read entry weight: %w", err)
		}

		// Read learn (4 bytes, big-endian)
		if err := binary.Read(file, binary.BigEndian, &entry.Learn); err != nil {
			return fmt.Errorf("failed to read entry learn: %w", err)
		}

		pb.Entries = append(pb.Entries, entry)
	}

	logger.Info("POLYGLOT_BOOK", "Loaded %d entries from %s", len(pb.Entries), filepath)

	// Sort entries by key for binary search
	sort.Slice(pb.Entries, func(i, j int) bool {
		return pb.Entries[i].Key < pb.Entries[j].Key
	})

	return nil
}

// Probe looks up moves for a given position
func (pb *PolyglotBook) Probe(position *chess.Position) []string {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	// Calculate Zobrist hash for the position
	key := pb.zobristHash(position)

	// Binary search for entries with this key
	matches := []PolyglotEntry{}
	idx := sort.Search(len(pb.Entries), func(i int) bool {
		return pb.Entries[i].Key >= key
	})

	// Collect all entries with matching key
	for idx < len(pb.Entries) && pb.Entries[idx].Key == key {
		matches = append(matches, pb.Entries[idx])
		idx++
	}

	if len(matches) == 0 {
		return nil
	}

	// Convert Polyglot moves to UCI notation
	moves := make([]string, 0, len(matches))
	for _, entry := range matches {
		uciMove := pb.polyglotMoveToUCI(entry.Move)
		if uciMove != "" {
			moves = append(moves, uciMove)
		}
	}

	return moves
}

// ProbeWeighted returns a random move weighted by the book entries
func (pb *PolyglotBook) ProbeWeighted(position *chess.Position) string {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	// Calculate Zobrist hash for the position
	key := pb.zobristHash(position)

	// Binary search for entries with this key
	matches := []PolyglotEntry{}
	idx := sort.Search(len(pb.Entries), func(i int) bool {
		return pb.Entries[i].Key >= key
	})

	// Collect all entries with matching key
	for idx < len(pb.Entries) && pb.Entries[idx].Key == key {
		matches = append(matches, pb.Entries[idx])
		idx++
	}

	if len(matches) == 0 {
		return ""
	}

	// Calculate total weight
	totalWeight := uint64(0)
	for _, entry := range matches {
		totalWeight += uint64(entry.Weight)
	}

	if totalWeight == 0 {
		// If all weights are 0, pick randomly
		entry := matches[rand.Intn(len(matches))]
		return pb.polyglotMoveToUCI(entry.Move)
	}

	// Pick a random move weighted by frequency
	r := rand.Uint64() % totalWeight
	sum := uint64(0)
	for _, entry := range matches {
		sum += uint64(entry.Weight)
		if sum > r {
			return pb.polyglotMoveToUCI(entry.Move)
		}
	}

	// Fallback (shouldn't happen)
	return pb.polyglotMoveToUCI(matches[0].Move)
}

// polyglotMoveToUCI converts a Polyglot move format to UCI notation
func (pb *PolyglotBook) polyglotMoveToUCI(move uint16) string {
	// Polyglot move format (16 bits):
	// bits 0-5: TO square (0=a1, 1=b1, ..., 63=h8)
	// bits 6-11: FROM square
	// bits 12-14: promotion piece (0=none, 1=knight, 2=bishop, 3=rook, 4=queen)
	// Note: This is opposite of what some documentation says!

	to := int(move & 0x3F)
	from := int((move >> 6) & 0x3F)
	promo := int((move >> 12) & 0x7)

	// Convert square index to algebraic notation
	fromSquare := indexToSquare(from)
	toSquare := indexToSquare(to)

	uci := fromSquare + toSquare

	// Add promotion piece if present
	if promo > 0 {
		promoChar := ""
		switch promo {
		case 1:
			promoChar = "n"
		case 2:
			promoChar = "b"
		case 3:
			promoChar = "r"
		case 4:
			promoChar = "q"
		}
		uci += promoChar
	}

	return uci
}

// indexToSquare converts a square index (0-63) to algebraic notation
func indexToSquare(idx int) string {
	file := idx % 8
	rank := idx / 8
	return string(rune('a'+file)) + string(rune('1'+rank))
}

// zobristHash calculates the Polyglot Zobrist hash for a position
func (pb *PolyglotBook) zobristHash(position *chess.Position) uint64 {
	hash := uint64(0)

	// Hash pieces on the board
	board := position.Board()
	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece != chess.NoPiece {
			hash ^= pb.pieceHash(piece, sq)
		}
	}

	// Hash castling rights
	if position.CastleRights().CanCastle(chess.White, chess.KingSide) {
		hash ^= polyglotRandom64[polyglotCastleOffset+0]
	}
	if position.CastleRights().CanCastle(chess.White, chess.QueenSide) {
		hash ^= polyglotRandom64[polyglotCastleOffset+1]
	}
	if position.CastleRights().CanCastle(chess.Black, chess.KingSide) {
		hash ^= polyglotRandom64[polyglotCastleOffset+2]
	}
	if position.CastleRights().CanCastle(chess.Black, chess.QueenSide) {
		hash ^= polyglotRandom64[polyglotCastleOffset+3]
	}

	// Hash en passant square
	if position.EnPassantSquare() != chess.NoSquare {
		file := int(position.EnPassantSquare()) % 8
		hash ^= polyglotRandom64[polyglotEnPassantOffset+file]
	}

	// Hash side to move
	// Note: Polyglot XORs this value when WHITE is to move (not black)
	// This is counterintuitive but verified against python-chess
	if position.Turn() == chess.White {
		hash ^= polyglotRandom64[polyglotBlackToMoveOffset]
	}

	return hash
}

// pieceHash returns the Zobrist hash for a piece at a square
func (pb *PolyglotBook) pieceHash(piece chess.Piece, square int) uint64 {
	// Polyglot piece ordering (alternating black/white):
	// black pawn=0, white pawn=1, black knight=2, white knight=3,
	// black bishop=4, white bishop=5, black rook=6, white rook=7,
	// black queen=8, white queen=9, black king=10, white king=11

	pieceType := piece.Type()
	color := piece.Color()

	// Map Go chess library piece types to Polyglot piece types
	// Go chess: King=1, Queen=2, Rook=3, Bishop=4, Knight=5, Pawn=6
	// Polyglot: Pawn=1, Knight=2, Bishop=3, Rook=4, Queen=5, King=6
	var polyglotPieceType int
	switch pieceType {
	case chess.Pawn:
		polyglotPieceType = 1
	case chess.Knight:
		polyglotPieceType = 2
	case chess.Bishop:
		polyglotPieceType = 3
	case chess.Rook:
		polyglotPieceType = 4
	case chess.Queen:
		polyglotPieceType = 5
	case chess.King:
		polyglotPieceType = 6
	default:
		return 0
	}

	// Calculate index in Polyglot's piece array
	// Formula: (polyglotPieceType-1)*2 + colorOffset
	// where colorOffset is 0 for black, 1 for white
	colorOffset := 0
	if color == chess.White {
		colorOffset = 1
	}
	pieceIdx := (polyglotPieceType-1)*2 + colorOffset

	// Polyglot formula: offset_piece = 64*kind_of_piece + 8*row + file
	// Since 8*row + file == square (for 0-63 indexing), we can simplify:
	offset := 64*pieceIdx + square

	return polyglotRandom64[polyglotPieceOffset+offset]
}

// Hash table offsets in polyglotRandom64 array
const (
	polyglotPieceOffset       = 0   // Pieces: 12 types x 64 squares = 768 values
	polyglotCastleOffset      = 768 // Castle rights: 4 values
	polyglotEnPassantOffset   = 772 // En passant files: 8 values
	polyglotBlackToMoveOffset = 780 // Black to move: 1 value
)
