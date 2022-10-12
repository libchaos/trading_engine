package trading_engine

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type TradeResult struct {
	Symbol        string          `json:"symbol"`
	AskOrderId    string          `json:"ask_order_id"`
	BidOrderId    string          `json:"bid_order_id"`
	TradeQuantity decimal.Decimal `json:"trade_quantity"`
	TradePrice    decimal.Decimal `json:"trade_price"`
	TradeAmount   decimal.Decimal `json:"trade_amount"`
	TradeTime     int64           `json:"trade_time"`
	MarketOrder   string          `json:"market_order"` //市价订单标记，用于结算时候取消剩余未成交的部分
}

type TradePair struct {
	Symbol         string
	ChTradeResult  chan TradeResult
	ChNewOrder     chan QueueItem
	ChCancelResult chan string

	priceDigit    int
	quantityDigit int
	latestPrice   decimal.Decimal

	askQueue *OrderQueue
	bidQueue *OrderQueue

	w sync.Mutex
}

func NewTradePair(symbol string, priceDigit, quantityDigit int) *TradePair {
	t := &TradePair{
		Symbol:         symbol,
		ChTradeResult:  make(chan TradeResult, 10),
		ChNewOrder:     make(chan QueueItem),
		ChCancelResult: make(chan string, 10),

		priceDigit:    priceDigit,
		quantityDigit: quantityDigit,

		askQueue: NewQueue(),
		bidQueue: NewQueue(),
	}
	go t.depthTicker(t.askQueue)
	go t.depthTicker(t.bidQueue)
	go t.matching()
	return t
}

func (t *TradePair) PushNewOrder(item QueueItem) {
	// t.ChNewOrder <- item
	t.handlerNewOrder(item)
}

func (t *TradePair) CancelOrder(side OrderSide, uniq string) {
	//todo 最好根据订单编号知道是买单还是卖单，方便直接查找到相应的队列，从中删除
	if side == OrderSideSell {
		t.askQueue.Remove(uniq)
	} else {
		t.bidQueue.Remove(uniq)
	}
	//删除成功后需要发送通知
	t.ChCancelResult <- uniq
}

func (t *TradePair) AskLen() int {
	t.w.Lock()
	defer t.w.Unlock()

	return t.askQueue.Len()
}

func (t *TradePair) BidLen() int {
	t.w.Lock()
	defer t.w.Unlock()

	return t.bidQueue.Len()
}

func (t *TradePair) LatestPrice() decimal.Decimal {
	t.w.Lock()
	defer t.w.Unlock()

	return t.latestPrice
}

func (t *TradePair) cleanAll() {
	t.w.Lock()
	defer t.w.Unlock()

	//同时清空两个队列
	t.askQueue.clean()
	t.bidQueue.clean()
}

func (t *TradePair) matching() {

	for {
		select {
		case newOrder := <-t.ChNewOrder:
			go t.handlerNewOrder(newOrder)
		default:
			t.handlerLimitOrder()
		}

	}

}

func (t *TradePair) handlerNewOrder(newOrder QueueItem) {
	t.w.Lock()
	defer t.w.Unlock()

	if newOrder.GetPriceType() == PriceTypeLimit {
		if newOrder.GetOrderSide() == OrderSideSell {
			t.askQueue.Push(newOrder)
		} else {
			t.bidQueue.Push(newOrder)
		}
	} else {
		//市价单处理
		if newOrder.GetOrderSide() == OrderSideSell {
			t.doMarketSell(newOrder)
		} else {
			t.doMarketBuy(newOrder)
		}
	}

}

