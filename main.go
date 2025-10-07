package main

import (
	"bufio"
	"fmt"
	"os"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

func commandExit() error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Print(`Welcome to the Pokedex!
Usage:

help: Displays a help message
exit: Exit the Pokedex`)
	return nil
}

func main() {
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
			if err := commands["exit"].callback(); err != nil {
				fmt.Println(err)
			}
		case commands["help"].name:
			if err := commands["help"].callback(); err != nil {
				fmt.Println(err)
			}
		}
		fmt.Printf("Your command was: %v\n", list[0])
	}
}
