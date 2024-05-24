package view

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rivo/tview"
)

func (app *App) ShowTransaction(client *ethclient.Client, hash string) {
	transaction, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(hash))
	if err != nil {
		app.UpdateContext(fmt.Sprintf("[red]Error getting transaction: %s[-]", err.Error()))
		return
	}
	app.UpdateControls(&ControlMapping{
		NormalControls: map[rune]Control{
			'r': {
				Key:         "r",
				Description: "Refresh",
				Order:       4,
				Fn: func() {
					app.ShowTransaction(client, hash)
				},
			},
		},
	})

	if isPending {
		app.UpdateContext(fmt.Sprintf("[yellow]Transaction %s is pending[-]", transaction.Hash().Hex()))
		return
	}
	transactionFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	transactionView := tview.NewTextView()
	transactionView.SetBorder(true).SetTitle(fmt.Sprintf("Transaction: %s", transaction.Hash().Hex()))
	transactionFlex.AddItem(transactionView, 0, 1, false)
	app.Main.Clear()
	app.Main.AddItem(transactionFlex, 0, 1, true)
}
