package builtin

import (
	"fmt"
	"time"

	"github.com/notnil/chess"
)

// AnalysisInfo represents analysis data from the engine
type AnalysisInfo struct {
	Depth     int
	Score     int
	BestMove  string
	Nodes     int
	NPS       int64
	Time      int
	ScoreType string // "cp" or "mate"
}

// Analyze performs iterative deepening analysis on a position
// Sends analysis updates through the provided channel
// Returns when maxDepth is reached or stopCh receives a signal
func (e *InternalEngine) Analyze(fen string, maxDepth int, stopCh <-chan bool, resultCh chan<- AnalysisInfo) error {
	// Parse FEN
	fenFunc, err := chess.FEN(fen)
	if err != nil {
		return fmt.Errorf("invalid FEN: %v", err)
	}

	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Perform iterative deepening
	for depth := 1; depth <= maxDepth; depth++ {
		// Check if we should stop
		select {
		case <-stopCh:
			return nil
		default:
		}

		startTime := time.Now()
		score, bestMove, nodes := e.searchWithStats(pos, depth, stopCh)
		elapsed := time.Since(startTime)

		// Check if search was interrupted
		select {
		case <-stopCh:
			return nil
		default:
		}

		elapsedMs := int(elapsed.Milliseconds())
		if elapsedMs == 0 {
			elapsedMs = 1
		}

		nps := int64(nodes) * 1000 / int64(elapsedMs)

		// Determine score type
		scoreType := "cp"
		displayScore := score

		if score > 9000 {
			scoreType = "mate"
			displayScore = (10000 - score + 1) / 2
		} else if score < -9000 {
			scoreType = "mate"
			displayScore = -(10000 + score + 1) / 2
		}

		bestMoveStr := ""
		if bestMove != nil {
			bestMoveStr = bestMove.String()
		}

		info := AnalysisInfo{
			Depth:     depth,
			Score:     displayScore,
			BestMove:  bestMoveStr,
			Nodes:     nodes,
			NPS:       nps,
			Time:      elapsedMs,
			ScoreType: scoreType,
		}

		// Send analysis info
		select {
		case resultCh <- info:
		case <-stopCh:
			return nil
		}
	}

	return nil
}

// searchWithStats performs a search and returns score, best move, and node count
func (e *InternalEngine) searchWithStats(pos *chess.Position, depth int, stopCh <-chan bool) (int, *chess.Move, int) {
	nodeCount := 0
	alpha := -10000
	beta := 10000

	moves := pos.ValidMoves()
	if len(moves) == 0 {
		return 0, nil, 0
	}

	var bestMove *chess.Move
	bestScore := -10000

	// Order moves for better pruning
	e.orderMoves(moves)

	for _, move := range moves {
		// Check if we should stop
		select {
		case <-stopCh:
			return bestScore, bestMove, nodeCount
		default:
		}

		newPos := pos.Update(move)
		score := -e.alphaBetaWithStop(newPos, depth-1, -beta, -alpha, &nodeCount, stopCh)

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score > alpha {
			alpha = score
		}

		if alpha >= beta {
			break
		}
	}

	return bestScore, bestMove, nodeCount
}

// alphaBetaWithStop performs alpha-beta search with stop channel support
func (e *InternalEngine) alphaBetaWithStop(pos *chess.Position, depth int, alpha, beta int, nodeCount *int, stopCh <-chan bool) int {
	*nodeCount++

	// Check if we should stop
	select {
	case <-stopCh:
		return 0
	default:
	}

	// Base case
	if depth == 0 {
		return e.quiescenceWithStop(pos, alpha, beta, 4, nodeCount, stopCh)
	}

	// Check for terminal positions
	if pos.Status() == chess.Checkmate {
		return -10000 + (6 - depth)
	}
	if pos.Status() == chess.Stalemate || pos.Status() == chess.ThreefoldRepetition {
		return 0
	}

	moves := pos.ValidMoves()
	if len(moves) == 0 {
		return e.evaluate(pos)
	}

	// Order moves for better pruning
	e.orderMoves(moves)

	for _, move := range moves {
		// Check if we should stop
		select {
		case <-stopCh:
			return alpha
		default:
		}

		newPos := pos.Update(move)
		score := -e.alphaBetaWithStop(newPos, depth-1, -beta, -alpha, nodeCount, stopCh)

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

// quiescenceWithStop performs quiescence search with stop channel support
func (e *InternalEngine) quiescenceWithStop(pos *chess.Position, alpha, beta int, depth int, nodeCount *int, stopCh <-chan bool) int {
	*nodeCount++

	// Check if we should stop
	select {
	case <-stopCh:
		return 0
	default:
	}

	standPat := e.evaluate(pos)

	if depth <= 0 {
		return standPat
	}

	if standPat >= beta {
		return beta
	}
	if standPat > alpha {
		alpha = standPat
	}

	moves := pos.ValidMoves()
	for _, move := range moves {
		// Only search captures in quiescence
		if move.HasTag(chess.Capture) || move.HasTag(chess.EnPassant) {
			// Check if we should stop
			select {
			case <-stopCh:
				return alpha
			default:
			}

			newPos := pos.Update(move)
			score := -e.quiescenceWithStop(newPos, -beta, -alpha, depth-1, nodeCount, stopCh)

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
