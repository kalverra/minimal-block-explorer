package view

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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
		block *types.Block
		err   error
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
	app.UpdateContext("")

	blockView := tview.NewFlex().SetDirection(tview.FlexRow)
	blockView.SetBorder(true).SetTitle(fmt.Sprintf("Block %d | %s", number.Uint64(), block.Hash().Hex()))
	blockView.AddItem(tview.NewTextView().SetText(fmt.Sprintf(
		"Time: %d\nWithdrawals: %d\nGas Used: %d\nUncles: %d",
		block.Time(), block.Withdrawals().Len(), block.GasUsed(), len(block.Uncles()))),
		0, 1, false,
	)
	if block.Transactions().Len() > 0 {
		transactions := tview.NewTable().SetBorders(true)
		transactions.SetTitle("Transactions")
		transactions.SetBorder(true)
		transactions.SetSelectable(true, false)
		transactions.SetSelectedFunc(func(row, _ int) {
			app.ShowTransaction(client, block.Transactions()[row].Hash().Hex())
		})
	} else {
		transactions := tview.NewTextView().SetText("No transactions")
		transactions.SetBorder(true).SetTitle("Transactions")
		blockView.AddItem(transactions, 0, 1, false)
	}

	app.Main.Clear()
	app.Main.AddItem(blockView, 0, 1, true)
	app.UpdateControls(nil)
}
