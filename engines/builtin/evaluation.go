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

	// Add king safety evaluation
	score += evaluateKingSafety(pos, chess.White)
	score -= evaluateKingSafety(pos, chess.Black)

	// Add pawn structure evaluation
	score += evaluatePawnStructure(pos, chess.White)
	score -= evaluatePawnStructure(pos, chess.Black)

	// Add mobility bonus
	score += evaluateMobility(pos, chess.White)
	score -= evaluateMobility(pos, chess.Black)

	// Return score from white's perspective
	return score
}

// evaluateKingSafety evaluates king safety for a given color
func evaluateKingSafety(pos *chess.Position, color chess.Color) int {
	score := 0
	board := pos.Board()

	// Find king position
	var kingSquare chess.Square
	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece.Type() == chess.King && piece.Color() == color {
			kingSquare = chess.Square(sq)
			break
		}
	}

	// Pawn shield bonus (pawns in front of king)
	rank := int(kingSquare) / 8
	file := int(kingSquare) % 8

	if color == chess.White && rank < 2 {
		// Check for pawns in front of king
		for f := max(0, file-1); f <= min(7, file+1); f++ {
			sq := chess.Square(rank*8 + f + 8)
			if sq < 64 {
				piece := board.Piece(sq)
				if piece.Type() == chess.Pawn && piece.Color() == color {
					score += 10 // Bonus for pawn shield
				}
			}
		}
	} else if color == chess.Black && rank > 5 {
		// Check for pawns in front of black king
		for f := max(0, file-1); f <= min(7, file+1); f++ {
			sq := chess.Square(rank*8 + f - 8)
			if sq >= 0 {
				piece := board.Piece(sq)
				if piece.Type() == chess.Pawn && piece.Color() == color {
					score += 10 // Bonus for pawn shield
				}
			}
		}
	}

	return score
}

// evaluatePawnStructure evaluates pawn structure for a given color
func evaluatePawnStructure(pos *chess.Position, color chess.Color) int {
	score := 0
	board := pos.Board()

	// Track pawns by file
	var pawnFiles [8]int

	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece.Type() == chess.Pawn && piece.Color() == color {
			file := int(sq) % 8
			pawnFiles[file]++
		}
	}

	// Penalize doubled pawns
	for file := 0; file < 8; file++ {
		if pawnFiles[file] > 1 {
			score -= 10 * (pawnFiles[file] - 1)
		}
	}

	// Bonus for passed pawns (simplified - just check if no enemy pawns on file)
	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece.Type() == chess.Pawn && piece.Color() == color {
			file := int(sq) % 8
			rank := int(sq) / 8

			// Check if passed pawn (no enemy pawns ahead)
			isPassed := true
			enemyColor := chess.Black
			if color == chess.Black {
				enemyColor = chess.White
			}

			// Check file and adjacent files
			for checkFile := max(0, file-1); checkFile <= min(7, file+1); checkFile++ {
				for checkRank := 0; checkRank < 8; checkRank++ {
					if color == chess.White && checkRank <= rank {
						continue
					}
					if color == chess.Black && checkRank >= rank {
						continue
					}

					checkSq := chess.Square(checkRank*8 + checkFile)
					checkPiece := board.Piece(checkSq)
					if checkPiece.Type() == chess.Pawn && checkPiece.Color() == enemyColor {
						isPassed = false
						break
					}
				}
				if !isPassed {
					break
				}
			}

			if isPassed {
				// Bonus increases as pawn advances
				advancement := rank
				if color == chess.Black {
					advancement = 7 - rank
				}
				score += 10 + advancement*5
			}
		}
	}

	return score
}

// evaluateMobility evaluates piece mobility for a given color
func evaluateMobility(pos *chess.Position, color chess.Color) int {
	// Simple mobility: count legal moves when it's this color's turn
	// Only do this if it's actually their turn to avoid expensive calculation
	if pos.Turn() != color {
		return 0
	}

	moves := pos.ValidMoves()
	// Small bonus for having more moves (mobility)
	return len(moves) / 2
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
