package jsonrpc

import (
	"github.com/essensio_network/common"
	"github.com/essensio_network/core"
	"github.com/essensio_network/core/txpool"
	"github.com/essensio_network/network"
	"net/http"

	"github.com/essensio_network/core/chainmgr"
)

type API struct {
	chain   *chainmgr.ChainManager
	txpool  txpool.TxnPool
	network *network.Server
}

func NewAPI(chain *chainmgr.ChainManager, txPool txpool.TxnPool) *API {

	return &API{chain: chain, txpool: txPool}
}

func (api *API) AddTransaction(req *http.Request, args *SendTransactionArgs, resp *Response) error {

	tx := &core.Transaction{
		Value: uint64(args.Value),
		Nonce: uint64(args.Nonce),
		From:  common.Address(args.From),
		To:    common.Address(args.To),
	}

	api.txpool.Insert(tx)
	api.txpool.BroadcastTx(tx)

	return nil
}

func (api *API) Stop() {
	api.chain.Stop()
}
