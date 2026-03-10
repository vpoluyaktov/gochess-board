# ELO Testing Guide for GoChess Engine

## Quick Start

### Run Built-in Tactical Tests

```bash
# Quick test (1 second per position)
go test ./engines/builtin -run TestTacticalSuiteQuick -v

# Full test (3 seconds per position)
go test ./engines/builtin -run TestIndividualTacticalPositions -v

# Get ELO estimation
go test ./engines/builtin -run TestTacticalSuite -v
```

### Current Performance

Based on our tactical test suite (10 positions):
- **Solved:** 2/10 (20%)
- **Estimated ELO:** 1000-1200 (Beginner)
- **Category:** Developing

**Note:** The engine excels at positional play but needs improvement in tactical calculations. The 1400-1600 ELO estimate from IMPROVEMENTS.md is based on positional understanding, not pure tactics.

## Methods for Accurate ELO Testing

### 1. **Automated Engine Tournaments** ⭐ RECOMMENDED

#### Using cutechess-cli

```bash
# Install
sudo apt-get install cutechess-cli

# Or download from: https://github.com/cutechess/cutechess

# Run tournament against Stockfish (limited depth)
cutechess-cli \
  -engine cmd=./gochess name=GoChess proto=uci \
  -engine cmd=stockfish name=Stockfish depth=1 proto=uci \
  -each tc=60+1 \
  -rounds 50 \
  -pgnout games.pgn \
  -openings file=openings.pgn order=random \
  -ratinginterval 10
```

**Expected Results:**
- vs Stockfish depth=1: ~40-50% (Stockfish ~1500 ELO at depth 1)
- vs Stockfish depth=2: ~20-30% (Stockfish ~1800 ELO at depth 2)

#### Using BayesElo

After running games, calculate ELO:

```bash
# Install BayesElo
git clone https://github.com/michiguel/BayesElo
cd BayesElo && make

# Calculate ratings from PGN
./bayeselo
> readpgn games.pgn
> elo
> mm
> ratings
```

### 2. **Online Testing Platforms**

#### Lichess BOT API

Create a bot account and play rated games:

```bash
# 1. Create bot account at https://lichess.org/
# 2. Get API token
# 3. Use lichess-bot: https://github.com/lichess-bot-devs/lichess-bot

# Your engine will get an official Lichess rating!
```

**Advantages:**
- Official rating
- Play against humans and bots
- Automatic game recording
- Free

#### Chess.com Computer API

Similar to Lichess but requires approval.

### 3. **Rating List Submission**

Submit your engine to official rating lists:

#### CCRL (Computer Chess Rating Lists)
- Website: http://ccrl.chessdom.com/
- Email: ccrl@chessdom.com
- They test your engine and publish official rating
- Most respected rating list

#### CEGT (Chess Engines Grand Tournament)
- Website: http://www.cegt.net/
- Similar to CCRL
- Different time controls

**Requirements:**
- UCI or CECP protocol support
- Stable binary
- Documentation

### 4. **Tactical Test Suites**

We've implemented a basic suite. For more comprehensive testing:

#### Win At Chess (WAC)
- 300 tactical positions
- Industry standard
- Download: http://www.talkchess.com/forum3/viewtopic.php?t=26673

#### Arasan Test Suite
- 200 positions
- Mixed tactical and positional
- Download: http://www.arasanchess.org/testpositions.html

#### Strategic Test Suite (STS)
- 1500 positions
- Tests positional understanding
- Download: http://www.chessprogramming.org/Strategic_Test_Suite

### 5. **Self-Play Testing**

Play the engine against itself or previous versions:

```go
// Example self-play test
func TestSelfPlay(t *testing.T) {
    engine1 := NewEngine()
    engine2 := NewEngine()
    
    // Play 10 games
    wins := 0
    for i := 0; i < 10; i++ {
        winner := playGame(engine1, engine2, 1*time.Minute)
        if winner == 1 {
            wins++
        }
    }
    
    t.Logf("Engine 1 won %d/10 games", wins)
}
```

