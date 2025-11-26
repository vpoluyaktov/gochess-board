package builtin

import (
	"fmt"
	"math"
	"time"

	"github.com/notnil/chess"
)

// InternalEngine is a simple built-in chess engine
type InternalEngine struct {
	name string
}

// NewEngine creates a new built-in chess engine
func NewEngine() *InternalEngine {
	return &InternalEngine{
		name: "GoChess Basic",
	}
}

// NewInternalEngine is deprecated, use NewEngine instead
func NewInternalEngine() *InternalEngine {
	return NewEngine()
}

// Piece values in centipawns
const (
	pawnValue   = 100
	knightValue = 320
	bishopValue = 330
	rookValue   = 500
	queenValue  = 900
	kingValue   = 20000
)

// Piece-square tables for positional evaluation
// Values are from white's perspective (flip for black)
var pawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTable = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopTable = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0,
}

var queenTable = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingMiddleGameTable = [64]int{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}

// getPieceValue returns the material value of a piece
func getPieceValue(piece chess.Piece) int {
	switch piece.Type() {
	case chess.Pawn:
		return pawnValue
	case chess.Knight:
		return knightValue
	case chess.Bishop:
		return bishopValue
	case chess.Rook:
		return rookValue
	case chess.Queen:
		return queenValue
	case chess.King:
		return kingValue
	default:
		return 0
	}
}

// getPieceSquareValue returns the positional value for a piece on a square
func getPieceSquareValue(piece chess.Piece, sq chess.Square) int {
	// Convert square to array index
	index := int(sq)

	// Flip index for black pieces (they see the board upside down)
	if piece.Color() == chess.Black {
		index = 63 - index
	}

	switch piece.Type() {
	case chess.Pawn:
		return pawnTable[index]
	case chess.Knight:
		return knightTable[index]
	case chess.Bishop:
		return bishopTable[index]
	case chess.Rook:
		return rookTable[index]
	case chess.Queen:
		return queenTable[index]
	case chess.King:
		return kingMiddleGameTable[index]
	default:
		return 0
	}
}

// evaluate returns the evaluation of the position in centipawns
// Positive values favor white, negative values favor black
func (e *InternalEngine) evaluate(pos *chess.Position) int {
	// Check for checkmate or stalemate
	if pos.Status() == chess.Checkmate {
		if pos.Turn() == chess.White {
			return -kingValue // Black wins
		}
		return kingValue // White wins
	}
	if pos.Status() == chess.Stalemate {
		return 0 // Draw
	}

	score := 0
	board := pos.Board()

	// Evaluate all pieces
	for sq := chess.A1; sq <= chess.H8; sq++ {
		piece := board.Piece(sq)
		if piece == chess.NoPiece {
			continue
		}

		// Material value + positional value
		value := getPieceValue(piece) + getPieceSquareValue(piece, sq)

		if piece.Color() == chess.White {
			score += value
		} else {
			score -= value
		}
	}

	// Mobility bonus (number of legal moves)
	moves := pos.ValidMoves()
	mobilityBonus := len(moves) * 10
	if pos.Turn() == chess.White {
		score += mobilityBonus
	} else {
		score -= mobilityBonus
	}

	return score
}

// quiescence performs a quiescence search (only captures)
func (e *InternalEngine) quiescence(pos *chess.Position, alpha, beta int, depth int) int {
	// Limit quiescence depth to prevent infinite loops
	if depth <= 0 {
		return e.evaluate(pos)
	}

	standPat := e.evaluate(pos)

	if standPat >= beta {
		return beta
	}
	if alpha < standPat {
		alpha = standPat
	}

	// Only search captures
	moves := pos.ValidMoves()
	for _, move := range moves {
		// Only consider captures
		if move.HasTag(chess.Capture) || move.HasTag(chess.EnPassant) {
			// Create new position with the move
			newPos := pos.Update(move)
			score := -e.quiescence(newPos, -beta, -alpha, depth-1)

			if score >= beta {
				return beta
			}
			if score > alpha {
				alpha = score
			}
		}
	}

	return alpha
}

