package builtin

import "github.com/notnil/chess"

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
	sqIndex := int(sq)

	// Flip the square for black pieces (mirror vertically)
	if piece.Color() == chess.Black {
		sqIndex = sqIndex ^ 56 // XOR with 56 flips the rank
	}

	switch piece.Type() {
	case chess.Pawn:
		return pawnTable[sqIndex]
	case chess.Knight:
		return knightTable[sqIndex]
	case chess.Bishop:
		return bishopTable[sqIndex]
	case chess.Rook:
		return rookTable[sqIndex]
	case chess.Queen:
		return queenTable[sqIndex]
	case chess.King:
		return kingMiddleGameTable[sqIndex]
	default:
		return 0
	}
}

// evaluate returns the evaluation of the position in centipawns
// Positive values favor white, negative values favor black
func (e *InternalEngine) evaluate(pos *chess.Position) int {
	// Check for checkmate or stalemate
	if pos.Status() == chess.Checkmate {
		// If it's checkmate, the side to move has lost
		if pos.Turn() == chess.White {
			return -10000 // Black wins
		}
		return 10000 // White wins
	}

	if pos.Status() == chess.Stalemate || pos.Status() == chess.ThreefoldRepetition {
		return 0 // Draw
	}

	score := 0
	board := pos.Board()

	// Evaluate material and position
	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece != chess.NoPiece {
			materialValue := getPieceValue(piece)
			positionalValue := getPieceSquareValue(piece, chess.Square(sq))

			if piece.Color() == chess.White {
				score += materialValue + positionalValue
			} else {
				score -= materialValue + positionalValue
			}
		}
	}

	// Return score from white's perspective
	return score
}
