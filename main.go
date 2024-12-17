package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
    "time"
    "math/rand"
    "log"
    "github.com/RodolfoCamposGlz/pokedexcli/pokecache"
)

// Config struct to store pagination URLs and shared state
type Config struct {
	Next     string
	Previous string
}

// Response struct to parse the JSON response from PokeAPI
type PokeAPIResponse struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []LocationEntry `json:"results"`
}
type PokemonEncounter struct {
    Pokemon struct {
        Name string `json:"name"`
        URL  string `json:"url"`
    } `json:"pokemon"`
    VersionDetails []struct {
        Version struct {
            Name string `json:"name"`
            URL  string `json:"url"`
        } `json:"version"`
        EncounterDetails []struct {
            MinLevel int `json:"min_level"`
            MaxLevel int `json:"max_level"`
        } `json:"encounter_details"`
    } `json:"version_details"`
}

type PokeAreaNameResp struct {
    PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

// LocationEntry represents each location-area entry in the API response
type LocationEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Pokemon struct {
	Name      string  `json:"name"`
	Height    int     `json:"height"`
	Weight    int     `json:"weight"`
	Stats         []struct {
		Stat struct {
			Name string `json:"name"` // stat name (e.g., "hp", "attack")
			URL  string `json:"url"`  // URL to more details about the stat (optional)
		} `json:"stat"`
		BaseStat int `json:"base_stat"` // the base stat value (integer)
	} `json:"stats"`
	Types         []struct {
		Type struct {
			Name string `json:"name"` // type name (e.g., "normal", "flying")
		} `json:"type"`
	} `json:"types"`
	BaseExperience int `json:"base_experience"`
	URL       string  `json:"url"`
}
var pokedex = make(map[string]Pokemon)

// cliCommand defines a command structure with name, description, and a callback
type cliCommand struct {
	name string
	description string
	callback  func(*Config, []string) error // Accepts a pointer to Config
}

var cacheInstance *cache.Cache


func main() {
    cacheInstance = cache.NewCache(10 * time.Second)
	// Initialize the configuration
	config := &Config{
		Next: "https://pokeapi.co/api/v2/location-area?offset=0&limit=20", // Initial endpoint
		Previous: "",
	}

	// Define the available commands
	commands := map[string]cliCommand{
		"exit": {
			name: "exit",
			description: "Exit the Pokedex",
			callback: commandExit,
		},
		"help": {
			name: "help",
			description: "Show the list of available commands",
			callback: commandHelp,
		},
		"map": {
			name: "map",
			description: "Show next paginated location areas",
			callback: commandMap,
		},
        "mapb": {
			name: "mapb",
			description: "Show previous paginated location areas",
			callback: commandMapBack,
		},
        "explore": {
            name: "explore",
            description: "It takes the name of a location area as an argument",
            callback: commandExplore,
        },
        "catch": {
            name: "catch",
            description: "Catch a pokemon!",
            callback: commandCatch,
        },
        "inspect": {
            name: "inspect",
            description: "Inspect a pokemon!",
            callback: commandInspect,
        },
        "pokedex": {
            name: "pokedex",
            description: "See all caught pokemons",
            callback: commandPokedex,
        },
	}

	// Start the input scanner
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")

		if scanner.Scan() {
			input := scanner.Text()
			words := cleanInput(input)

			// Process the input
			if len(words) > 0 {
				command, exists := commands[words[0]]
				if exists {
					// Run the command's callback function, passing the config
					err := command.callback(config, words[1:])
					if err != nil {
						fmt.Println("Error:", err)
					}
				} else {
					fmt.Println("Unknown command. Type 'help' to see available commands.")
				}
			} else {
				fmt.Println("No command entered.")
			}
		}
	}
}

// cleanInput normalizes input and splits it into words
func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

