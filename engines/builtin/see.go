package builtin

import "github.com/notnil/chess"

// SEE (Static Exchange Evaluation) evaluates capture sequences
// Returns the material gain/loss from a capture sequence
// Positive = winning capture, Negative = losing capture, Zero = equal

// pieceValues for SEE (indexed by chess.PieceType)
// chess.PieceType values: NoPieceType=0, King=1, Queen=2, Rook=3, Bishop=4, Knight=5, Pawn=6
var seePieceValues = [7]int{
	0,     // NoPieceType (0)
	20000, // King (1)
	900,   // Queen (2)
	500,   // Rook (3)
	330,   // Bishop (4)
	320,   // Knight (5)
	100,   // Pawn (6)
}

// SEE evaluates a capture move and returns the expected material gain
// This is used to prune losing captures in quiescence search
func (e *InternalEngine) see(pos *chess.Position, move *chess.Move) int {
	// Get the target square and initial captured piece value
	toSq := move.S2()
	fromSq := move.S1()
	board := pos.Board()

	// Get the piece being moved (attacker)
	attacker := board.Piece(fromSq)
	if attacker == chess.NoPiece {
		return 0
	}

	// Get the piece being captured (victim)
	victim := board.Piece(toSq)

	// Handle en passant - captured pawn is not on target square
	if move.HasTag(chess.EnPassant) {
		return seePieceValues[chess.Pawn] // En passant always captures a pawn
	}

	// If no capture, return 0
	if victim == chess.NoPiece {
		// Check for promotion value
		if move.Promo() != chess.NoPieceType {
			return seePieceValues[move.Promo()] - seePieceValues[chess.Pawn]
		}
		return 0
	}

	// Simple SEE: just compare piece values
	// Full SEE would simulate the entire capture sequence
	capturedValue := seePieceValues[victim.Type()]
	attackerValue := seePieceValues[attacker.Type()]

	// If we capture with a less valuable piece, it's likely good
	if attackerValue <= capturedValue {
		return capturedValue
	}

	// For more complex cases, do a simplified exchange evaluation
	// Check if the target square is defended
	return e.seeCapture(pos, toSq, attacker.Color(), capturedValue, attackerValue)
}

// seeCapture performs the actual SEE calculation
// It simulates the capture sequence to determine material outcome
func (e *InternalEngine) seeCapture(pos *chess.Position, targetSq chess.Square, attackerColor chess.Color, capturedValue, attackerValue int) int {
	board := pos.Board()

	// Count attackers and defenders on the target square
	whiteAttackers := e.getAttackers(board, targetSq, chess.White)
	blackAttackers := e.getAttackers(board, targetSq, chess.Black)

	var attackers, defenders []int
	if attackerColor == chess.White {
		attackers = whiteAttackers
		defenders = blackAttackers
	} else {
		attackers = blackAttackers
		defenders = whiteAttackers
	}

	// If no defenders, we win the captured piece
	if len(defenders) == 0 {
		return capturedValue
	}

	// Simulate the exchange
	// gain[0] = initial capture value
	// Each subsequent capture alternates sides
	gain := make([]int, 32)
	depth := 0

	gain[depth] = capturedValue
	pieceOnSquare := attackerValue

	// Remove the initial attacker from the list (it's already captured)
	if len(attackers) > 0 {
		attackers = attackers[1:]
	}

	// Alternate captures until one side runs out
	for {
		depth++

		// Defender captures
		if len(defenders) == 0 {
			break
		}
		defenderValue := defenders[0]
		defenders = defenders[1:]

		gain[depth] = pieceOnSquare - gain[depth-1]
		pieceOnSquare = defenderValue

		// Check for king capture (illegal)
		if defenderValue >= seePieceValues[chess.King] {
			break
		}

		depth++

		// Attacker recaptures
		if len(attackers) == 0 {
			break
		}
		attackerVal := attackers[0]
		attackers = attackers[1:]

		gain[depth] = pieceOnSquare - gain[depth-1]
		pieceOnSquare = attackerVal

		// Check for king capture (illegal)
		if attackerVal >= seePieceValues[chess.King] {
			break
		}
	}

	// Negamax the gain array
	for depth--; depth > 0; depth-- {
		if -gain[depth] < gain[depth-1] {
			gain[depth-1] = -gain[depth]
		}
	}

	return gain[0]
}

