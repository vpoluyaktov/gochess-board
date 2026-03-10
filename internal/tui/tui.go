package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gochess-board/internal/engines"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
)

type tickMsg time.Time

type model struct {
	spinner      spinner.Model
	serverURL    string
	startTime    time.Time
	engines      []engines.EngineInfo
	monitor      *engines.EngineMonitor
	enginesTable table.Model
	activeTable  table.Model
	width        int
	height       int
	openingStats map[string]int
	bookLoaded   bool
	bookEntries  int
}

func InitialModel(serverURL string, engines []engines.EngineInfo, monitor *engines.EngineMonitor, openingStats map[string]int, bookLoaded bool, bookEntries int) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Engines table (static list)
	enginesColumns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "Name", Width: 28},
		{Title: "Version", Width: 10},
		{Title: "ELO / Strength", Width: 30},
		{Title: "Protocol", Width: 12},
	}
	t := table.New(
		table.WithColumns(enginesColumns),
		table.WithRows(buildEngineRows(engines)),
	)
	t.SetHeight(10)
	t.SetStyles(defaultTableStyles())

	// Active engines table (dynamic, refreshed on ticks)
	activeColumns := []table.Column{
		{Title: "Type", Width: 15},
		{Title: "Name", Width: 19},
		{Title: "ELO", Width: 6},
		{Title: "wtime", Width: 10},
		{Title: "btime", Width: 10},
		{Title: "winc", Width: 8},
		{Title: "binc", Width: 8},
	}
	activeT := table.New(
		table.WithColumns(activeColumns),
		table.WithRows(nil),
	)
	activeT.SetHeight(6)
	activeT.SetStyles(defaultTableStyles())

	m := model{
		spinner:      s,
		serverURL:    serverURL,
		startTime:    time.Now(),
		engines:      engines,
		monitor:      monitor,
		enginesTable: t,
		activeTable:  activeT,
		openingStats: openingStats,
		bookLoaded:   bookLoaded,
		bookEntries:  bookEntries,
	}

	// Initialize active engines table rows
	m.refreshActiveEnginesTable()

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
		tea.WindowSize(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "r" {
			m.refreshActiveEnginesTable()
		}

	case tickMsg:
		m.refreshActiveEnginesTable()
		cmds = append(cmds, tickCmd())

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
	}

	var cmd tea.Cmd
	m.enginesTable, cmd = m.enginesTable.Update(msg)
	cmds = append(cmds, cmd)
	m.activeTable, cmd = m.activeTable.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// Fallback size if we haven't received a WindowSizeMsg yet
	width := m.width
	height := m.height
	if width == 0 {
		width = 120
	}
	if height == 0 {
		height = 40
	}

	// Title
	title := titleStyle.Render("♖  GOCHESS BOARD SERVER  ♖")

	// Server status line
	serverStatus := labelStyle.Render("Server: ") + valueStyle.Render(m.serverURL) +
		"  " + m.spinner.View() + " " + valueStyle.Render("Running")

	// Server uptime
	uptime := time.Since(m.startTime).Round(time.Second)

	// Left column: Server Info (compact format)
	openingsInfo := ""
	if m.openingStats != nil && m.openingStats["total_openings"] > 0 {
		openingsInfo = fmt.Sprintf("\n📖 OPENINGS: %d | Nodes: %d | Depth: %d",
			m.openingStats["total_openings"],
			m.openingStats["total_nodes"],
			m.openingStats["max_depth"])
	}

	bookInfo := ""
	if m.bookLoaded {
		bookInfo = fmt.Sprintf("\n📚 BOOK: %d entries", m.bookEntries)
	}

	serverInfoContent := fmt.Sprintf("🖥️  SERVER STATUS\n"+
		"URL:     %s\n"+
		"Uptime:  %s%s%s\n\n"+
		"📡 API ENDPOINTS\n"+
		"• /api/computer-move\n"+
		"• /api/analysis\n"+
		"• /api/engines\n"+
		"• /api/opening\n",
		m.serverURL, uptime.String(), openingsInfo, bookInfo)

	// Layout calculations: split the screen vertically into top/bottom halves
	topHeight, bottomHeight := calculateHeights(height)

	// Split width between server info and engines list
	serverWidth := int(float64(width) * 0.32)
	if serverWidth < 30 {
		serverWidth = 30
	}
	enginesWidth := width - serverWidth - 6
	if enginesWidth < 40 {
		enginesWidth = 40
	}

	serverInfo := boxStyle.
		Width(serverWidth).
		Height(topHeight).
		Render(serverInfoContent)

	// Right column: Engines Info (table or fallback text)
	var enginesDisplay string
	if len(m.engines) == 0 {
		enginesContent := "⚠️  NO ENGINES FOUND\n\n" +
			"No UCI chess engines were discovered on this system.\n" +
			"Please install a chess engine like Stockfish."
		enginesDisplay = boxStyle.
			Width(enginesWidth).
			Height(topHeight).
			Render(enginesContent)
	} else {
		enginesHeader := fmt.Sprintf("🎮 DISCOVERED ENGINES (%d)", len(m.engines))
		enginesContent := enginesHeader + "\n\n" + m.enginesTable.View()
		enginesDisplay = boxStyle.
			Width(enginesWidth).
			Height(topHeight).
			Render(enginesContent)
	}

	// Arrange in columns
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		serverInfo,
		enginesDisplay,
	)

	// Active engines runtime box (table) occupying the bottom half
	// Count is based on table rows (already de-duplicated in refreshActiveEnginesTable)
	activeHeader := fmt.Sprintf("⚡ ACTIVE ENGINES (%d)", len(m.activeTable.Rows()))
	activeContent := activeHeader + "\n\n" + m.activeTable.View()
	activeEnginesBox := boxStyle.
		Width(width - 4).
		Height(bottomHeight).
		Render(activeContent)

	// Help text inside a small status box at the very bottom
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Press 'r' to refresh engines • 'q' or 'Ctrl+C' to quit")
	helpBox := boxStyle.
		Width(width - 4).
		Height(3). // fixed height for help box
		Render(helpText)

	// Combine all sections vertically without extra blank lines
	output := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s",
		title,
		serverStatus,
		columns,
		activeEnginesBox,
		helpBox,
	)

	return output
}