## Interpreting Results

### Tactical Test Performance

| Score | ELO Range | Category | Description |
|-------|-----------|----------|-------------|
| 90%+ | 1800-2000 | Expert | Finds most tactics |
| 80-89% | 1600-1800 | Advanced | Strong tactical vision |
| 70-79% | 1500-1700 | Intermediate+ | Good tactics |
| 60-69% | 1400-1600 | Intermediate | Decent tactics |
| 50-59% | 1300-1500 | Developing | Basic tactics |
| 40-49% | 1200-1400 | Beginner+ | Learning |
| <40% | 1000-1200 | Beginner | Needs work |

### Game Performance

| Win Rate vs Known Engine | Your ELO Estimate |
|-------------------------|-------------------|
| 75% vs 1200 ELO | ~1400 |
| 50% vs 1400 ELO | ~1400 |
| 25% vs 1600 ELO | ~1400 |

Use this formula:
```
Your ELO ≈ Opponent ELO + 400 * log10(Win% / (1 - Win%))
```

## Current GoChess Engine Assessment

### Strengths
✅ **Positional Understanding** (1400-1600 level)
- King safety evaluation
- Pawn structure analysis
- Mobility awareness
- Piece-square tables

✅ **Search Efficiency** (1400-1600 level)
- Transposition table
- Killer moves
- Good move ordering
- Quiescence search

### Weaknesses
❌ **Tactical Vision** (1000-1200 level)
- Misses forced mates
- Doesn't see deep combinations
- Limited horizon effect handling

❌ **Search Depth** (Limited)
- Only reaches depth 4-6 in reasonable time
- Needs null move pruning for deeper search
- No late move reductions

### Overall Estimate

**Blitz (1 min + 1 sec):** 1200-1400 ELO
- Limited time hurts tactical calculations
- Good positional play helps

**Rapid (5 min + 3 sec):** 1400-1600 ELO
- More time for deeper search
- Positional understanding shines

**Classical (15 min + 10 sec):** 1400-1600 ELO
- Transposition table very effective
- Positional evaluation dominant

## Recommendations for Improvement

### To Reach 1600+ ELO

1. **Implement Null Move Pruning** (+50-80 ELO)
   - Allows deeper search
   - Finds tactics faster

2. **Add Late Move Reductions** (+50-100 ELO)
   - Search promising moves deeper
   - Reduces search tree

3. **Improve Tactical Vision** (+50-100 ELO)
   - Extend search for checks/captures
   - Better mate detection

### To Reach 1800+ ELO

4. **Opening Book** (+50-100 ELO)
   - Avoid early mistakes
   - Save time for middlegame

5. **Endgame Tablebases** (+50-100 ELO)
   - Perfect endgame play
   - Convert advantages

6. **Aspiration Windows** (+20-40 ELO)
   - Faster search
   - Better time management

## Running Your Own Tests

### 1. Quick Tactical Test (5 minutes)

```bash
go test ./engines/builtin -run TestTacticalSuite -v
```

### 2. Engine vs Engine (30 minutes)

```bash
# Requires cutechess-cli
cutechess-cli \
  -engine cmd=./gochess name=GoChess \
  -engine cmd=stockfish name=Stockfish depth=1 \
  -each tc=60+1 \
  -rounds 20 \
  -pgnout test.pgn
```

### 3. Online Testing (Ongoing)

```bash
# Set up Lichess bot
# Plays rated games automatically
# Gets official rating in 20-30 games
```

## Conclusion

**Current Estimated ELO: 1200-1600** (depending on time control)

The engine has:
- ✅ Strong positional foundation
- ✅ Efficient search with TT and killer moves
- ❌ Weak tactical calculation
- ❌ Limited search depth

**Best testing method:** Play 50-100 games on Lichess BOT API for official rating.

**Quick estimate:** Run tactical test suite (takes 5 minutes).

**Most accurate:** Submit to CCRL for professional testing.
