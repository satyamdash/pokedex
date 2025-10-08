package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	pokecache "github.com/satyamdash/pokedex/internal"
)

type Config struct {
	Next     string
	Previous string
	Cache    *pokecache.Cache
	Location string
	Pokename string
	Pokedex  map[string]PokemonDetail
}

type cliCommand struct {
	name        string
	description string
	callback    func(c *Config) error
}

type LocationResponse struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"` // null in JSON → pointer so it can be nil
	Results  []Location `json:"results"`
}

type Location struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type LocationArea struct {
	Areas []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"areas"`
}

type PokemonLocation struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type PokemonDetail struct {
	BaseExperience int `json:"base_experience"`
	Height         int `json:"height"`
	Weight         int `json:"weight"`
	Types          []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
}

func printLocations(location LocationResponse) {
	for _, result := range location.Results {
		fmt.Println(result.Name)
	}
}

// ---------------------Command exit---------------------------------//
func commandExit(c *Config) error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

// ---------------------Command MapB---------------------------------//
func commandMapB(c *Config) error {
	if c.Previous == "" {
		fmt.Print("you're on the first page")
		return nil
	}
	url := c.Previous
	var location LocationResponse
	stream, flag := c.Cache.Get(url)
	if flag {
		fmt.Println("------------------------------------Cache Hit----------------------------")
		if err := json.Unmarshal(stream, &location); err != nil {
			return err
		}
		printLocations(location)
		return nil

	}
	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	//Cache URL and data
	c.Cache.Add(url, data)

	if err := json.Unmarshal(data, &location); err != nil {
		return err
	}

	c.Previous = location.Previous
	printLocations(location)

	return nil
}

// ---------------------Command Map---------------------------------//
func commandMap(c *Config) error {
	var url string
	if c.Next != "" {
		url = c.Next
	} else {
		url = "https://pokeapi.co/api/v2/location"
	}

	stream, flag := c.Cache.Get(url)
	var location LocationResponse

	if flag {
		fmt.Println("------------------------------------Cache Hit----------------------------")
		if err := json.Unmarshal(stream, &location); err != nil {
			return err
		}
		printLocations(location)
		return nil

	}

	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	//Cache URL and data
	c.Cache.Add(url, data)

	if err := json.Unmarshal(data, &location); err != nil {
		return err
	}

	c.Next = location.Next
	c.Previous = location.Previous

	printLocations(location)

	return nil
}

// ---------------------Command Help---------------------------------//
func commandHelp(c *Config) error {
	fmt.Print(`Welcome to the Pokedex!
Usage:

help: Displays a help message
exit: Exit the Pokedex`)
	return nil
}

// ---------------------Command EXPLORE---------------------------------//
func commandExplore(c *Config) error {
	url := "https://pokeapi.co/api/v2/location/" + c.Location

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	var locationarea LocationArea
	if err := json.Unmarshal(data, &locationarea); err != nil {
		return err
	}
	var allurls []string
	for _, val := range locationarea.Areas {
		allurls = append(allurls, val.URL)
	}

	for _, url := range allurls {
		res, err := http.Get(url)
		if err != nil {
			return err
		}

		defer res.Body.Close()
		data, _ := io.ReadAll(res.Body)

		var pokeloc PokemonLocation

		if err := json.Unmarshal(data, &pokeloc); err != nil {
			return err
		}
		for _, encounter := range pokeloc.PokemonEncounters {
			fmt.Println(encounter.Pokemon.Name)
		}
	}
	return nil

}

func commandCatch(c *Config) error {
	url := "https://pokeapi.co/api/v2/pokemon/" + c.Pokename
	fmt.Printf("Throwing a Pokeball at %s...", c.Pokename)
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	data, _ := io.ReadAll(res.Body)

	var pokedetail PokemonDetail
	if err := json.Unmarshal(data, &pokedetail); err != nil {
		return err
	}
	chance := rand.Intn(pokedetail.BaseExperience + 120) // random value between 0 and baseExp+100
	if chance > pokedetail.BaseExperience {
		fmt.Printf("%s was caught!", c.Pokename)
		c.Pokedex[c.Pokename] = pokedetail
	} else {
		fmt.Printf("%s escaped!", c.Pokename)
	}
	return nil
}

func commandInspect(c *Config) error {
	pokemon, ok := c.Pokedex[c.Pokename]

	if !ok {
		fmt.Println("you have not caught that pokemon")
	}
	fmt.Println("Name:", c.Pokename)
	fmt.Println("Height:", pokemon.Height)
	fmt.Println("Weight:", pokemon.Weight)

	fmt.Println("Stats:")
	for _, s := range pokemon.Stats {
		fmt.Printf("  - %s: %d\n", s.Stat.Name, s.BaseStat)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}
	return nil
}

func commandPokedex(c *Config) error {
	fmt.Println("Your Pokedex:")
	if len(c.Pokedex) == 0 {
		fmt.Println("Sorry no pokemon caught")
		return nil
	}
	for k := range c.Pokedex {
		fmt.Println("- ", k)
	}
	return nil
}

func main() {
	cfg := &Config{
		Cache:   pokecache.NewCache(5 * time.Second),
		Pokedex: make(map[string]PokemonDetail),
	}
	scanner := bufio.NewScanner(os.Stdin)
	commands := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Show available commands",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Shows next 20 location at once",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Shows previous 20 location at once",
			callback:    commandMapB,
		},
		"explore": {
			name:        "explore",
			description: "explore the given location",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "catch the pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "inspect the pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Show all my pokemons",
			callback:    commandPokedex,
		},
	}

	for {
		fmt.Print("Pokedex >")

		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		// fmt.Println(line)
		list := cleanInput(line)
		switch list[0] {
		case commands["exit"].name:
			if err := commands["exit"].callback(cfg); err != nil {
				fmt.Println(err)
			}

		case commands["help"].name:
			if err := commands["help"].callback(cfg); err != nil {
				fmt.Println(err)
			}

		case commands["map"].name:
			if err := commands["map"].callback(cfg); err != nil {
				fmt.Println(err)
			}
		case commands["mapb"].name:
			if err := commands["mapb"].callback(cfg); err != nil {
				fmt.Println(err)
			}
		case commands["explore"].name:
			cityname := list[1]
			cfg.Location = cityname
			fmt.Println(cfg.Location)
			// if err := commands["explore"].callback(cfg); err != nil {
			if err := commands["explore"].callback(cfg); err != nil {
				fmt.Println(err)
			}
		case commands["catch"].name:
			pokename := list[1]
			cfg.Pokename = pokename
			fmt.Println(cfg.Pokename)
			// if err := commands["explore"].callback(cfg); err != nil {
			if err := commands["catch"].callback(cfg); err != nil {
				fmt.Println(err)
			}
		case commands["inspect"].name:
			pokename := list[1]
			cfg.Pokename = pokename
			fmt.Println(cfg.Pokename)
			// if err := commands["explore"].callback(cfg); err != nil {
			if err := commands["inspect"].callback(cfg); err != nil {
				fmt.Println(err)
			}
		case commands["pokedex"].name:
			// if err := commands["explore"].callback(cfg); err != nil {
			if err := commands["pokedex"].callback(cfg); err != nil {
				fmt.Println(err)
			}
		}
		fmt.Printf("Your command was: %v\n", list[0])
	}
}
