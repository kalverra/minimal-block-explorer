package view

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rivo/tview"
)

func NewChainSummary(app *App, client *ethclient.Client) (*ChainSummary, error) {
	return &ChainSummary{
		app:    app,
		client: client,
		endCh:  make(chan struct{}),
	}, nil
}

type ChainSummary struct {
	app    *App
	client *ethclient.Client
	endCh  chan struct{}
}

func (c *ChainSummary) Show() tview.Primitive {
	// TODO: Enable live updates for the chain summary
	latestHeader, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		c.app.ContextView.AddError("Error getting latest header: " + err.Error())
		return nil
	}

	sum := tview.NewTextView().SetText(fmt.Sprintf("Latest block: %s | %s", latestHeader.Number.String(), latestHeader.Hash().Hex()))
	sum.SetBorder(true).SetTitle("Chain Summary")
	return sum
}
func (c *ChainSummary) Controls() ControlMapping {
	return ControlMapping{}
}

func (c *ChainSummary) End() {
	close(c.endCh)
}
