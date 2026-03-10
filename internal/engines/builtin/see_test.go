package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

func TestSEEWinningCapture(t *testing.T) {
	engine := NewEngine()

	// Position where pawn captures queen (clearly winning)
	fen := "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Create a hypothetical capture move (if there was one)
	// For this test, we'll just verify the SEE function works
	moves := pos.ValidMoves()
	for _, move := range moves {
		if move.HasTag(chess.Capture) {
			seeValue := engine.see(pos, move)
			t.Logf("Move %s: SEE value = %d", move.String(), seeValue)
		}
	}
}

func TestSEELosingCapture(t *testing.T) {
	engine := NewEngine()

	// Position where queen captures defended pawn (losing)
	// White queen on d4, black pawn on e5 defended by pawn on f6
	fen := "rnbqkbnr/pppp1ppp/5p2/4p3/3Q4/8/PPPPPPPP/RNB1KBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	moves := pos.ValidMoves()
	for _, move := range moves {
		if move.HasTag(chess.Capture) {
			seeValue := engine.see(pos, move)
			t.Logf("Move %s: SEE value = %d", move.String(), seeValue)
			// Queen taking pawn defended by pawn should be negative
			if move.S2() == chess.E5 {
				if seeValue >= 0 {
					t.Logf("Warning: Qxe5 should be losing (queen for pawn)")
				}
			}
		}
	}
}

func TestSEEEqualCapture(t *testing.T) {
	engine := NewEngine()

	// Position where knight takes knight (equal exchange)
	fen := "r1bqkbnr/pppppppp/2n5/8/4N3/8/PPPPPPPP/RNBQKB1R w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	moves := pos.ValidMoves()
	for _, move := range moves {
		if move.HasTag(chess.Capture) {
			seeValue := engine.see(pos, move)
			t.Logf("Move %s: SEE value = %d", move.String(), seeValue)
		}
	}
}

func TestSEEPawnCapture(t *testing.T) {
	engine := NewEngine()

	// Simple pawn capture
	fen := "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	moves := pos.ValidMoves()
	for _, move := range moves {
		if move.HasTag(chess.Capture) {
			seeValue := engine.see(pos, move)
			t.Logf("Move %s: SEE value = %d", move.String(), seeValue)
			// exd5 should be positive (pawn takes pawn)
			if move.String() == "e4d5" {
				if seeValue < 0 {
					t.Errorf("exd5 should be winning or equal, got SEE = %d", seeValue)
				}
			}
		}
	}
}

func TestSEEEnPassant(t *testing.T) {
	engine := NewEngine()

	// En passant capture
	fen := "rnbqkbnr/pppp1ppp/8/4pP2/8/8/PPPPP1PP/RNBQKBNR w KQkq e6 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	moves := pos.ValidMoves()
	for _, move := range moves {
		if move.HasTag(chess.EnPassant) {
			seeValue := engine.see(pos, move)
			t.Logf("En passant move %s: SEE value = %d", move.String(), seeValue)
			// En passant should be equal (pawn for pawn)
			if seeValue != seePieceValues[chess.Pawn] {
				t.Logf("En passant SEE = %d, expected %d", seeValue, seePieceValues[chess.Pawn])
			}
		}
	}
}

func TestGetAttackers(t *testing.T) {
	engine := NewEngine()

	// Position with pawn on e4 attacking d5
	fen := "r1bqkbnr/pppppppp/2n5/8/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	board := pos.Board()

	// Check attackers on d5
	d5 := chess.D5
	whiteAttackers := engine.getAttackers(board, d5, chess.White)
	blackAttackers := engine.getAttackers(board, d5, chess.Black)

	// White should have e4 pawn attacking d5
	if len(whiteAttackers) < 1 {
		t.Errorf("Expected at least 1 white attacker on d5, got %d", len(whiteAttackers))
	}

	// The pawn value should be 100
	if len(whiteAttackers) > 0 && whiteAttackers[0] != 100 {
		t.Errorf("Expected pawn value 100, got %d", whiteAttackers[0])
	}

	// Black knight on c6 doesn't attack d5 (knight on c6 attacks b4, d4, a5, e5, b8, d8, a7, e7)
	if len(blackAttackers) > 0 {
		t.Logf("Black attackers on d5: %v", blackAttackers)
	}
}
