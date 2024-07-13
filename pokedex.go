package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type PokedexViewModel struct {
	Display   PokedexDisplay
	TextInput textinput.Model
}

type PokedexDisplay struct {
	Header string
	Body   string
}

func NewPokedexViewModel() PokedexViewModel {
	ti := textinput.New()
	ti.Placeholder = "Search for a pokemon"
	ti.PromptStyle.Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("#11cc11")).Padding(1)
	ti.CharLimit = 64
	ti.Focus()

	return PokedexViewModel{
		Display: PokedexDisplay{
			Header: "Pokedex",
			Body:   "Search for a pokemon",
		},
		TextInput: ti,
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

type PokemonMsg struct {
	Pokemon Pokemon
}

type PokemonErrorMsg struct {
	Err error
}

func (p PokemonErrorMsg) Error() string {
	return p.Err.Error()
}
