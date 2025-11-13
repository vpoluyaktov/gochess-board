# Chess Clock Feature Implementation

## Overview
Implemented a comprehensive chess clock system with standard time controls (e.g., 5+5, 10+0, etc.) that tracks time for each player (human or engine).

## Features Implemented

### 1. Time Control Options
Standard chess time controls available:
- **Bullet**: 1+0, 1+1, 2+1
- **Blitz**: 3+0, 3+2, 5+0, 5+3, 5+5 (default)
- **Rapid**: 10+0, 10+5, 15+10
- **Classical**: 30+0, 30+20, 60+0, 90+30
- **Unlimited**: 0+0 (no time limit)

Format: `Initial Minutes + Increment Seconds`
- Initial time: Starting time for each player
- Increment: Time added after each move

### 2. Visual Chess Clocks
- **Dual clock display**: Separate clocks for White and Black
- **Active player highlighting**: Current player's clock is highlighted
- **Time warnings**:
  - Orange border when < 60 seconds
  - Red border + pulsing animation when < 10 seconds
- **Time formatting**: MM:SS or H:MM:SS for longer games
- **Monospace font**: Clear, easy-to-read digital clock display

### 3. Clock Behavior
- **Auto-start**: Clock starts on the first move
- **Increment system**: Time added after each move completes
- **Pause on game end**: Clock stops when game is over
- **Time-out detection**: Automatic loss when time runs out
- **Persistent settings**: Time control preference saved in localStorage

### 4. Backend Implementation

#### New API Endpoints
- `POST /api/clock/set` - Set time control (initial + increment)
- `GET /api/clock/get` - Get current clock state
- `POST /api/clock/start` - Start the chess clock

#### GameState Enhancements
- `TimeControl` struct with InitialTime and Increment
- `WhiteTimeLeft` and `BlackTimeLeft` tracking
- `ClockRunning` state management
- Time deduction and increment logic integrated with move handling

### 5. Frontend Implementation

#### New Functions
- `formatTime()` - Convert milliseconds to MM:SS format
- `updateClockDisplay()` - Update visual clock display
- `startClock()` - Start clock with 100ms update interval
- `stopClock()` - Stop clock
- `addIncrement()` - Add increment after move
- `setTimeControl()` - Configure time control
- `syncClockWithServer()` - Sync with backend state

#### Integration Points
- Clock starts on first move in `onDrop()`
- Increment added after each move (human and engine)
- Time control selector updates both frontend and backend
- Visual warnings for low time situations

## Files Modified

### Backend
1. `server/gamestate.go`
   - Added `TimeControl` struct
   - Added clock fields to `GameState`
   - Implemented clock management methods
   - Integrated clock updates with move handling

2. `server/chess.go`
   - Added `ClockRequest` and `ClockResponse` types
   - Implemented clock API handlers

3. `server/server.go`
   - Registered new clock API endpoints

### Frontend
1. `server/templates/index.html`
   - Added time control selector dropdown
   - Added chess clock display elements

2. `server/assets/css/chess-ui.css`
   - Styled time control selector
   - Styled chess clocks with active/warning states
   - Added pulse animation for critical time

3. `server/assets/js/chess-ui.js`
   - Added clock variables and state
   - Implemented clock functions
   - Integrated clock with move handlers
   - Added time control change handler

## Usage

1. **Select Time Control**: Choose from dropdown before starting game
2. **Make First Move**: Clock starts automatically
3. **Time Management**: Each move adds increment time
4. **Visual Feedback**: Active clock is highlighted, warnings appear when low on time
5. **Time Out**: Game ends automatically if a player runs out of time

## Technical Details

### Time Tracking Strategy
- Frontend: 100ms update interval for smooth display
- Backend: Precise time tracking on move completion
- Increment: Added immediately after move is made
- Synchronization: Frontend can sync with backend state

### Clock States
- **Not Started**: Initial state, clocks show starting time
- **Running**: Active clock counts down
- **Paused**: Game over or manually stopped
- **Time Out**: Player has run out of time

### Edge Cases Handled
- Unlimited time (0+0) - clock doesn't start
- Very long games (hours) - formatted as H:MM:SS
- Negative time - clamped to 0:00
- Page reload - time control preference restored
- Engine vs Engine - clock works for both sides

## Testing Recommendations

1. Test different time controls (bullet, blitz, rapid)
2. Verify increment is added after each move
3. Test time-out scenarios
4. Verify visual warnings appear correctly
5. Test with Human vs Human, Human vs Engine, Engine vs Engine
6. Verify clock state persists across page reloads (time control setting)
7. Test unlimited time mode (0+0)

## Future Enhancements (Optional)

- Pause/Resume button
- Custom time controls
- Time usage statistics
- Sound alerts for low time
- Pre-move support
- Delay increment (Fischer delay)
- Bronstein delay
- Time odds (different times for each player)
