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

func NewBlock(app *App, client *ethclient.Client, numberOrHash any) (*Block, error) {
	block := &Block{
		number: big.NewInt(0),
		hash:   "",
	}
	switch numberOrHash := numberOrHash.(type) {
	case int:
		block.number.SetInt64(int64(numberOrHash))
	case int64:
		block.number.SetInt64(numberOrHash)
	case uint:
		block.number.SetUint64(uint64(numberOrHash))
	case uint64:
		block.number.SetUint64(numberOrHash)
	case string:
		block.hash = numberOrHash
	default:
		return nil, fmt.Errorf("'%v' is not a valid block number or hash", numberOrHash)
	}

	var (
		ethBlock *types.Block
		err      error
	)
	if block.hash == "" {
		ethBlock, err = client.BlockByNumber(context.Background(), block.number)
	} else {
		ethBlock, err = client.BlockByHash(context.Background(), common.HexToHash(block.hash))
	}
	block.Block = ethBlock
	if err != nil {
		return nil, fmt.Errorf("error getting block '%v': %s", numberOrHash, err.Error())
	}
	return block, err
}

type Block struct {
	*types.Block

	number *big.Int
	hash   string
}

func (b *Block) Show() tview.Primitive {
	return tview.NewBox().SetBorder(true).SetTitle(fmt.Sprintf("Block %d | %s", b.Number().Uint64(), b.Hash().Hex()))
}

func (b *Block) Controls() ControlMapping {
	return ControlMapping{}
}

func (b *Block) End() {

}
