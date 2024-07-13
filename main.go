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

	DisplayHeaderStyle lipgloss.Style
	DisplayBodyStyle   lipgloss.Style
}

type Model struct {
	Sidebar SidebarModel
	Main    MainModel

	Width  int
	Height int

	Loading bool

	styles *Styles
}

type MainModel struct {
	Display   MainDisplay
	TextInput textinput.Model
}

type MainDisplay struct {
	Header string
	Body   string
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
	ti := textinput.New()
	ti.Placeholder = "Search for a pokemon"
	ti.PromptStyle.Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("#11cc11")).Padding(1)
	ti.CharLimit = 64
	ti.Focus()

	m := MainModel{
		Display: MainDisplay{
			Header: "Pokedex",
			Body:   "Search for a pokemon",
		},
		TextInput: ti,
	}

	s := SidebarModel{
		Routes:         []string{"Pokedex", "Moves", "Abilities", "Items", "Locations", "Type Chart"},
		SelectedRouted: 0,
		Styles:         defaultSidebarStyle(),
	}

	return Model{
		styles:  defaultStyles(),
		Sidebar: s,
		Main:    m,
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
			if m.Main.TextInput.Focused() {
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
			if m.Sidebar.IsFocused {
				if m.Sidebar.Routes[m.Sidebar.SelectedRouted] == "Pokedex" {
					m.Main.TextInput.Focus()
					m.Sidebar.IsFocused = false
				}
			}
		case "tab":
			if m.Main.TextInput.Focused() {
				m.Main.TextInput.Blur()
				m.Sidebar.IsFocused = true
			}
		}

	case PokemonMsg:
		m.Main.Display.Body = fmt.Sprintf("Name: %s\nHeight: %d\nWeight: %d\nTypes: %s\nAbilities: %s",
			msg.Pokemon.Name,
			msg.Pokemon.Height,
			msg.Pokemon.Weight,
			strings.Join(msg.Pokemon.Types, ", "),
			strings.Join(msg.Pokemon.Abilities, ", "),
		)

	case PokemonErrorMsg:
		m.Main.Display.Body = msg.Err.Error()
	}

	if m.Main.TextInput.Focused() {
		m.Main.TextInput, cmd = m.Main.TextInput.Update(msg)
	}
	if m.Sidebar.IsFocused {
		var sidebarCmd tea.Cmd
		var sidebarModel tea.Model
		sidebarModel, sidebarCmd = m.Sidebar.Update(msg)
		m.Sidebar = sidebarModel.(SidebarModel)
		cmd = tea.Batch(cmd, sidebarCmd)
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
			m.styles.MainLayoutStyle.Height(m.Height-3).Width(m.Width*4/5-3).Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					/* MAIN DISPLAY */
					m.styles.MainDisplayStyle.Height(m.Height-8).Width(m.Width*4/5-5).Render(
						lipgloss.JoinVertical(
							lipgloss.Left,
							m.styles.DisplayHeaderStyle.Render(m.Main.Display.Header),
							m.styles.DisplayBodyStyle.Render(m.Main.Display.Body),
						),
					),
					/* MAIN INPUT */
					m.styles.MainInputStyle.Width(m.Width*4/5-5).Render(m.Main.TextInput.View()),
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

func getPokemonList(page int) (PokemonList, error) {
	var POKEMON_API string
	if page != 0 {
		offset := page * 20
		POKEMON_API = fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/?offset=%d&limit=20", offset)
	} else {
		POKEMON_API = "https://pokeapi.co/api/v2/pokemon/?limit=20"
	}

	c := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := c.Get(POKEMON_API)
	if err != nil {
		return PokemonList{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return PokemonList{}, fmt.Errorf("pokemon not found")
		} else if resp.StatusCode == 429 {
			return PokemonList{}, fmt.Errorf("too many requests")
		}
		return PokemonList{}, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var pokemonListResponse PokemonListResponse
	err = json.NewDecoder(resp.Body).Decode(&pokemonListResponse)
	if err != nil {
		return PokemonList{}, err
	}

	var PokemonListResponse PokemonListResponse
	err = json.NewDecoder(resp.Body).Decode(&PokemonListResponse)
	if err != nil {
		return PokemonList{}, err
	}

	pokemonList := formatPokemonList(pokemonListResponse)
	return pokemonList, nil
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

func formatPokemonList(pokemonListResponse PokemonListResponse) PokemonList {
	PokemonListResults := []string{}
	for _, pokemon := range pokemonListResponse.Results {
		PokemonListResults = append(PokemonListResults, pokemon.Name)
	}

	return PokemonList{
		Count:    pokemonListResponse.Count,
		Next:     *pokemonListResponse.Next,
		Previous: *pokemonListResponse.Previous,
		Results:  PokemonListResults,
	}
}