// commandExit exits the program
func commandExit(config *Config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

// commandHelp displays all available commands
func commandHelp(config *Config, args []string) error {
	fmt.Println("Available commands:")
	fmt.Println("- exit: Exit the Pokedex")
	fmt.Println("- help: Show this help message")
	fmt.Println("- map: Show paginated location areas")
    fmt.Println("- explore: Explore areas")
    fmt.Println("- catch: catch a pokemon ")
    fmt.Println("- inspect: inspect a pokemon ")
    fmt.Println("- pokedex: see all catched pokemons ")
	return nil
}

// commandMap fetches and displays paginated location areas
func commandMap(config *Config, args []string) error {
    return fetchAndDisplayLocations(config.Next, config)
}

func commandMapBack(config *Config, args []string) error {
    if config.Previous == "" {
        fmt.Println("No previous pages to fetch.")
        return nil
    }
    return fetchAndDisplayLocations(config.Previous, config)
}

func fetchAndDisplayLocations(url string, config *Config) error {
    if url == "" {
        fmt.Println("No more pages to fetch.")
        return nil
    }
	cached, exists := cacheInstance.Get(url)
	if !exists {
		// Fetch data from the API
		response, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch data: %v", err)
		}
		defer response.Body.Close()

		// Decode the JSON response
		var apiResponse PokeAPIResponse
		if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
			return fmt.Errorf("failed to parse response: %v", err)
		}

		// Marshal the data and cache it
		data, err := json.Marshal(apiResponse)
		if err != nil {
			log.Fatal("Error marshaling data:", err)
		}
		cacheInstance.Add(url, data)

		// Display the results
		fmt.Println("\nLocation Areas:")
		for _, location := range apiResponse.Results {
			fmt.Println("-", location.Name)
		}

		config.Next = apiResponse.Next
		config.Previous = apiResponse.Previous

		if config.Next != "" {
			fmt.Println("\nType 'map' to see the next page of locations.")
		} else {
			fmt.Println("\nNo more pages available.")
		}
	} else {
		// If data exists in cache, unmarshal it back to []LocationEntry
		var cachedResponse PokeAPIResponse
		err := json.Unmarshal(cached, &cachedResponse)
		if err != nil {
			return fmt.Errorf("failed to unmarshal cached data: %v", err)
		}
		fmt.Println("\nLocation Areas Back:")
		for _, location := range cachedResponse.Results {
			fmt.Println("-", location.Name)
		}
        config.Next = cachedResponse.Next
        config.Previous = cachedResponse.Previous
	}

	return nil
}

func commandExplore(config *Config, args []string) error{
	if len(args) < 1 {
		return fmt.Errorf("you must specify a location area to explore")
	}
	location := args[0]
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", location)

    cached, exists := cacheInstance.Get(url)
    if !exists {
        response, err := http.Get(url)
        if err != nil {
            return fmt.Errorf("failed to fetch data: %v", err)
        }
        defer response.Body.Close()
    
        var apiResponse PokeAreaNameResp
        if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
            return fmt.Errorf("failed to parse response: %v", err)
        }
    		// Marshal the data and cache it
		data, err := json.Marshal(apiResponse)
		if err != nil {
			log.Fatal("Error marshaling data:", err)
		}
		cacheInstance.Add(url, data)

        // Display Pokemon encounters in the specified location area
        fmt.Println("Pokemon encounters in", location+":")
        for _, pokemon := range apiResponse.PokemonEncounters {
            fmt.Println("-", pokemon.Pokemon.Name)
        }
    } else {
        // If data exists in cache, unmarshal it back to []LocationEntry
		var cachedResponse PokeAreaNameResp
		err := json.Unmarshal(cached, &cachedResponse)
		if err != nil {
			return fmt.Errorf("failed to unmarshal cached data: %v", err)
		}
		fmt.Println("\nLocation Areas Back:")
		for _, pokemon := range cachedResponse.PokemonEncounters {
			fmt.Println("-", pokemon.Pokemon.Name)
		}
    }
	return nil
}
func commandCatch(config *Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("you must specify a location area to explore")
	}
	pokemonName := args[0]
	// Fetch the Pokémon details from the PokeAPI
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch Pokémon: %v", err)
	}
	defer response.Body.Close()

	// Parse the Pokémon data
	var pokemon Pokemon
	if err := json.NewDecoder(response.Body).Decode(&pokemon); err != nil {
		return fmt.Errorf("failed to parse Pokémon response: %v", err)
	}

	// Print message before attempting to catch
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	// Use base experience to calculate the catch chance
	rand.Seed(time.Now().UnixNano())
	catchChance := 100 - pokemon.BaseExperience // Lower base experience = higher chance of catching
	if rand.Intn(100) < catchChance {
		// Pokémon is caught
		fmt.Printf("%s was caught!\n", pokemonName)
		pokedex[pokemonName] = pokemon
	} else {
		// Pokémon escaped
		fmt.Printf("%s escaped!\n", pokemonName)
	}

	return nil
}

func commandInspect(config *Config, args []string) error {
    if len(args) < 1 {
		return fmt.Errorf("you must specify a location area to explore")
	}
	pokemonName := args[0]

	// Check if the Pokémon has been caught
	pokemon, caught := pokedex[pokemonName]
	if !caught {
		fmt.Printf("You have not caught %s yet.\n", pokemonName)
		return nil
	}

	// Display detailed information about the Pokémon
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}

	return nil
}

func commandPokedex(config *Config, args []string)error{
    for _, pokemon := range pokedex {
		fmt.Printf("  - %s\n", pokemon.Name)
	}
    return nil
}