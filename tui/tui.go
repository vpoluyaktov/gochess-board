package tui

import (
	"fmt"
	"strings"
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
	spinner   spinner.Model
	serverURL string
	startTime time.Time
	engines   []server.EngineInfo
	monitor   *server.EngineMonitor
}

func InitialModel(serverURL string, engines []server.EngineInfo, monitor *server.EngineMonitor) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	return model{
		spinner:   s,
		serverURL: serverURL,
		startTime: time.Now(),
		engines:   engines,
		monitor:   monitor,
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
	// Title
	title := titleStyle.Render("♟️  GO CHESS SERVER  ♟️")
	
	// Server status line
	serverStatus := labelStyle.Render("Server: ") + valueStyle.Render(m.serverURL) + 
		"  " + m.spinner.View() + " " + valueStyle.Render("Running")
	
	// Server uptime
	uptime := time.Since(m.startTime).Round(time.Second)
	
	// Left column: Server Info
	serverInfoContent := fmt.Sprintf("🖥️  SERVER STATUS\n\n"+
		"URL:     %s\n"+
		"Uptime:  %s\n"+
		"Mode:    Stateless\n\n"+
		"📡 API ENDPOINTS\n\n"+
		"• /api/computer-move\n"+
		"• /api/analysis\n"+
		"• /api/engines\n\n"+
		"ℹ️  INFO\n\n"+
		"Multi-user support\n"+
		"Client-side state\n"+
		"LocalStorage persist",
		m.serverURL, uptime.String())
	serverInfo := boxStyle.Copy().Width(43).Render(serverInfoContent)
	
	// Right column: Engines Info
	var enginesDisplay string
	if len(m.engines) == 0 {
		enginesContent := "⚠️  NO ENGINES FOUND\n\n" +
			"No UCI chess engines were discovered on this system.\n" +
			"Please install a chess engine like Stockfish."
		enginesDisplay = boxStyle.Copy().Width(83).Render(enginesContent)
	} else {
		// Build engines list as plain text
		enginesContent := fmt.Sprintf("🎮 DISCOVERED ENGINES (%d)\n\n", len(m.engines))
		
		for i, engine := range m.engines {
			if i > 0 {
				enginesContent += "\n───────────────────────────────────────────────────────────────────────────\n"
			}
			
			// Engine name and path
			enginesContent += fmt.Sprintf("%d. %s\n", i+1, engine.Name)
			enginesContent += fmt.Sprintf("   Path: %s\n", engine.Path)
			
			// ELO support
			if engine.SupportsLimitStrength {
				enginesContent += fmt.Sprintf("   ELO:  %d - %d (default: %d)\n",
					engine.MinElo, engine.MaxElo, engine.DefaultElo)
				enginesContent += "   Features: ✓ Strength Limiting\n"
			} else {
				enginesContent += "   ELO:  Full strength only\n"
			}
			
			// Options count
			optionCount := len(engine.Options)
			if optionCount > 0 {
				enginesContent += fmt.Sprintf("   UCI Options: %d available\n", optionCount)
			}
		}
		
		enginesDisplay = boxStyle.Copy().Width(83).Render(enginesContent)
	}
	
	// Arrange in columns
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		serverInfo,
		enginesDisplay,
	)
	
	// Active engines runtime box
	var activeEnginesBox string
	activeEngines := m.monitor.GetActiveEngines()
	if len(activeEngines) > 0 {
		// Separate analysis and move engines
		var analysisEngines []*server.ActiveEngine
		var moveEngines []*server.ActiveEngine
		
		for _, ae := range activeEngines {
			if strings.Contains(ae.Name, "(Analysis)") {
				analysisEngines = append(analysisEngines, ae)
			} else {
				moveEngines = append(moveEngines, ae)
			}
		}
		
		// Build content as plain text
		content := fmt.Sprintf("⚡ ACTIVE ENGINES (%d)\n\n", len(activeEngines))
		
		// Show analysis engines first
		for _, ae := range analysisEngines {
			eloStr := "N/A"
			if ae.ELO > 0 {
				eloStr = fmt.Sprintf("%d", ae.ELO)
			}
			content += fmt.Sprintf("• %s ELO:%s wtime:%dms btime:%dms winc:%dms binc:%dms\n",
				ae.Name, eloStr, ae.WhiteTime, ae.BlackTime, ae.WhiteIncrement, ae.BlackIncrement)
		}
		
		// Then show move engines
		for _, ae := range moveEngines {
			eloStr := "N/A"
			if ae.ELO > 0 {
				eloStr = fmt.Sprintf("%d", ae.ELO)
			}
			content += fmt.Sprintf("• %s ELO:%s wtime:%dms btime:%dms winc:%dms binc:%dms\n",
				ae.Name, eloStr, ae.WhiteTime, ae.BlackTime, ae.WhiteIncrement, ae.BlackIncrement)
		}
		
		activeEnginesBox = boxStyle.Copy().Width(128).Render(content)
	} else {
		content := fmt.Sprintf("⚡ ACTIVE ENGINES (0)\n\nNo engines currently running")
		activeEnginesBox = boxStyle.Copy().Width(128).Render(content)
	}
	
	// Help text
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Press 'r' to refresh engines • 'q' or 'Ctrl+C' to quit")
	
	// Combine all sections vertically
	output := fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s\n\n%s",
		title,
		serverStatus,
		columns,
		activeEnginesBox,
		help,
	)
	
	return output
}

func RunTUI(serverURL string, engines []server.EngineInfo, monitor *server.EngineMonitor) error {
	p := tea.NewProgram(InitialModel(serverURL, engines, monitor))
	_, err := p.Run()
	return err
}
