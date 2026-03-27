package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entireio/cli/cmd/entire/cli/search"
	"github.com/entireio/cli/cmd/entire/cli/stringutil"
)

// searchMode tracks whether the user is browsing results or editing the search bar.
type searchMode int

const (
	modeBrowse searchMode = iota
	modeSearch
)

// searchResultsMsg is sent when a search API call completes.
type searchResultsMsg struct {
	results []search.Result
	total   int
	err     error
}

// searchStyles holds all lipgloss styles for the search TUI.
type searchStyles struct {
	useColor     bool
	sectionTitle lipgloss.Style // bold uppercase section headers
	label        lipgloss.Style // dim key labels in detail panel
	id           lipgloss.Style // amber for IDs/SHAs
	branch       lipgloss.Style // cyan for branch names
	dim          lipgloss.Style // dimmed secondary text
	bold         lipgloss.Style // bold emphasis
	selected     lipgloss.Style // highlighted selected row
	match        lipgloss.Style // green for match type
	helpKey      lipgloss.Style // colored key hints in footer
	helpSep      lipgloss.Style // dim separator dots in footer
	detailTitle  lipgloss.Style // colored title inside detail card
	detailBorder lipgloss.Style // border style for detail card
	errStyle     lipgloss.Style // red for errors
}

func newSearchStyles(colorEnabled bool) searchStyles {
	s := searchStyles{useColor: colorEnabled}
	if !colorEnabled {
		return s
	}
	s.sectionTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
	s.label = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	s.id = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	s.branch = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	s.dim = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	s.bold = lipgloss.NewStyle().Bold(true)
	s.selected = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	s.match = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	s.helpKey = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	s.helpSep = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	s.detailTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	s.detailBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("243")).
		Padding(1, 2)
	s.errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	return s
}

// searchModel is the bubbletea model for interactive search results.
type searchModel struct {
	results   []search.Result
	cursor    int
	total     int
	width     int
	height    int
	mode      searchMode
	loading   bool
	searchErr string
	input     textinput.Model
	searchCfg search.Config
	styles    searchStyles
}

func newSearchModel(results []search.Result, query string, total int, cfg search.Config, ss statusStyles) searchModel {
	styles := newSearchStyles(ss.colorEnabled)

	ti := textinput.New()
	ti.SetValue(query)
	ti.Prompt = " › "
	ti.Placeholder = "type a query to search checkpoints..."
	ti.CharLimit = 200
	ti.Width = max(ss.width-6, 30)
	if ss.colorEnabled {
		ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		ti.TextStyle = lipgloss.NewStyle()
		ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	}

	return searchModel{
		results:   results,
		cursor:    0,
		total:     total,
		width:     ss.width,
		mode:      modeBrowse,
		input:     ti,
		searchCfg: cfg,
		styles:    styles,
	}
}

func (m searchModel) Init() tea.Cmd {
	if m.mode == modeSearch {
		return textinput.Blink
	}
	return nil
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn,cyclop // bubbletea interface
	switch msg := msg.(type) {
	case searchResultsMsg:
		m.loading = false
		if msg.err != nil {
			m.searchErr = msg.err.Error()
			return m, nil
		}
		m.searchErr = ""
		m.results = msg.results
		m.total = msg.total
		m.cursor = 0
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = max(msg.Width-6, 30)
		return m, nil

	case tea.KeyMsg:
		if m.mode == modeSearch {
			return m.updateSearchMode(msg)
		}
		return m.updateBrowseMode(msg)
	}
	return m, nil
}

func (m searchModel) updateSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) { //nolint:ireturn // bubbletea pattern
	switch msg.String() {
	case "esc":
		m.mode = modeBrowse
		m.input.Blur()
		return m, nil
	case "enter":
		query := strings.TrimSpace(m.input.Value())
		if query == "" {
			return m, nil
		}
		m.mode = modeBrowse
		m.input.Blur()
		m.loading = true
		m.searchErr = ""
		cfg := m.searchCfg
		cfg.Query = query
		return m, performSearch(cfg)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m searchModel) updateBrowseMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) { //nolint:ireturn // bubbletea pattern
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
	case "/":
		m.mode = modeSearch
		m.input.Focus()
		return m, m.input.Cursor.SetMode(cursor.CursorBlink)
	}
	return m, nil
}