// getAttackers returns a sorted list of piece values attacking a square
// Sorted from least valuable to most valuable (for optimal SEE)
func (e *InternalEngine) getAttackers(board *chess.Board, sq chess.Square, color chess.Color) []int {
	var attackers []int
	sqRank := int(sq) / 8
	sqFile := int(sq) % 8

	// Check all squares for pieces that can attack the target
	for fromSq := 0; fromSq < 64; fromSq++ {
		piece := board.Piece(chess.Square(fromSq))
		if piece == chess.NoPiece || piece.Color() != color {
			continue
		}

		fromRank := fromSq / 8
		fromFile := fromSq % 8
		rankDiff := abs(sqRank - fromRank)
		fileDiff := abs(sqFile - fromFile)

		canAttack := false

		switch piece.Type() {
		case chess.Pawn:
			// Pawns attack diagonally forward
			// White pawns attack squares one rank higher
			// Black pawns attack squares one rank lower
			if fileDiff == 1 && rankDiff == 1 {
				if color == chess.White && sqRank > fromRank {
					canAttack = true
				} else if color == chess.Black && sqRank < fromRank {
					canAttack = true
				}
			}

		case chess.Knight:
			// Knight moves in L-shape
			if (rankDiff == 2 && fileDiff == 1) || (rankDiff == 1 && fileDiff == 2) {
				canAttack = true
			}

		case chess.Bishop:
			// Bishop moves diagonally
			if rankDiff == fileDiff && rankDiff > 0 {
				if e.isDiagonalClear(board, chess.Square(fromSq), sq) {
					canAttack = true
				}
			}

		case chess.Rook:
			// Rook moves in straight lines
			if (rankDiff == 0 || fileDiff == 0) && (rankDiff > 0 || fileDiff > 0) {
				if e.isStraightClear(board, chess.Square(fromSq), sq) {
					canAttack = true
				}
			}

		case chess.Queen:
			// Queen moves like bishop or rook
			if rankDiff == fileDiff && rankDiff > 0 {
				if e.isDiagonalClear(board, chess.Square(fromSq), sq) {
					canAttack = true
				}
			} else if (rankDiff == 0 || fileDiff == 0) && (rankDiff > 0 || fileDiff > 0) {
				if e.isStraightClear(board, chess.Square(fromSq), sq) {
					canAttack = true
				}
			}

		case chess.King:
			// King attacks adjacent squares
			if rankDiff <= 1 && fileDiff <= 1 && (rankDiff > 0 || fileDiff > 0) {
				canAttack = true
			}

		default:
			// Unknown piece type - skip
			continue
		}

		if canAttack {
			attackers = append(attackers, seePieceValues[piece.Type()])
		}
	}

	// Sort attackers by value (least valuable first - optimal for SEE)
	sortAttackers(attackers)

	return attackers
}

// isDiagonalClear checks if the diagonal path between two squares is clear
func (e *InternalEngine) isDiagonalClear(board *chess.Board, from, to chess.Square) bool {
	fromRank := int(from) / 8
	fromFile := int(from) % 8
	toRank := int(to) / 8
	toFile := int(to) % 8

	rankDir := 1
	if toRank < fromRank {
		rankDir = -1
	}
	fileDir := 1
	if toFile < fromFile {
		fileDir = -1
	}

	rank := fromRank + rankDir
	file := fromFile + fileDir

	for rank != toRank && file != toFile {
		sq := chess.Square(rank*8 + file)
		if board.Piece(sq) != chess.NoPiece {
			return false
		}
		rank += rankDir
		file += fileDir
	}

	return true
}

// isStraightClear checks if the straight path between two squares is clear
func (e *InternalEngine) isStraightClear(board *chess.Board, from, to chess.Square) bool {
	fromRank := int(from) / 8
	fromFile := int(from) % 8
	toRank := int(to) / 8
	toFile := int(to) % 8

	if fromRank == toRank {
		// Horizontal
		dir := 1
		if toFile < fromFile {
			dir = -1
		}
		for file := fromFile + dir; file != toFile; file += dir {
			sq := chess.Square(fromRank*8 + file)
			if board.Piece(sq) != chess.NoPiece {
				return false
			}
		}
	} else {
		// Vertical
		dir := 1
		if toRank < fromRank {
			dir = -1
		}
		for rank := fromRank + dir; rank != toRank; rank += dir {
			sq := chess.Square(rank*8 + fromFile)
			if board.Piece(sq) != chess.NoPiece {
				return false
			}
		}
	}

	return true
}

// sortAttackers sorts attackers by value (insertion sort - small arrays)
func sortAttackers(attackers []int) {
	for i := 1; i < len(attackers); i++ {
		key := attackers[i]
		j := i - 1
		for j >= 0 && attackers[j] > key {
			attackers[j+1] = attackers[j]
			j--
		}
		attackers[j+1] = key
	}
}

// abs returns absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
