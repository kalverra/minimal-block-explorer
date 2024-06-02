package view

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (app *App) ShowBlock(client *ethclient.Client, numberOrHash any) {
	app.UpdateContext("[yellow]Loading block...[-]")
	number := big.NewInt(0)
	hash := ""
	switch numberOrHash := numberOrHash.(type) {
	case int:
		number.SetInt64(int64(numberOrHash))
	case int64:
		number.SetInt64(numberOrHash)
	case uint:
		number.SetUint64(uint64(numberOrHash))
	case uint64:
		number.SetUint64(numberOrHash)
	case string:
		hash = numberOrHash
	case common.Hash:
		hash = numberOrHash.Hex()
	default:
		app.UpdateContext(fmt.Sprintf("[red]'%v' is not a valid block number or hash[-]", numberOrHash))
		return
	}

	var (
		block    *types.Block
		err      error
		controls = &ControlMapping{}
	)
	if hash == "" {
		block, err = client.BlockByNumber(context.Background(), number)
	} else {
		block, err = client.BlockByHash(context.Background(), common.HexToHash(hash))
	}
	if err != nil {
		app.UpdateContext(fmt.Sprintf("[red]Error getting block '%v': %s[-]", numberOrHash, err.Error()))
		return
	}
	blockTime := time.Unix(int64(block.Time()), 0)
	app.UpdateContext("")

	blockView := tview.NewFlex().SetDirection(tview.FlexRow)
	blockView.SetBorder(true).SetTitle(fmt.Sprintf("Block %d | %s", number.Uint64(), block.Hash().Hex()))
	blockView.AddItem(tview.NewTextView().SetText(fmt.Sprintf(
		"%s\nWithdrawals: %d\nGas Used: %d\nUncles: %d",
		blockTime.String(), block.Withdrawals().Len(), block.GasUsed(), len(block.Uncles()))),
		0, 1, false,
	)
	if block.Transactions().Len() > 0 {
		transactions := tview.NewTable().SetBorders(false)
		transactions.SetTitle(fmt.Sprintf("%d Transactions", block.Transactions().Len()))
		transactions.SetBorder(true)
		transactions.SetSelectable(true, false)
		// Header
		for i, tx := range block.Transactions() {
			transactions.SetCell(i, 0, tview.NewTableCell(tx.Hash().Hex()).SetAlign(tview.AlignCenter))
		}
		blockView.AddItem(transactions, 0, 4, false)

		controls.SpecialControls = SpecialKeyControls{
			tcell.KeyUp: Control{
				Key:         "Up",
				Description: "Scroll up tx list",
				Order:       0,
				Fn: func() {
					row, _ := transactions.GetSelection()
					if row > 0 {
						transactions.Select(row-1, 0)
					}
				},
			},
			tcell.KeyDown: Control{
				Key:         "Down",
				Description: "Scroll down tx list",
				Order:       1,
				Fn: func() {
					row, _ := transactions.GetSelection()
					if row < transactions.GetRowCount()-1 {
						transactions.Select(row+1, 0)
					}
				},
			},
			tcell.KeyEnter: Control{
				Key:         "Enter",
				Description: "Show transaction",
				Order:       2,
				Fn: func() {
					row, _ := transactions.GetSelection()
					app.ShowTransaction(client, block.Transactions()[row].Hash().Hex())
				},
			},
		}
	} else {
		transactions := tview.NewTextView().SetText("No transactions")
		blockView.AddItem(transactions, 0, 1, false)
	}

	app.Main.Clear()
	app.Main.AddItem(blockView, 0, 1, true)
	app.UpdateControls(controls)
}