func performSearch(cfg search.Config) tea.Cmd {
	return func() tea.Msg {
		resp, err := search.Search(context.Background(), cfg)
		if err != nil {
			return searchResultsMsg{err: err}
		}
		return searchResultsMsg{results: resp.Results, total: resp.Total}
	}
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m searchModel) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder
	pad := " "

	// Section: SEARCH
	b.WriteString("\n")
	b.WriteString(pad + m.s(m.styles.sectionTitle, "SEARCH"))
	b.WriteString("\n\n")

	// Search input
	if m.mode == modeSearch {
		b.WriteString(pad + m.input.View())
	} else {
		query := m.input.Value()
		b.WriteString(pad + m.s(m.styles.id, "›") + " " + m.s(m.styles.bold, query))
	}
	b.WriteString("\n\n")

	// Loading / error / empty states
	if m.loading {
		b.WriteString(pad + m.s(m.styles.dim, "Searching...") + "\n")
		b.WriteString(m.viewHelp())
		return b.String()
	}
	if m.searchErr != "" {
		b.WriteString(pad + m.s(m.styles.errStyle, "Error: "+m.searchErr) + "\n")
		b.WriteString(m.viewHelp())
		return b.String()
	}
	if len(m.results) == 0 {
		b.WriteString(pad + m.s(m.styles.dim, "No results found.") + "\n")
		b.WriteString(m.viewHelp())
		return b.String()
	}

	// Section: RESULTS
	b.WriteString(pad + m.s(m.styles.sectionTitle, "RESULTS"))
	b.WriteString("\n\n")

	// Table
	b.WriteString(m.viewTable())
	b.WriteString("\n")

	// Detail card
	if m.cursor >= 0 && m.cursor < len(m.results) {
		b.WriteString(m.viewDetailCard(m.results[m.cursor]))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString(m.viewHelp())

	return b.String()
}

func (m searchModel) viewTable() string {
	contentWidth := m.width - 2 // 1 char padding each side
	cols := computeColumns(contentWidth)
	pad := " "

	var b strings.Builder

	// Column headers
	hdr := fmt.Sprintf("%-*s %-*s %-*s %-*s %s",
		cols.age, "Age",
		cols.id, "ID",
		cols.branch, "Branch",
		cols.prompt, "Prompt",
		"Author",
	)
	b.WriteString(pad + m.s(m.styles.dim, hdr) + "\n")

	// Header separator
	b.WriteString(pad + m.s(m.styles.dim, strings.Repeat("─", contentWidth)) + "\n")

	// Rows
	for i, r := range m.results {
		row := m.viewRow(r, cols)
		if i == m.cursor && m.styles.useColor {
			b.WriteString(pad + m.styles.selected.Render(row))
		} else {
			b.WriteString(pad + row)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m searchModel) viewRow(r search.Result, cols columnLayout) string {
	age := fmt.Sprintf("%-*s", cols.age, stringutil.TruncateRunes(formatSearchAge(r.Data.CreatedAt), cols.age, ""))
	id := fmt.Sprintf("%-*s", cols.id, stringutil.TruncateRunes(r.Data.ID, cols.id-1, "…"))
	branch := fmt.Sprintf("%-*s", cols.branch, stringutil.TruncateRunes(r.Data.Branch, cols.branch-1, "…"))
	prompt := fmt.Sprintf("%-*s", cols.prompt, stringutil.TruncateRunes(
		stringutil.CollapseWhitespace(r.Data.Prompt), cols.prompt-1, "…",
	))
	author := r.Data.Author

	return fmt.Sprintf("%s %s %s %s %s", age, id, branch, prompt, author)
}

func (m searchModel) viewDetailCard(r search.Result) string {
	const labelWidth = 12
	innerWidth := m.width - 8 // border + padding eats ~6-8 chars

	var content strings.Builder

	// Title
	content.WriteString(m.s(m.styles.detailTitle, "Checkpoint Detail"))
	content.WriteString("\n\n")

	writeField := func(label, value string) {
		lbl := fmt.Sprintf("%-*s", labelWidth, label+":")
		content.WriteString(m.s(m.styles.label, lbl) + " " + value + "\n")
	}

	writeField("ID", r.Data.ID)
	writeField("Prompt", r.Data.Prompt)

	// Commit
	commitSHA := derefStr(r.Data.CommitSHA, "—")
	if r.Data.CommitSHA != nil && len(*r.Data.CommitSHA) > 7 {
		commitSHA = (*r.Data.CommitSHA)[:7]
	}
	commitMsg := derefStr(r.Data.CommitMessage, "")
	if commitMsg != "" {
		writeField("Commit", commitSHA+" "+commitMsg)
	} else {
		writeField("Commit", commitSHA)
	}

	writeField("Branch", r.Data.Branch)
	writeField("Repo", r.Data.Org+"/"+r.Data.Repo)
	writeField("Author", formatAuthor(r.Data.Author, r.Data.AuthorUsername))
	writeField("Created", formatCreatedAt(r.Data.CreatedAt))
	writeField("Match", formatMatch(r.Meta))

	if r.Meta.Snippet != "" {
		content.WriteString("\n")
		content.WriteString(m.s(m.styles.label, "Snippet:") + "\n")
		content.WriteString(r.Meta.Snippet + "\n")
	}

	if len(r.Data.FilesTouched) > 0 {
		content.WriteString("\n")
		content.WriteString(m.s(m.styles.label, "Files:") + "\n")
		for _, f := range r.Data.FilesTouched {
			content.WriteString(f + "\n")
		}
	}

	cardContent := strings.TrimRight(content.String(), "\n")

	if !m.styles.useColor {
		// Plain text fallback — simple indent
		lines := strings.Split(cardContent, "\n")
		var plain strings.Builder
		for _, line := range lines {
			plain.WriteString(" " + line + "\n")
		}
		return plain.String()
	}

	card := m.styles.detailBorder.Width(max(innerWidth, 40)).Render(cardContent)

	// Indent the card by 1 space
	lines := strings.Split(card, "\n")
	var indented strings.Builder
	for _, line := range lines {
		indented.WriteString(" " + line + "\n")
	}
	return indented.String()
}

func (m searchModel) viewHelp() string {
	dot := m.s(m.styles.helpSep, " · ")

	if m.mode == modeSearch {
		return m.s(m.styles.helpKey, "enter") + " search" + dot +
			m.s(m.styles.helpKey, "esc") + " cancel" + "\n"
	}

	left := m.s(m.styles.helpKey, "/") + " search" + dot +
		m.s(m.styles.helpKey, "enter") + " select" + dot +
		m.s(m.styles.helpKey, "esc") + " unfocus" + dot +
		m.s(m.styles.helpKey, "j/k") + " navigate" + dot +
		m.s(m.styles.helpKey, "q") + " quit"

	right := fmt.Sprintf("%d results", m.total)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + m.s(m.styles.dim, right) + "\n"
}

// s applies a lipgloss style, returning plain text when color is off.
func (m searchModel) s(style lipgloss.Style, text string) string {
	if !m.styles.useColor {
		return text
	}
	return style.Render(text)
}

// ─── Column Layout ───────────────────────────────────────────────────────────

// columnLayout holds computed column widths for the search results table.
// Author column takes remaining space and is not width-constrained.
type columnLayout struct {
	age    int
	id     int
	branch int
	prompt int
}

// computeColumns calculates column widths from terminal width.
func computeColumns(width int) columnLayout {
	const (
		ageWidth    = 10
		idWidth     = 12
		authorWidth = 0 // author takes remaining
		gaps        = 4 // spaces between columns
	)

	remaining := width - ageWidth - idWidth - gaps
	if remaining < 20 {
		remaining = 20
	}

	branchWidth := max(remaining*20/100, 8)
	promptWidth := remaining - branchWidth

	return columnLayout{
		age:    ageWidth,
		id:     idWidth,
		branch: branchWidth,
		prompt: promptWidth,
	}
}

// ─── Formatting Helpers ──────────────────────────────────────────────────────

// formatSearchAge parses an RFC3339 timestamp and returns a relative time string.
func formatSearchAge(createdAt string) string {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return createdAt
	}
	return timeAgo(t)
}

// formatCommit renders commit SHA + message, handling nil pointers.
func formatCommit(sha, message *string) string {
	s := derefStr(sha, "—")
	if sha != nil && len(*sha) > 7 {
		s = (*sha)[:7]
	}
	msg := derefStr(message, "")
	if msg != "" {
		s += "  " + msg
	}
	return s
}

// formatAuthor renders author name with optional username.
func formatAuthor(author string, username *string) string {
	if username != nil && *username != "" {
		return author + " (@" + *username + ")"
	}
	return author
}

// formatCreatedAt renders a timestamp with relative time.
func formatCreatedAt(createdAt string) string {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return createdAt
	}
	return t.Format("Jan 02, 2006") + " (" + timeAgo(t) + ")"
}

