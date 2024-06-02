package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kalverra/minimal-block-explorer/internal/view"
)

var config Config

type Config struct {
	Default string
	Chains  map[string]ChainConfig
}

type ChainConfig struct {
	URL string `toml:"url"`
}

func init() {
	settingsFile := "./settings.toml"
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		return
	}
	if _, err := toml.DecodeFile(settingsFile, &config); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	fmt.Printf("Loaded config from %s\n", settingsFile)
}

func main() {
	var (
		chainURL      string
		selectedChain string
	)
	flag.StringVar(&chainURL, "url", "", "URL of the blockchain to connect to.")
	flag.StringVar(&selectedChain, "chain", "", "Name of the chain to connect to, defined in the settings.toml file.")
	flag.Parse()

	if (chainURL != "" && selectedChain != "") || (chainURL == "" && selectedChain == "" && config.Default == "") {
		fmt.Println("Provide either url, chain, or set a default chain in settings.toml.")
		os.Exit(1)
	}

	if chainURL != "" {
		fmt.Printf("Connecting to chain at %s\n", chainURL)
	} else if selectedChain != "" {
		fmt.Printf("Connecting to chain '%s'\n", selectedChain)
		chainConfig, ok := config.Chains[selectedChain]
		if !ok {
			fmt.Printf("Chain '%s' not found in settings.toml\n", selectedChain)
			if len(config.Chains) == 0 {
				fmt.Println("No chains found in settings.toml")
			} else {
				fmt.Println("All chains in settings.toml:")
				for chain := range config.Chains {
					fmt.Println(chain)
				}
			}
			os.Exit(1)
		}
		chainURL = chainConfig.URL
	} else if config.Default != "" {
		fmt.Printf("Connecting to default chain '%s'\n", config.Default)
		chainConfig, ok := config.Chains[config.Default]
		if !ok {
			fmt.Printf("Default chain '%s' not found in settings.toml\n", config.Default)
			if len(config.Chains) == 0 {
				fmt.Println("No chains found in settings.toml")
			} else {
				fmt.Println("All chains in settings.toml:")
				for chain := range config.Chains {
					fmt.Println(chain)
				}
			}
			os.Exit(1)
		}
		chainURL = chainConfig.URL
	} else {
		fmt.Println("Provide either url, chain, or set a default chain in settings.toml.")
		os.Exit(1)
	}

	fmt.Printf("Connecting to chain at %s\n", chainURL)
	client, err := ethclient.Dial(chainURL)
	if err != nil {
		fmt.Printf("Error connecting to chain at %s: %s", chainURL, err.Error())
		os.Exit(1)
	}
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		fmt.Printf("Error getting chain ID: %s", err.Error())
		os.Exit(1)
	}

	app, err := view.NewApp(chainURL, chainID)
	if err != nil {
		fmt.Printf("Error creating app: %s", err.Error())
		os.Exit(1)
	}
	if err := app.Run(); err != nil {
		fmt.Printf("Error running app: %s", err.Error())
		os.Exit(1)
	}
}