func (t *TradePair) handlerLimitOrder() {
	ok := func() bool {
		t.w.Lock()
		defer t.w.Unlock()

		if t.askQueue == nil || t.bidQueue == nil {
			return false
		}

		if t.askQueue.Len() == 0 || t.bidQueue.Len() == 0 {
			return false
		}

		askTop := t.askQueue.Top()
		bidTop := t.bidQueue.Top()

		defer func() {
			if askTop.GetQuantity().Equal(decimal.Zero) {
				t.askQueue.Remove(askTop.GetUniqueId())
			}
			if bidTop.GetQuantity().Equal(decimal.Zero) {
				t.bidQueue.Remove(bidTop.GetUniqueId())
			}
		}()

		if bidTop.GetPrice().Cmp(askTop.GetPrice()) >= 0 {
			curTradeQty := decimal.Zero
			curTradePrice := decimal.Zero
			if bidTop.GetQuantity().Cmp(askTop.GetQuantity()) >= 0 {
				curTradeQty = askTop.GetQuantity()
			} else if bidTop.GetQuantity().Cmp(askTop.GetQuantity()) == -1 {
				curTradeQty = bidTop.GetQuantity()
			}
			askTop.SetQuantity(askTop.GetQuantity().Sub(curTradeQty))
			bidTop.SetQuantity(bidTop.GetQuantity().Sub(curTradeQty))

			if askTop.GetCreateTime() >= bidTop.GetCreateTime() {
				curTradePrice = bidTop.GetPrice()
			} else {
				curTradePrice = askTop.GetPrice()
			}

			t.sendTradeResultNotify(askTop, bidTop, curTradePrice, curTradeQty, "")
			return true
		} else {
			return false
		}

	}()

	if !ok {
		time.Sleep(time.Duration(200) * time.Millisecond)
	} else {
		if Debug {
			time.Sleep(time.Second * time.Duration(1))
		}
	}
}

