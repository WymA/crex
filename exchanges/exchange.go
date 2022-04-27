package exchanges

import (
	"fmt"

	. "github.com/WymA/crex"
	"github.com/WymA/crex/exchanges/binancefutures"
	"github.com/WymA/crex/exchanges/bitmex"
	"github.com/WymA/crex/exchanges/bybit"
	"github.com/WymA/crex/exchanges/deribit"
	"github.com/WymA/crex/exchanges/hbdm"
	"github.com/WymA/crex/exchanges/hbdmswap"
	"github.com/WymA/crex/exchanges/okexfutures"
	"github.com/WymA/crex/exchanges/okexswap"
)

func NewExchange(name string, opts ...ApiOption) Exchange {
	params := &Parameters{}

	for _, opt := range opts {
		opt(params)
	}

	return NewExchangeFromParameters(name, params)
}

func NewExchangeFromParameters(name string, params *Parameters) Exchange {
	switch name {
	case BinanceFutures:
		return binancefutures.NewBinanceFutures(params)
	case BitMEX:
		return bitmex.NewBitMEX(params)
	case Deribit:
		return deribit.NewDeribit(params)
	case Bybit:
		return bybit.NewBybit(params)
	case Hbdm:
		return hbdm.NewHbdm(params)
	case HbdmSwap:
		return hbdmswap.NewHbdmSwap(params)
	case OkexFutures:
		return okexfutures.NewOkexFutures(params)
	case OkexSwap:
		return okexswap.NewOkexSwap(params)
	default:
		panic(fmt.Sprintf("new exchange error [%v]", name))
	}
}
