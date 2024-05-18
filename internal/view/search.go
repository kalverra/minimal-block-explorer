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

func Search(app *App) error {
	if app.fullTopView.GetItemCount() > 1 { // prevents multiple search inputs
		return nil
	}

	searchInput := tview.NewInputField().
		SetLabel("/ ").
		SetFieldWidth(0)
	searchInput.SetBorder(true).SetBackgroundColor(tcell.ColorDefault)

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
				app.ContextView.AddError(fmt.Sprintf("Error connecting to chain at %s: %s", app.chainURL, err.Error()))
				return
			}
			newView, err := searchChain(app, client, searchInput.GetText())
			if err != nil {
				app.ContextView.AddError(err.Error())
			} else {
				app.Update(newView)
			}
		}
	})
	return nil
}

func searchChain(app *App, client *ethclient.Client, searchText string) (View, error) {
	searchType, err := determineSearchType(searchText)
	if err != nil {
		return nil, err
	}

	switch searchType {
	case BlockNumber:
		blockNumber, err := strconv.ParseInt(searchText, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("input '%s' is not a valid block number", searchText)
		}
		return NewBlock(app, client, blockNumber)
	case Hash:
		hash := common.HexToHash(searchText)
		block, err := client.BlockByHash(context.Background(), hash)
		if err == nil {
			return NewBlock(app, client, block.Number().Uint64())
		}
		// TODO: Search for transactions
	}
	return nil, fmt.Errorf("unknown search type for '%s'", searchText)
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