func (m *model) refreshActiveEnginesTable() {
	activeEngines := m.monitor.GetActiveEngines()

	// Separate analysis and move engines
	var analysisEngines []*engines.ActiveEngine
	var moveEngines []*engines.ActiveEngine

	for _, ae := range activeEngines {
		if ae.Type == engines.EngineTypeAnalysis {
			analysisEngines = append(analysisEngines, ae)
		} else {
			moveEngines = append(moveEngines, ae)
		}
	}

	// Build a unified list of all engines with their display info
	type engineRow struct {
		typ  string
		name string
		elo  string
		col3 string
		col4 string
		col5 string
		col6 string
	}
	var allEngines []engineRow

	for _, ae := range analysisEngines {
		eloStr := "N/A"
		if ae.ELO > 0 {
			eloStr = fmt.Sprintf("%d", ae.ELO)
		}
		allEngines = append(allEngines, engineRow{
			typ:  string(engines.EngineTypeAnalysis),
			name: ae.Name,
			elo:  eloStr,
			col3: fmt.Sprintf("%dms", ae.WhiteTime),
			col4: fmt.Sprintf("%dms", ae.BlackTime),
			col5: fmt.Sprintf("%dms", ae.WhiteIncrement),
			col6: fmt.Sprintf("%dms", ae.BlackIncrement),
		})
	}

	for _, ae := range moveEngines {
		eloStr := "N/A"
		if ae.ELO > 0 {
			eloStr = fmt.Sprintf("%d", ae.ELO)
		}
		allEngines = append(allEngines, engineRow{
			typ:  string(engines.EngineTypeMove),
			name: ae.Name,
			elo:  eloStr,
			col3: fmt.Sprintf("%dms", ae.WhiteTime),
			col4: fmt.Sprintf("%dms", ae.BlackTime),
			col5: fmt.Sprintf("%dms", ae.WhiteIncrement),
			col6: fmt.Sprintf("%dms", ae.BlackIncrement),
		})
	}

	// Add pooled engines (persistent engine mode)
	// Skip engines that are already shown as "Move" (currently active)
	if engines.GlobalEnginePool != nil {
		poolStats := engines.GlobalEnginePool.Stats()
		if pooledEngines, ok := poolStats["engines"].([]map[string]interface{}); ok {
			for _, pe := range pooledEngines {
				name := pe["name"].(string)
				gameID := pe["game_id"].(string)
				idleTime := pe["idle_time"].(string)

				// Skip if this engine is currently shown as a Move engine (same name AND gameId)
				isActive := false
				for _, ae := range moveEngines {
					if ae.Name == name && ae.GameID == gameID {
						isActive = true
						break
					}
				}
				if isActive {
					continue
				}

				allEngines = append(allEngines, engineRow{
					typ:  string(engines.EngineTypeIdle) + " (" + idleTime + ")",
					name: name,
					elo:  "-",
					col3: "-",
					col4: "-",
					col5: "-",
					col6: "-",
				})
			}
		}
	}

	// Sort by name for stable ordering
	sort.Slice(allEngines, func(i, j int) bool {
		return allEngines[i].name < allEngines[j].name
	})

	// Convert to table rows
	rows := make([]table.Row, 0, len(allEngines))
	for _, e := range allEngines {
		rows = append(rows, table.Row{e.typ, e.name, e.elo, e.col3, e.col4, e.col5, e.col6})
	}

	// Note: table.Model has value semantics, so we must reassign
	m.activeTable.SetRows(rows)
}

func defaultTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	return s
}

func buildEngineRows(engines []engines.EngineInfo) []table.Row {
	rows := make([]table.Row, 0, len(engines))
	for i, e := range engines {
		strength := "Full strength only"
		if e.SupportsLimitStrength {
			strength = fmt.Sprintf("%d-%d (default %d)", e.MinElo, e.MaxElo, e.DefaultElo)
		}
		protocol := strings.ToUpper(e.Type)
		if protocol == "" {
			protocol = "UCI"
		}
		version := e.Version
		if version == "" {
			version = "-"
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			e.Name,
			version,
			strength,
			protocol,
		})
	}
	return rows
}

// calculateHeights computes the heights for top and bottom sections based on terminal height
func calculateHeights(terminalHeight int) (topHeight, bottomHeight int) {
	titleLines := 1
	serverStatusLines := 1
	newlines := 4      // newlines between sections in fmt.Sprintf
	helpBoxHeight := 3 // small fixed height for help/status box

	reservedLines := titleLines + serverStatusLines + newlines + helpBoxHeight + 5
	availableHeight := terminalHeight - reservedLines
	if availableHeight < 10 {
		availableHeight = 10
	}

	// Split remaining space 50/50 between top row and active engines
	topHeight = availableHeight / 2
	bottomHeight = availableHeight - topHeight
	return topHeight, bottomHeight
}

// updateLayout adjusts table heights based on the current terminal size
func (m *model) updateLayout() {
	if m.height == 0 {
		return
	}
	topHeight, bottomHeight := calculateHeights(m.height)

	// Leave some room inside the boxes for headers and padding
	topTableHeight := topHeight - 4
	if topTableHeight < 3 {
		topTableHeight = 3
	}
	bottomTableHeight := bottomHeight - 4
	if bottomTableHeight < 3 {
		bottomTableHeight = 3
	}

	m.enginesTable.SetHeight(topTableHeight)
	m.activeTable.SetHeight(bottomTableHeight)
}

func RunTUI(url string, engines []engines.EngineInfo, monitor *engines.EngineMonitor, openingStats map[string]int, bookLoaded bool, bookEntries int) error {
	p := tea.NewProgram(InitialModel(url, engines, monitor, openingStats, bookLoaded, bookEntries), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
