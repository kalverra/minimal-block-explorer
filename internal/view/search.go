package view

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (app *App) Search() {
	if app.fullTopView.GetItemCount() > 1 { // prevents multiple search inputs
		return
	}

	searchInput := tview.NewInputField().
		SetLabel("/ ").
		SetFieldWidth(0)

	// Prevent the search input from capturing the '/' key
	searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == '/' {
			return nil
		}
		return event
	})

	app.fullTopView.AddItem(searchInput, 0, 1, true)
	app.SetFocus(searchInput)

	searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			defer app.fullTopView.RemoveItem(searchInput)

			client, err := ethclient.Dial(app.chainURL)
			if err != nil {
				app.UpdateContext((fmt.Sprintf("[red]Error connecting to chain at %s: %s[-]", app.chainURL, err.Error())))
				return
			}
			err = app.searchChain(client, searchInput.GetText())
			if err != nil {
				app.UpdateContext((fmt.Sprintf("[red]Error searching chain: %s[-]", err.Error())))
			}
		}
	})
}

func (app *App) searchChain(client *ethclient.Client, searchText string) error {
	searchType, err := determineSearchType(searchText)
	if err != nil {
		return err
	}

	switch searchType {
	case BlockNumber:
		blockNumber, err := strconv.ParseInt(searchText, 10, 64)
		if err != nil {
			return fmt.Errorf("input '%s' is not a valid block number", searchText)
		}
		app.ShowBlock(client, blockNumber)
		return nil
	case Hash:
		hash := common.HexToHash(searchText)
		_, err := client.BlockByHash(context.Background(), hash)
		if err == nil {
			app.ShowBlock(client, hash)
		} else {
			app.ShowTransaction(client, searchText)
		}
	}
	return fmt.Errorf("unknown search type for '%s'", searchText)
}

type SearchType string

const (
	BlockNumber SearchType = "Block Number"
	Hash        SearchType = "Hash"
)

func determineSearchType(searchText string) (SearchType, error) {
	searchText = strings.TrimPrefix(searchText, "0x")
	// Regular expression for a hexadecimal string
	hexRegex := regexp.MustCompile("^[a-fA-F0-9]+$")

	// Check if the input is an integer
	if _, err := strconv.Atoi(searchText); err == nil {
		return BlockNumber, nil
	}

	// Check if the input is a valid hexadecimal string
	if hexRegex.MatchString(searchText) {
		if len(searchText) == 64 {
			return Hash, nil
		}
		return "", fmt.Errorf("input '%s' has invalid hash length: %d", searchText, len(searchText))
	}

	return "", fmt.Errorf("unknown search type for '%s'", searchText)
}
