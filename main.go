package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	SidebarLayoutStyle lipgloss.Style
	MainLayoutStyle    lipgloss.Style

	MainDisplayStyle lipgloss.Style
	MainInputStyle   lipgloss.Style

	DisplayHeaderStyle lipgloss.Style
	DisplayBodyStyle   lipgloss.Style
}

type Model struct {
	// SIDEBAR
	Sidebar SidebarModel

	// ROUTES
	Pokedex     PokedexViewModel
	PokemonList PokemonListModel

	// DIMENSIONS
	Width  int
	Height int

	Loading bool

	//STYLES
	styles *Styles
}

func defaultStyles() *Styles {
	return &Styles{
		SidebarLayoutStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#aa00bb")).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#11cc11")),

		MainLayoutStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#aa00bb")).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#11cc11")),

		MainDisplayStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#aa00bb")).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#11cc11")),

		MainInputStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#aa00bb")).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#11cc11")),

		DisplayHeaderStyle: lipgloss.NewStyle().
			PaddingRight(1).
			PaddingLeft(1).
			Bold(true).
			Underline(true).
			Foreground(lipgloss.Color("#cc3555")),

		DisplayBodyStyle: lipgloss.NewStyle().
			PaddingRight(1).
			PaddingLeft(1).
			PaddingTop(1).
			Foreground(lipgloss.Color("#af7fef")),
	}
}

func defaultSidebarStyle() *SidebarStyle {
	return &SidebarStyle{
		FocusedSelectedStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")),
		FocusedUnselectedStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff")),
		UnfocusedSelectedStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#cc3555")),
		UnfocusedUnselectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#aa00aa")),
	}
}

func New() Model {
	m := NewPokedexViewModel()

	s := SidebarModel{
		Routes:         []string{"Pokedex", "Pokemon List"},
		SelectedRouted: 0,
		Styles:         defaultSidebarStyle(),
	}

	pl := NewPokemonListModel()

	return Model{
		styles:      defaultStyles(),
		PokemonList: pl,
		Sidebar:     s,
		Pokedex:     m,
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		pl, err := getPokemonList(0)
		if err != nil {
			return PokemonErrorMsg{Err: err}
		}
		return PokemonListMsg{PokemonList: pl}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":

			if m.Pokedex.isFocused {
				if m.Pokedex.TextInput.Focused() {
					searchValue := m.Pokedex.TextInput.Value()
					if searchValue == "" {
						break
					}
					m.Pokedex.TextInput.SetValue("")
					return m, func() tea.Msg {
						pokemon, err := getPokemon(strings.ToLower(searchValue))
						if err != nil {
							return PokemonErrorMsg{Err: err}
						}
						return PokemonMsg{Pokemon: pokemon}
					}
				}
			}

			if m.PokemonList.isFocused {
				m.PokemonList.isFocused = false
				m.Pokedex.isFocused = true
				m.Sidebar.SelectedRouted = 0
				m.Pokedex.TextInput.Focus()

				return m, func() tea.Msg {
					selectedItem := m.PokemonList.PokemonList.SelectedItem()
					pokemon, err := getPokemon(strings.ToLower(selectedItem.FilterValue()))
					if err != nil {
						return PokemonErrorMsg{Err: err}
					}
					return PokemonMsg{Pokemon: pokemon}
				}
			}

			if m.Sidebar.IsFocused {
				if m.Sidebar.Routes[m.Sidebar.SelectedRouted] == "Pokedex" {
					m.Pokedex.isFocused = true
					m.Pokedex.TextInput.Focus()
					m.Sidebar.IsFocused = false
				}
				if m.Sidebar.Routes[m.Sidebar.SelectedRouted] == "Pokemon List" {
					m.PokemonList.isFocused = true
					m.Sidebar.IsFocused = false

				}
			}
		case "tab":
			if m.Pokedex.isFocused {
				if m.Pokedex.TextInput.Focused() {
					m.Pokedex.TextInput.Blur()
					m.Sidebar.IsFocused = true
				}
			}
			if m.PokemonList.isFocused {
				m.Sidebar.IsFocused = true
			}

		case "left", "right":
			if m.PokemonList.isFocused {
				if msg.String() == "left" {
					if m.PokemonList.Page-1 > 0 {
						m.PokemonList.Page--
					}
				}
				if msg.String() == "right" {
					m.PokemonList.Page++
				}
				return m, func() tea.Msg {
					pl, err := getPokemonList(m.PokemonList.Page)
					if err != nil {
						return PokemonErrorMsg{Err: err}
					}
					return PokemonListMsg{PokemonList: pl}
				}

			}
		}

	case PokemonMsg:
		m.Pokedex.Display.Body = fmt.Sprintf("Name: %s\nHeight: %d\nWeight: %d\nTypes: %s\nAbilities: %s",
			msg.Pokemon.Name,
			msg.Pokemon.Height,
			msg.Pokemon.Weight,
			strings.Join(msg.Pokemon.Types, ", "),
			strings.Join(msg.Pokemon.Abilities, ", "),
		)

	case PokemonErrorMsg:
		m.Pokedex.Display.Body = msg.Err.Error()

	case PokemonListMsg:
		items := make([]list.Item, len(msg.PokemonList.Results))
		for i, listItem := range msg.PokemonList.Results {
			items[i] = PokemonListItem{
				title: listItem,
				desc:  "desc",
			}
		}
		m.PokemonList.PokemonList.SetItems(items)
	}

	if m.Sidebar.IsFocused {
		var sidebarCmd tea.Cmd
		var sidebarModel tea.Model
		sidebarModel, sidebarCmd = m.Sidebar.Update(msg)
		m.Sidebar = sidebarModel.(SidebarModel)
		cmd = tea.Batch(cmd, sidebarCmd)
	} else {
		if m.Sidebar.Routes[m.Sidebar.SelectedRouted] == "Pokedex" {
			m.Pokedex.TextInput, cmd = m.Pokedex.TextInput.Update(msg)
		}
		if m.Sidebar.Routes[m.Sidebar.SelectedRouted] == "Pokemon List" {
			m.PokemonList.PokemonList, cmd = m.PokemonList.PokemonList.Update(msg)
		}
	}

	return m, cmd
}

