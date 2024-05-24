package view

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rivo/tview"
)

func (app *App) ShowChainSummary(client *ethclient.Client) {
	latestHeader, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		app.UpdateContext(fmt.Sprintf("[red]Error getting latest block: %s[-]", err.Error()))
		return
	}
	app.UpdateControls(&ControlMapping{
		NormalControls: map[rune]Control{
			'r': {
				Key:         "r",
				Description: "Refresh",
				Order:       4,
				Fn: func() {
					app.ShowChainSummary(client)
				},
			},
		},
	})

	sumFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	sum := tview.NewTextView().SetText(fmt.Sprintf("Latest block: %s | %s", latestHeader.Number.String(), latestHeader.Hash().Hex()))
	sum.SetBorder(true).SetTitle("Chain Summary")
	sumFlex.AddItem(sum, 0, 1, false)
	app.Main.Clear()
	app.Main.AddItem(sumFlex, 0, 1, true)
}
