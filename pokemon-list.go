package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type PokemonListModel struct {
	PokemonList list.Model
	Navigation  PokemonListNavigation
	isFocused   bool
	Page        int
}

type PokemonListNavigation struct {
	Next struct{}
	Prev struct{}
}

type PokemonListItem struct {
	title, desc string
}

func (i PokemonListItem) Title() string       { return i.title }
func (i PokemonListItem) Description() string { return i.desc }
func (i PokemonListItem) FilterValue() string { return i.title }

type listKeyMap struct {
	NextPage key.Binding
	PrevPage key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		PrevPage: key.NewBinding(
			key.WithKeys("<-"),
			key.WithHelp("<-", "Previous Page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("->"),
			key.WithHelp("->", "Next Page"),
		),
	}
}

func NewPokemonListModel() PokemonListModel {
	items := []list.Item{}
	listKeys := newListKeyMap()
	pl := list.New(items, list.NewDefaultDelegate(), 0, 0)
	pl.SetShowStatusBar(false)
	pl.SetShowTitle(false)
	pl.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.NextPage,
			listKeys.PrevPage,
		}
	}
	return PokemonListModel{
		PokemonList: pl,
		Navigation: PokemonListNavigation{
			Next: struct{}{},
			Prev: struct{}{},
		},
		isFocused: false,
		Page:      0,
	}
}

func getPokemonList(page int) (PokemonList, error) {
	offset := 20 * page
	POKEMON_API := fmt.Sprintf(`https://pokeapi.co/api/v2/pokemon/?offset=%d&limit=20`, offset)

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
			return PokemonList{}, fmt.Errorf("Pokemon list not found")
		} else if resp.StatusCode == 429 {
			return PokemonList{}, fmt.Errorf("too many requests")
		}
		return PokemonList{}, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var pokemonResponse PokemonListResponse
	err = json.NewDecoder(resp.Body).Decode(&pokemonResponse)
	if err != nil {
		return PokemonList{}, err
	}

	pokemonList := formatPokemonList(pokemonResponse)
	return pokemonList, nil
}

func formatPokemonList(pokemonListResponse PokemonListResponse) PokemonList {
	PokemonListResults := []string{}
	for _, pokemon := range pokemonListResponse.Results {
		PokemonListResults = append(PokemonListResults, pokemon.Name)
	}

	return PokemonList{
		Count:   pokemonListResponse.Count,
		Results: PokemonListResults,
	}
}

type PokemonListMsg struct {
	PokemonList PokemonList
}
