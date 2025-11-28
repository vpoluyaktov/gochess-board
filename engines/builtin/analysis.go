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
	PV        []string // Principal Variation (sequence of best moves)
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

	// Increment TT age for new analysis session
	if e.tt != nil {
		e.tt.incrementAge()
	}

	// Perform iterative deepening
	for depth := 1; depth <= maxDepth; depth++ {
		// Check if we should stop
		select {
		case <-stopCh:
			return nil
		default:
		}

		startTime := time.Now()
		score, bestMove, pv, nodes := e.searchWithStats(pos, depth, stopCh)
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

		// Convert PV moves to strings
		pvStrings := make([]string, len(pv))
		for i, move := range pv {
			pvStrings[i] = move.String()
		}

		info := AnalysisInfo{
			Depth:     depth,
			Score:     displayScore,
			BestMove:  bestMoveStr,
			PV:        pvStrings,
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

// searchWithStats performs a search and returns score, best move, PV, and node count
func (e *InternalEngine) searchWithStats(pos *chess.Position, depth int, stopCh <-chan bool) (int, *chess.Move, []*chess.Move, int) {
	nodeCount := 0
	alpha := -10000
	beta := 10000
	alphaOrig := alpha

	// Get zobrist key for transposition table
	zobristKey := getZobristKey(pos)

	// Probe transposition table
	if found, score, ttMove := e.tt.probe(zobristKey, depth, alpha, beta); found {
		// For analysis, we still want the PV, so only use TT for move ordering
		// Don't return early to ensure we get complete PV
		_ = score // Use score for move ordering hint
		_ = ttMove
	}

	moves := pos.ValidMoves()
	if len(moves) == 0 {
		return 0, nil, nil, 0
	}

	var bestMove *chess.Move
	var bestPV []*chess.Move
	bestScore := -10000

	// Get TT move for ordering
	var ttMove *chess.Move
	if _, _, move := e.tt.probe(zobristKey, 0, alpha, beta); move != nil {
		ttMove = move
	}

	// Order moves for better pruning (with TT move and killer moves)
	e.orderMoves(moves, ttMove, 0)

	for _, move := range moves {
		// Check if we should stop
		select {
		case <-stopCh:
			return bestScore, bestMove, bestPV, nodeCount
		default:
		}

		newPos := pos.Update(move)

		// Create PV table for this branch
		var childPV []*chess.Move
		score := -e.alphaBetaWithPV(newPos, depth-1, -beta, -alpha, &nodeCount, &childPV, stopCh)

		if score > bestScore {
			bestScore = score
			bestMove = move

			// Build PV: current move + child PV
			bestPV = make([]*chess.Move, 0, depth)
			bestPV = append(bestPV, move)
			bestPV = append(bestPV, childPV...)
		}

		if score > alpha {
			alpha = score
		}

		if alpha >= beta {
			// Store killer move
			if e.killerMoves != nil && !move.HasTag(chess.Capture) {
				e.killerMoves.add(move, depth)
			}
			break
		}
	}

	// Store in transposition table
	if bestMove != nil {
		entryType := TTExact
		if alpha <= alphaOrig {
			entryType = TTAlpha
		} else if alpha >= beta {
			entryType = TTBeta
		}
		e.tt.store(zobristKey, depth, bestScore, entryType, bestMove)
	}

	return bestScore, bestMove, bestPV, nodeCount
}

// alphaBetaWithPV performs alpha-beta search with PV tracking
func (e *InternalEngine) alphaBetaWithPV(pos *chess.Position, depth int, alpha, beta int, nodeCount *int, pv *[]*chess.Move, stopCh <-chan bool) int {
	*nodeCount++
	alphaOrig := alpha

	// Check if we should stop
	select {
	case <-stopCh:
		return 0
	default:
	}

	// Get zobrist key for transposition table
	zobristKey := getZobristKey(pos)

	// Probe transposition table
	if found, score, _ := e.tt.probe(zobristKey, depth, alpha, beta); found {
		// For PV nodes, we need to search to get the full PV
		// So we only use TT cutoffs for non-PV nodes
		if depth > 1 {
			return score
		}
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

	// Get TT move for ordering
	var ttMove *chess.Move
	if _, _, move := e.tt.probe(zobristKey, 0, alpha, beta); move != nil {
		ttMove = move
	}

	// Order moves for better pruning (with TT move and killer moves)
	ply := 6 - depth // Approximate ply from depth
	if ply < 0 {
		ply = 0
	}
	e.orderMoves(moves, ttMove, ply)

	var localPV []*chess.Move
	var bestMove *chess.Move

	for _, move := range moves {
		// Check if we should stop
		select {
		case <-stopCh:
			return alpha
		default:
		}

		newPos := pos.Update(move)

		var childPV []*chess.Move
		score := -e.alphaBetaWithPV(newPos, depth-1, -beta, -alpha, nodeCount, &childPV, stopCh)

		if score >= beta {
			// Beta cutoff - store killer move
			if e.killerMoves != nil && !move.HasTag(chess.Capture) {
				e.killerMoves.add(move, ply)
			}
			// Store in TT
			e.tt.store(zobristKey, depth, beta, TTBeta, move)
			return beta
		}

		if score > alpha {
			alpha = score
			bestMove = move

			// Update PV: current move + child PV
			localPV = make([]*chess.Move, 0, depth)
			localPV = append(localPV, move)
			localPV = append(localPV, childPV...)
		}
	}

	// Store in transposition table
	entryType := TTExact
	if alpha <= alphaOrig {
		entryType = TTAlpha
	}
	if bestMove != nil {
		e.tt.store(zobristKey, depth, alpha, entryType, bestMove)
	}

	// Copy local PV to output
	*pv = localPV
	return alpha
}

// alphaBetaWithStop performs alpha-beta search with stop channel support
func (e *InternalEngine) alphaBetaWithStop(pos *chess.Position, depth int, alpha, beta int, nodeCount *int, stopCh <-chan bool) int {
	*nodeCount++
	alphaOrig := alpha

	// Check if we should stop
	select {
	case <-stopCh:
		return 0
	default:
	}

	// Get zobrist key for transposition table
	zobristKey := getZobristKey(pos)

	// Probe transposition table
	if found, score, _ := e.tt.probe(zobristKey, depth, alpha, beta); found {
		return score
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

	// Get TT move for ordering
	var ttMove *chess.Move
	if _, _, move := e.tt.probe(zobristKey, 0, alpha, beta); move != nil {
		ttMove = move
	}

	// Order moves for better pruning (with TT move and killer moves)
	ply := 6 - depth // Approximate ply from depth
	if ply < 0 {
		ply = 0
	}
	e.orderMoves(moves, ttMove, ply)

	var bestMove *chess.Move

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
			// Beta cutoff - store killer move
			if e.killerMoves != nil && !move.HasTag(chess.Capture) {
				e.killerMoves.add(move, ply)
			}
			// Store in TT
			e.tt.store(zobristKey, depth, beta, TTBeta, move)
			return beta
		}
		if score > alpha {
			alpha = score
			bestMove = move
		}
	}

	// Store in transposition table
	entryType := TTExact
	if alpha <= alphaOrig {
		entryType = TTAlpha
	}
	if bestMove != nil {
		e.tt.store(zobristKey, depth, alpha, entryType, bestMove)
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