func (t *TradePair) doMarketBuy(item QueueItem) {

    maxQty := func(remainAmount, marketPrice, needQty decimal.Decimal){
        qty := remainAmount.Div(marketPrice)
		return decimal.Min(qty, needQty)
    }


	for {
		ok := func() bool {

			if t.askQueue.Len() == 0 {
				return false
			}

			ask := t.askQueue.Top()
			if item.GetPriceType() == PriceTypeMarketQuantity {
                maxTradeQty := maxQty(item.GetAmount(), ask.GetPrice(), item.GetQuantity())
				curTradeQty := decimal.Zero

				//市价按买入数量
				if maxTradeQty.Cmp(decimal.New(1, int32(-t.quantityDigit))) < 0 {
					return false
				}

				if ask.GetQuantity().Cmp(maxTradeQty) <= 0 {
					curTradeQty = ask.GetQuantity()
					t.askQueue.Remove(ask.GetUniqueId())
				} else {
					curTradeQty = maxTradeQty
					ask.SetQuantity(ask.GetQuantity().Sub(curTradeQty))
				}

				item.SetQuantity(item.GetQuantity().Sub(curTradeQty))
				item.SetAmount(item.GetAmount().Sub(curTradeQty.Mul(ask.GetPrice())))

				//检查本次循环撮合是否是该订单最后一次撮合
				//如果是则标记该市价订单已经完成了
                //结束的条件：
                // a.对面订单列表空了
                // b.已经达到了用户需要的数量 
                // c.剩余资金已经不能达到最小成交需求 todo
				if t.askQueue.Len() == 0 || item.GetQuantity() == decimal.Zero {
					t.sendTradeResultNotify(ask, item, ask.GetPrice(), curTradeQty, item.GetUniqueId())
				} else {
					t.sendTradeResultNotify(ask, item, ask.GetPrice(), curTradeQty, "")
				}

				return true
			} else if item.GetPriceType() == PriceTypeMarketAmount {
				//市价-按成交金额
				//成交金额不包含手续费，手续费应该由上层系统计算提前预留
				//撮合会针对这个金额最大限度的买入
				if ask.GetPrice().Cmp(decimal.Zero) <= 0 {
					return false
				}

				maxTradeQty := item.GetAmount().Div(ask.GetPrice())
				curTradeQty := decimal.Zero

				if maxTradeQty.Cmp(decimal.New(1, int32(-t.quantityDigit))) < 0 {
					return false
				}
				if ask.GetQuantity().Cmp(maxTradeQty) <= 0 {
					curTradeQty = ask.GetQuantity()
					t.askQueue.Remove(ask.GetUniqueId())
				} else {
					curTradeQty = maxTradeQty
					ask.SetQuantity(ask.GetQuantity().Sub(curTradeQty))
				}

				//部分成交了，需要更新这个单的剩余可成交金额，用于下一轮重新计算最大成交量
				item.SetAmount(item.GetAmount().Sub(curTradeQty.Mul(ask.GetPrice())))
				item.SetQuantity(item.GetQuantity().Add(curTradeQty))

				//检查本次循环撮合是否是该订单最后一次撮合
				//如果是则标记该市价订单已经完成了
				if t.askQueue.Len() == 0 || item.GetQuantity().Cmp(decimal.Zero) == 0 {
					t.sendTradeResultNotify(ask, item, ask.GetPrice(), curTradeQty, item.GetUniqueId())
				} else {
					t.sendTradeResultNotify(ask, item, ask.GetPrice(), curTradeQty, "")
				}
				return true
			}

			return false
		}()

		if !ok {
			break
		}

	}
}
func (t *TradePair) doMarketSell(item QueueItem) {

	for {
		ok := func() bool {

			if t.bidQueue.Len() == 0 {
				return false
			}

			bid := t.bidQueue.Top()
			if item.GetPriceType() == PriceTypeMarketQuantity {

				curTradeQuantity := decimal.Zero
				//市价按买入数量
				if item.GetQuantity().Equal(decimal.Zero) {
					return false
				}

				if bid.GetQuantity().Cmp(item.GetQuantity()) <= 0 {
					curTradeQuantity = bid.GetQuantity()
					defer t.bidQueue.Remove(bid.GetUniqueId())
				} else {
					curTradeQuantity = item.GetQuantity()
					bid.SetQuantity(bid.GetQuantity().Sub(curTradeQuantity))
				}

				t.sendTradeResultNotify(item, bid, bid.GetPrice(), curTradeQuantity)
				item.SetQuantity(item.GetQuantity().Sub(curTradeQuantity))


                if t.bidQueue.Len() == 0 || item.GetQuantity()
                
				return true
			} else if item.GetPriceType() == PriceTypeMarketAmount {
				//市价-按成交金额成交
				if bid.GetPrice().Cmp(decimal.Zero) <= 0 {
					return false
				}

				maxTradeQty := item.GetAmount().Div(bid.GetPrice()).Truncate(int32(t.quantityDigit))

				//还需要用户是否持有这么多资产来卖出的条件限制
				maxTradeQty = decimal.Min(maxTradeQty, item.GetQuantity())

				curTradeQty := decimal.Zero
				if maxTradeQty.Cmp(decimal.New(1, int32(-t.quantityDigit))) < 0 {
					return false
				}

				if bid.GetQuantity().Cmp(maxTradeQty) <= 0 {
					curTradeQty = bid.GetQuantity()
					defer t.bidQueue.Remove(bid.GetUniqueId())
				} else {
					curTradeQty = maxTradeQty
					bid.SetQuantity(bid.GetQuantity().Sub(curTradeQty))
				}

				t.sendTradeResultNotify(item, bid, bid.GetPrice(), curTradeQty)
				item.SetAmount(item.GetAmount().Sub(curTradeQty.Mul(bid.GetPrice())))

				//市价 按成交额卖出时，需要用户持有的资产数量来进行限制
				item.SetQuantity(item.GetQuantity().Sub(curTradeQty))

				return true
			}

			return false
		}()

		if !ok {
			t.ChCancelResult <- item.GetUniqueId()
			break
		}

	}
}

func (t *TradePair) sendTradeResultNotify(ask, bid QueueItem, price, tradeQty decimal.Decimal, market_done string) {
	tradelog := TradeResult{}
	tradelog.Symbol = t.Symbol
	tradelog.AskOrderId = ask.GetUniqueId()
	tradelog.BidOrderId = bid.GetUniqueId()
	tradelog.TradeQuantity = tradeQty
	tradelog.TradePrice = price
	tradelog.TradeTime = time.Now().UnixNano()
	tradelog.TradeAmount = tradeQty.Mul(price)
	tradelog.MarketOrder = market_done //标记市价订单已经完成，结算时候标记这个市价订单完成

	t.latestPrice = price

	if Debug {
		logrus.Infof("%s tradelog: %+v", t.Symbol, tradelog)
	}

	t.ChTradeResult <- tradelog
}
