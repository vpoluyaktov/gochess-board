package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go-chess/server"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			MarginTop(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))
)

type tickMsg time.Time

type model struct {
	spinner  spinner.Model
	gameState *server.GameState
	serverURL string
}

func InitialModel(serverURL string) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	return model{
		spinner:   s,
		gameState: server.GetGameState(),
		serverURL: serverURL,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	
	case tickMsg:
		return m, tickCmd()
	
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	
	return m, nil
}

func (m model) View() string {
	moves, lastMove, lastMoveTime, stockfishTime, fen, _, whiteMoves, blackMoves := m.gameState.GetStats()
	
	// Title
	title := titleStyle.Render("♟️  CHESS vs STOCKFISH  ♟️")
	
	// Server status line
	serverStatus := labelStyle.Render("Server: ") + valueStyle.Render(m.serverURL) + 
		"  " + m.spinner.View() + " " + valueStyle.Render("Running")
	
	// Game Stats (left column)
	uptime := time.Since(m.gameState.GameStarted).Round(time.Second)
	gameStats := boxStyle.Copy().Width(35).Render(
		statsStyle.Render("📊 GAME STATS\n") +
		labelStyle.Render("Total Moves:  ") + valueStyle.Render(fmt.Sprintf("%d", moves)) + "\n" +
		labelStyle.Render("White Moves:  ") + valueStyle.Render(fmt.Sprintf("%d", whiteMoves)) + "\n" +
		labelStyle.Render("Black Moves:  ") + valueStyle.Render(fmt.Sprintf("%d", blackMoves)) + "\n" +
		labelStyle.Render("Duration:     ") + valueStyle.Render(uptime.String()),
	)
	
	// Stockfish Info (middle column)
	var stockfishInfo string
	if lastMove != "" {
		lastMoveAgo := time.Since(lastMoveTime).Round(time.Second).String() + " ago"
		stockfishInfo = boxStyle.Copy().Width(40).Render(
			statsStyle.Render("🤖 STOCKFISH\n") +
			labelStyle.Render("Last Move:    ") + valueStyle.Render(lastMove) + "\n" +
			labelStyle.Render("Think Time:   ") + valueStyle.Render(stockfishTime.Round(time.Millisecond).String()) + "\n" +
			labelStyle.Render("Played:       ") + valueStyle.Render(lastMoveAgo) + "\n" +
			labelStyle.Render("ELO:          ") + valueStyle.Render("~3500"),
		)
	} else {
		stockfishInfo = boxStyle.Copy().Width(40).Render(
			statsStyle.Render("🤖 STOCKFISH\n") +
			labelStyle.Render("Status: ") + valueStyle.Render("Waiting..."),
		)
	}
	
	// Position Info (right column)
	fenDisplay := fen
	if len(fen) > 45 {
		fenDisplay = fen[:45] + "..."
	}
	positionInfo := boxStyle.Copy().Width(55).Render(
		statsStyle.Render("📍 POSITION\n") +
		labelStyle.Render("FEN:\n") + valueStyle.Render(fenDisplay),
	)
	
	// Arrange in columns
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		gameStats,
		stockfishInfo,
		positionInfo,
	)
	
	// Help text
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Press 'q' or 'Ctrl+C' to quit")
	
	// Combine all sections vertically
	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		title,
		serverStatus,
		columns,
		help,
	)
}

func RunTUI(serverURL string) error {
	p := tea.NewProgram(InitialModel(serverURL))
	_, err := p.Run()
	return err
}