// formatMatch renders match type and score.
func formatMatch(meta search.Meta) string {
	s := meta.MatchType
	if meta.Score > 0 {
		s += fmt.Sprintf(" (score: %.3f)", meta.Score)
	}
	return s
}

// derefStr returns the dereferenced string pointer, or fallback if nil.
func derefStr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

// ─── Static Fallback ─────────────────────────────────────────────────────────

// renderSearchStatic writes a non-interactive table for accessible mode.
func renderSearchStatic(w io.Writer, results []search.Result, query string, total int, styles statusStyles) {
	fmt.Fprintf(w, "Found %d checkpoints matching %q\n\n", total, query)

	cols := computeColumns(styles.width)

	fmt.Fprintf(w, "%-*s  %-*s  %-*s  %-*s  %s\n",
		cols.age, "AGE",
		cols.id, "ID",
		cols.branch, "BRANCH",
		cols.prompt, "PROMPT",
		"AUTHOR",
	)

	for _, r := range results {
		age := formatSearchAge(r.Data.CreatedAt)
		id := stringutil.TruncateRunes(r.Data.ID, cols.id, "")
		branch := stringutil.TruncateRunes(r.Data.Branch, cols.branch, "...")
		prompt := stringutil.TruncateRunes(
			stringutil.CollapseWhitespace(r.Data.Prompt), cols.prompt, "...",
		)
		author := r.Data.Author

		fmt.Fprintf(w, "%-*s  %-*s  %-*s  %-*s  %s\n",
			cols.age, age,
			cols.id, id,
			cols.branch, branch,
			cols.prompt, prompt,
			author,
		)
	}
}
