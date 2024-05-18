package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kalverra/minimal-block-explorer/internal/view"
)

func main() {
	var chainURL string
	flag.StringVar(&chainURL, "url", "", "URL of the blockchain to connect to.")
	flag.Parse()

	if chainURL == "" {
		fmt.Println("Please provide a URL to connect to with --url")
		os.Exit(1)
	}

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

	app, err := view.NewApp(chainURL, chainID.Uint64())
	if err != nil {
		fmt.Printf("Error creating app: %s", err.Error())
		os.Exit(1)
	}
	if err := app.Run(); err != nil {
		fmt.Printf("Error running app: %s", err.Error())
		os.Exit(1)
	}
}
