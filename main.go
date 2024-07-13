package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	SidebarLayoutStyle lipgloss.Style
	MainLayoutStyle    lipgloss.Style

	MainDisplayStyle lipgloss.Style
	MainInputStyle   lipgloss.Style
}

type Model struct {
	Sidebar string
	Main    MainModel

	Width  int
	Height int

	Loading bool

	styles *Styles
}

type MainModel struct {
	Display   string
	TextInput textinput.Model
}

type PokemonMsg struct {
	Pokemon Pokemon
}

type PokemonErrorMsg struct {
	Err error
}

func (p PokemonErrorMsg) Error() string {
	return p.Err.Error()
}

func defaultStyles() *Styles {
	return &Styles{
		SidebarLayoutStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Padding(1).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#00ff00")),

		MainLayoutStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Padding(1).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#00ff00")),

		MainDisplayStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Padding(1).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#00ff00")),

		MainInputStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Padding(1).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#00ff00")),
	}
}

func New() Model {
	styles := defaultStyles()
	ti := textinput.New()
	ti.Placeholder = "Search for a pokemon"
	ti.PromptStyle.Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("#00ff00")).Padding(1)
	ti.Focus()
	return Model{
		styles:  styles,
		Sidebar: "This is the side bar",
		Main: MainModel{
			Display:   "Search for a pokemon",
			TextInput: ti,
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
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
			searchValue := m.Main.TextInput.Value()
			if searchValue == "" {
				break
			}
			m.Main.TextInput.SetValue("")
			return m, func() tea.Msg {
				pokemon, err := getPokemon(strings.ToLower(searchValue))
				if err != nil {
					return PokemonErrorMsg{Err: err}
				}
				return PokemonMsg{Pokemon: pokemon}
			}
		}

	case PokemonMsg:
		m.Main.Display = fmt.Sprintf("Name: %s\nHeight: %d\nWeight: %d\nTypes: %s\nAbilities: %s",
			msg.Pokemon.Name,
			msg.Pokemon.Height,
			msg.Pokemon.Weight,
			strings.Join(msg.Pokemon.Types, ", "),
			strings.Join(msg.Pokemon.Abilities, ", "),
		)

	case PokemonErrorMsg:
		m.Main.Display = msg.Err.Error()
	}

	m.Main.TextInput, cmd = m.Main.TextInput.Update(msg)
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
			m.styles.SidebarLayoutStyle.Height(m.Height-3).Width(m.Width/5).Render(m.Sidebar),
			m.styles.MainLayoutStyle.Height(m.Height-3).Width(m.Width*4/5-3).Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					m.styles.MainDisplayStyle.Height(m.Height-12).Width(m.Width*4/5-9).Render(m.Main.Display),
					m.styles.MainInputStyle.Width(m.Width*4/5-9).Render(m.Main.TextInput.View()),
				),
			),
		),
	)
}

func main() {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func getPokemon(name string) (Pokemon, error) {
	const POKEMON_API = `https://pokeapi.co/api/v2/pokemon/`

	c := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := c.Get(POKEMON_API + name)
	if err != nil {
		return Pokemon{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return Pokemon{}, fmt.Errorf("pokemon not found")
		} else if resp.StatusCode == 429 {
			return Pokemon{}, fmt.Errorf("too many requests")
		}
		return Pokemon{}, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var pokemonResponse PokemonResponse
	err = json.NewDecoder(resp.Body).Decode(&pokemonResponse)
	if err != nil {
		return Pokemon{}, err
	}

	pokemon := formatPokemon(pokemonResponse)
	return pokemon, nil
}

func formatPokemon(pokemon PokemonResponse) Pokemon {
	PokemonTypes := []string{}
	for _, pokemonType := range pokemon.Types {
		PokemonTypes = append(PokemonTypes, pokemonType.Type.Name)
	}

	PokemonAbilities := []string{}
	for _, pokemonAbility := range pokemon.Abilities {
		PokemonAbilities = append(PokemonAbilities, pokemonAbility.Ability.Name)
	}

	return Pokemon{
		Name:      pokemon.Name,
		Height:    pokemon.Height,
		Weight:    pokemon.Weight,
		Types:     PokemonTypes,
		Abilities: PokemonAbilities,
	}
}
