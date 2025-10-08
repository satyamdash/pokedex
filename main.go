package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	pokecache "github.com/satyamdash/pokedex/internal"
)

type Config struct {
	Next     string
	Previous string
	Cache    *pokecache.Cache
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

func printLocations(location LocationResponse) {
	for _, result := range location.Results {
		fmt.Println(result.Name)
	}
}

func commandExit(c *Config) error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}
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
	// if *location.Previous != "null" {
	// 	c.Previous = *location.Previous
	// }
	printLocations(location)

	return nil
}

func commandHelp(c *Config) error {
	fmt.Print(`Welcome to the Pokedex!
Usage:

help: Displays a help message
exit: Exit the Pokedex`)
	return nil
}

func commandExplore(c *Config) error {

	return nil
}

func main() {
	cfg := &Config{
		Cache: pokecache.NewCache(5 * time.Second),
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
			println(cityname)
		}
		fmt.Printf("Your command was: %v\n", list[0])
	}
}