func (m Model) View() string {
	if m.Loading {
		return "Loading..."
	}

	return lipgloss.Place(
		m.Width,
		m.Height,
		lipgloss.Top,
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			/* SIDEBAR LAYOUT */
			m.styles.SidebarLayoutStyle.Height(m.Height-3).Width(m.Width/5).Render(m.Sidebar.View()),
			/* MAIN LAYOUT */
			CurrentView(m),
		),
	)
}

func main() {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func CurrentView(m Model) string {
	switch m.Sidebar.Routes[m.Sidebar.SelectedRouted] {
	case "Pokedex":
		return m.styles.MainLayoutStyle.Height(m.Height - 3).Width(m.Width*4/5 - 3).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				/* DISPLAY */
				m.styles.MainDisplayStyle.Height(m.Height-8).Width(m.Width*4/5-5).Render(
					lipgloss.JoinVertical(
						lipgloss.Left,
						m.styles.DisplayHeaderStyle.Render(m.Pokedex.Display.Header),
						m.styles.DisplayBodyStyle.Render(m.Pokedex.Display.Body),
					),
				),
				/* INPUT */
				m.styles.MainInputStyle.Width(m.Width*4/5-5).Render(m.Pokedex.TextInput.View()),
			),
		)
	case "Pokemon List":
		m.PokemonList.PokemonList.SetHeight(m.Height - 5)
		m.PokemonList.PokemonList.SetWidth(m.Width - 12)

		return m.styles.MainLayoutStyle.Height(m.Height - 3).Width(m.Width*4/5 - 3).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				/* LIST */
				m.styles.MainDisplayStyle.Height(m.Height-8).Width(m.Width*4/5-5).Render(
					lipgloss.JoinVertical(
						lipgloss.Left,
						m.PokemonList.PokemonList.View(),
						// m.styles.DisplayBodyStyle.Render(m.Pokedex.Display.Body),
					),
				),
			),
		)
	}
	return ""
}
