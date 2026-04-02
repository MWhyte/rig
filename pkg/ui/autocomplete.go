package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/sahilm/fuzzy"
)

// AutocompleteModel manages autocomplete suggestions for filter inputs.
type AutocompleteModel struct {
	textInput     textinput.Model
	suggestions   []string // All available suggestions
	filteredSugs  []string // Filtered/matched suggestions
	selectedIndex int      // Currently selected suggestion index
	scrollOffset  int      // Scroll offset for long lists
	fieldName     string   // Name of field being edited (e.g., "Country", "Genre")
}

// NewAutocompleteModel creates a new autocomplete model.
func NewAutocompleteModel() AutocompleteModel {
	ti := textinput.New()
	ti.Placeholder = "" // Empty placeholder to avoid rendering bug
	ti.CharLimit = 50

	return AutocompleteModel{
		textInput:     ti,
		suggestions:   []string{},
		filteredSugs:  []string{},
		selectedIndex: 0,
		scrollOffset:  0,
	}
}

// SetFieldName sets the name of the field being edited.
func (m *AutocompleteModel) SetFieldName(name string) {
	m.fieldName = name
}

// SetSuggestions updates the available suggestions and resets selection.
func (m *AutocompleteModel) SetSuggestions(suggestions []string) {
	m.suggestions = suggestions
	m.filteredSugs = suggestions
	m.selectedIndex = 0
	m.scrollOffset = 0
}

// Filter applies fuzzy matching to suggestions based on query.
func (m *AutocompleteModel) Filter(query string) {
	if query == "" {
		m.filteredSugs = m.suggestions
		m.selectedIndex = 0
		m.scrollOffset = 0
		return
	}

	// Use fuzzy matching
	matches := fuzzy.Find(query, m.suggestions)

	// Extract matched strings
	m.filteredSugs = make([]string, len(matches))
	for i, match := range matches {
		m.filteredSugs[i] = match.Str
	}

	// Reset selection
	m.selectedIndex = 0
	m.scrollOffset = 0
}

// GetSelected returns the currently selected suggestion, or empty string if none.
func (m *AutocompleteModel) GetSelected() string {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.filteredSugs) {
		return m.filteredSugs[m.selectedIndex]
	}
	return ""
}

// Update handles keyboard input for navigation.
func (m AutocompleteModel) Update(msg tea.Msg) (AutocompleteModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				// Adjust scroll offset if needed
				if m.selectedIndex < m.scrollOffset {
					m.scrollOffset = m.selectedIndex
				}
			}
			return m, nil

		case "down", "j":
			if m.selectedIndex < len(m.filteredSugs)-1 {
				m.selectedIndex++
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the autocomplete interface.
func (m *AutocompleteModel) View(width, height int) string {
	var content strings.Builder

	// Field name and text input
	content.WriteString("\n")
	fmt.Fprintf(&content, "  %s: ", m.fieldName)
	content.WriteString(m.textInput.View())
	content.WriteString("\n")

	// Calculate how many suggestions we can show
	// Reserve lines for: field name+input (2), help text (1) = 3 lines
	availableLines := height - 3
	if availableLines < 1 {
		availableLines = 1
	}

	// Show suggestions
	if len(m.filteredSugs) == 0 {
		// No matches
		noMatchStyle := lipgloss.NewStyle().Foreground(colorMuted)
		if m.textInput.Value() == "" {
			content.WriteString(noMatchStyle.Render("  Type to search..."))
		} else {
			content.WriteString(noMatchStyle.Render("  No matches found"))
		}
	} else {
		// Adjust scroll offset to keep selected item visible
		if m.selectedIndex >= m.scrollOffset+availableLines {
			m.scrollOffset = m.selectedIndex - availableLines + 1
		}
		if m.selectedIndex < m.scrollOffset {
			m.scrollOffset = m.selectedIndex
		}

		// Show visible suggestions
		endIndex := m.scrollOffset + availableLines
		if endIndex > len(m.filteredSugs) {
			endIndex = len(m.filteredSugs)
		}

		for i := m.scrollOffset; i < endIndex; i++ {
			suggestion := m.filteredSugs[i]

			// Truncate if too long
			maxLen := width - 6 // Account for "→ " prefix and padding
			if len(suggestion) > maxLen {
				suggestion = suggestion[:maxLen-3] + "..."
			}

			// Highlight selected item
			if i == m.selectedIndex {
				style := lipgloss.NewStyle().
					Foreground(colorAccent).
					Bold(true)
				fmt.Fprintf(&content, "  → %s\n", style.Render(suggestion))
			} else {
				style := lipgloss.NewStyle().Foreground(colorDim)
				fmt.Fprintf(&content, "    %s\n", style.Render(suggestion))
			}
		}

		// Show scroll indicator if there are more items
		if len(m.filteredSugs) > availableLines {
			scrollInfo := fmt.Sprintf("  (%d/%d)", m.selectedIndex+1, len(m.filteredSugs))
			scrollStyle := lipgloss.NewStyle().Foreground(colorMuted)
			content.WriteString(scrollStyle.Render(scrollInfo))
			content.WriteString("\n")
		}
	}

	// Help text (always at bottom)
	content.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(colorMuted)
	content.WriteString(helpStyle.Render("  ↑↓: select • enter: apply • esc: cancel"))

	return content.String()
}

// Focus focuses the text input.
func (m *AutocompleteModel) Focus() tea.Cmd {
	return m.textInput.Focus()
}

// Blur blurs the text input.
func (m *AutocompleteModel) Blur() {
	m.textInput.Blur()
}

// SetValue sets the text input value and resets cursor position.
func (m *AutocompleteModel) SetValue(value string) {
	m.textInput.SetValue(value)
	m.textInput.SetCursor(len(value)) // Move cursor to end
}

// Reset creates a fresh textinput with the given initial value.
func (m *AutocompleteModel) Reset(initialValue string) {
	// Create a completely new textinput to avoid cursor/state issues
	ti := textinput.New()
	ti.Placeholder = "" // Empty placeholder to avoid rendering bug
	ti.CharLimit = 50
	ti.SetValue(initialValue)
	ti.SetCursor(len(initialValue))
	m.textInput = ti
}

// Value returns the current text input value.
func (m *AutocompleteModel) Value() string {
	return m.textInput.Value()
}

// UpdateTextInput updates the text input component.
func (m AutocompleteModel) UpdateTextInput(msg tea.Msg) (AutocompleteModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}
