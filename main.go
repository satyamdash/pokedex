package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Config struct {
	Next     string
	Previous string
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

func commandExit(c *Config) error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}
func commandMapB(c *Config) error {
	var url string
	if c.Previous == "" {
		fmt.Print("you're on the first page")
		return nil
	}
	url = c.Previous
	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	var location LocationResponse
	if err := json.Unmarshal(data, &location); err != nil {
		return err
	}

	c.Previous = location.Previous
	var city []string
	for _, result := range location.Results {
		city = append(city, result.Name)
	}
	for _, cname := range city {
		fmt.Println(cname)
	}

	return nil
}
func commandMap(c *Config) error {
	var url string
	url = "https://pokeapi.co/api/v2/location"
	if c.Next != "" {
		url = c.Next
	}

	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	var location LocationResponse
	if err := json.Unmarshal(data, &location); err != nil {
		return err
	}

	c.Next = location.Next
	c.Previous = location.Previous
	// if *location.Previous != "null" {
	// 	c.Previous = *location.Previous
	// }
	var city []string
	for _, result := range location.Results {
		city = append(city, result.Name)
	}
	for _, cname := range city {
		fmt.Println(cname)
	}

	return nil
}

func commandHelp(c *Config) error {
	fmt.Print(`Welcome to the Pokedex!
Usage:

help: Displays a help message
exit: Exit the Pokedex`)
	return nil
}

func main() {
	cfg := &Config{}
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
		}
		fmt.Printf("Your command was: %v\n", list[0])
	}
}