// search performs minimax search with alpha-beta pruning
func (e *InternalEngine) search(pos *chess.Position, depth int, alpha, beta int) (int, *chess.Move) {
	// Base case: depth reached
	if depth == 0 {
		score := e.quiescence(pos, alpha, beta, 4)
		return score, nil
	}

	moves := pos.ValidMoves()
	if len(moves) == 0 {
		// Checkmate or stalemate
		return e.evaluate(pos), nil
	}

	// Order moves: captures first, then others
	e.orderMoves(moves)

	var bestMove *chess.Move

	for _, move := range moves {
		// Create new position with the move
		newPos := pos.Update(move)
		score, _ := e.search(newPos, depth-1, -beta, -alpha)
		score = -score

		if score >= beta {
			// Beta cutoff
			return beta, move
		}

		if score > alpha {
			alpha = score
			bestMove = move
		}
	}

	return alpha, bestMove
}

// orderMoves orders moves to improve alpha-beta pruning
// Captures and checks are searched first
func (e *InternalEngine) orderMoves(moves []*chess.Move) {
	// Simple ordering: captures first
	// This is a basic implementation - could be improved with MVV-LVA, killer moves, etc.
	for i := 0; i < len(moves); i++ {
		for j := i + 1; j < len(moves); j++ {
			scoreI := e.moveOrderScore(moves[i])
			scoreJ := e.moveOrderScore(moves[j])
			if scoreJ > scoreI {
				moves[i], moves[j] = moves[j], moves[i]
			}
		}
	}
}

// moveOrderScore assigns a score for move ordering
func (e *InternalEngine) moveOrderScore(move *chess.Move) int {
	score := 0

	// Prioritize captures
	if move.HasTag(chess.Capture) {
		score += 1000
	}

	// Prioritize checks
	if move.HasTag(chess.Check) {
		score += 500
	}

	// Prioritize promotions
	if move.Promo() != chess.NoPieceType {
		score += 800
	}

	return score
}

// GetBestMove implements the ChessEngine interface
func (e *InternalEngine) GetBestMove(fen string, moveTime time.Duration) (string, error) {
	// Parse FEN
	fenFunc, err := chess.FEN(fen)
	if err != nil {
		return "", fmt.Errorf("invalid FEN: %w", err)
	}

	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Use iterative deepening with time limit
	var bestMove *chess.Move
	startTime := time.Now()

	// Search depths 1-5 (adjust based on time available)
	maxDepth := 4
	if moveTime > 5*time.Second {
		maxDepth = 5
	} else if moveTime < 1*time.Second {
		maxDepth = 3
	}

	for depth := 1; depth <= maxDepth; depth++ {
		// Check if we're running out of time
		if time.Since(startTime) > moveTime*8/10 {
			break
		}

		_, move := e.search(pos, depth, -math.MaxInt32, math.MaxInt32)
		if move != nil {
			bestMove = move
		}
	}

	if bestMove == nil {
		// Fallback: return first legal move
		moves := pos.ValidMoves()
		if len(moves) > 0 {
			bestMove = moves[0]
		} else {
			return "", fmt.Errorf("no legal moves available")
		}
	}

	return bestMove.String(), nil
}

// GetBestMoveWithClock implements the ChessEngine interface
func (e *InternalEngine) GetBestMoveWithClock(fen string, moveHistory []string, whiteTime, blackTime, whiteInc, blackInc time.Duration) (string, error) {
	// Simple time management: use a fraction of remaining time
	var timeToUse time.Duration

	fenFunc, err := chess.FEN(fen)
	if err != nil {
		return "", fmt.Errorf("invalid FEN: %w", err)
	}
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	if pos.Turn() == chess.White {
		timeToUse = whiteTime / 20
		if whiteInc > 0 {
			timeToUse += whiteInc / 2
		}
	} else {
		timeToUse = blackTime / 20
		if blackInc > 0 {
			timeToUse += blackInc / 2
		}
	}

	// Minimum 100ms, maximum 10s
	if timeToUse < 100*time.Millisecond {
		timeToUse = 100 * time.Millisecond
	}
	if timeToUse > 10*time.Second {
		timeToUse = 10 * time.Second
	}

	return e.GetBestMove(fen, timeToUse)
}

// SetOption implements the ChessEngine interface
func (e *InternalEngine) SetOption(name, value string) error {
	// Internal engine doesn't have configurable options yet
	return nil
}

// Close implements the ChessEngine interface
func (e *InternalEngine) Close() error {
	// Nothing to clean up
	return nil
}
