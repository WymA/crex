package okexfutures

import (
	"time"

	. "github.com/WymA/crex"
	"github.com/WymA/crex/utils"
	"github.com/chuckpreslar/emission"
	"github.com/frankrap/okex-api"
)

type FuturesWebSocket struct {
	ws      *okex.FuturesWS
	emitter *emission.Emitter
}

func (s *FuturesWebSocket) SubscribeTrades(market Market, callback func(trades []*Trade)) error {
	s.emitter.On(WSEventTrade, callback)
	s.ws.SubscribeTrade("trade_1", market.Symbol)
	return nil
}

func (s *FuturesWebSocket) SubscribeLevel2Snapshots(market Market, callback func(ob *OrderBook)) error {
	s.emitter.On(WSEventL2Snapshot, callback)
	s.ws.SubscribeDepthL2Tbt("depthL2_1", market.Symbol)
	return nil
}

func (s *FuturesWebSocket) SubscribeOrders(market Market, callback func(orders []*Order)) error {
	s.emitter.On(WSEventOrder, callback)
	s.ws.SubscribeOrder("order_1", market.Symbol)
	return nil
}

func (s *FuturesWebSocket) SubscribePositions(market Market, callback func(positions []*Position)) error {
	s.emitter.On(WSEventPosition, callback)
	s.ws.SubscribePosition("position_1", market.Symbol)
	return nil
}

func (s *FuturesWebSocket) depth20SnapshotCallback(obRaw *okex.OrderBook) {
	// log.Printf("depthCallback %#v", *depth)
	// ch: market.BTC_CQ.depth.step0
	ob := &OrderBook{
		Symbol: obRaw.InstrumentID,
		Time:   time.Now(),
		Asks:   nil,
		Bids:   nil,
	}
	for _, v := range obRaw.Asks {
		ob.Asks = append(ob.Asks, Item{
			Price:  v.Price,
			Amount: v.Amount,
		})
	}
	for _, v := range obRaw.Bids {
		ob.Bids = append(ob.Bids, Item{
			Price:  v.Price,
			Amount: v.Amount,
		})
	}
	s.emitter.Emit(WSEventL2Snapshot, ob)
}

func (s *FuturesWebSocket) tradeCallback(_trades []okex.WSTrade) {
	// log.Printf("tradeCallback")
	var result []*Trade
	for _, v := range _trades {
		var direction Direction
		if v.Side == "buy" {
			direction = Buy
		} else if v.Side == "sell" {
			direction = Sell
		}
		t := Trade{
			ID:        v.TradeID,
			Direction: direction,
			Price:     utils.ParseFloat64(v.Price),
			Amount:    utils.ParseFloat64(v.Side),
			Ts:        v.Timestamp.UnixNano() / int64(time.Millisecond),
			Symbol:    v.InstrumentID,
		}
		result = append(result, &t)
	}
	s.emitter.Emit(WSEventTrade, result)
}

func (s *FuturesWebSocket) ordersCallback(orders []okex.WSOrder) {
	//log.Printf("ordersCallback")
	var eventData []*Order
	for _, v := range orders {
		o := s.convertOrder(&v)
		eventData = append(eventData, o)
	}
	s.emitter.Emit(WSEventOrder, eventData)
}

func (s *FuturesWebSocket) convertOrder(order *okex.WSOrder) *Order {
	o := &Order{}
	o.ID = order.OrderID
	o.Symbol = order.InstrumentID
	o.Price = utils.ParseFloat64(order.Price)
	o.AvgPrice = utils.ParseFloat64(order.PriceAvg)
	// o.StopPx = 0
	o.Amount = utils.ParseFloat64(order.Size)
	o.FilledAmount = utils.ParseFloat64(order.FilledQty)
	switch order.Type {
	case "1":
		o.Direction = Buy
	case "2":
		o.Direction = Sell
	case "3":
		o.Direction = Sell
		o.ReduceOnly = true
	case "4":
		o.Direction = Buy
		o.ReduceOnly = true
	}
	/*
		0：普通委托
		1：只做Maker（Post only）
		2：全部成交或立即取消（FOK）
		3：立即成交并取消剩余（IOC）
		4：市价委托
	*/
	switch order.OrderType {
	case "0":
		o.Type = OrderTypeLimit
	case "1":
		o.Type = OrderTypeMarket
		o.PostOnly = true
	case "2":
		o.Type = OrderTypeLimit
	case "3":
		o.Type = OrderTypeLimit
	case "4":
		o.Type = OrderTypeMarket
	default:
		o.Type = OrderTypeLimit
	}
	/*
		-2:失败
		-1:撤单成功
		0:等待成交
		1:部分成交
		2:完全成交
		3:下单中
		4:撤单中
	*/
	switch order.State {
	case "-2":
		o.Status = OrderStatusRejected
	case "-1":
		o.Status = OrderStatusCancelled
	case "0":
		o.Status = OrderStatusNew
	case "1":
		o.Status = OrderStatusPartiallyFilled
	case "2":
		o.Status = OrderStatusFilled
	case "3":
		o.Status = OrderStatusCreated
	case "4":
		o.Status = OrderStatusCancelPending
	}
	return o
}

func (s *FuturesWebSocket) positionsCallback(positions []okex.WSFuturesPosition) {
	//log.Printf("positionsCallback")
	var eventData []*Position
	for _, v := range positions {
		longQty := utils.ParseFloat64(v.LongQty)
		shortQty := utils.ParseFloat64(v.ShortQty)
		if longQty > 0 {
			var o Position
			o.Symbol = v.InstrumentID
			o.OpenTime = v.Timestamp
			o.Size = longQty
			o.OpenPrice = utils.ParseFloat64(v.LongAvgCost)
			o.AvgPrice = o.OpenPrice
			eventData = append(eventData, &o)
		} else if shortQty > 0 {
			var o Position
			o.Symbol = v.InstrumentID
			o.OpenTime = v.Timestamp
			o.Size = -shortQty
			o.OpenPrice = utils.ParseFloat64(v.ShortAvgCost)
			o.AvgPrice = o.OpenPrice
			eventData = append(eventData, &o)
		}
	}
	s.emitter.Emit(WSEventPosition, eventData)
}

func NewFuturesWebSocket(params *Parameters) *FuturesWebSocket {
	wsURL := "wss://real.okex.com:8443/ws/v3"
	s := &FuturesWebSocket{
		emitter: emission.NewEmitter(),
	}
	ws := okex.NewFuturesWS(wsURL,
		params.AccessKey, params.SecretKey, params.Passphrase, params.DebugMode)
	if params.ProxyURL != "" {
		ws.SetProxy(params.ProxyURL)
	}
	ws.SetDepth20SnapshotCallback(s.depth20SnapshotCallback)
	ws.SetTradeCallback(s.tradeCallback)
	ws.SetOrderCallback(s.ordersCallback)
	ws.SetPositionCallback(s.positionsCallback)
	ws.Start()
	s.ws = ws
	return s
}
