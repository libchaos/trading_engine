package trading_engine

import (
	"container/heap"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type MatterQueue struct {
	pq *PriorityQueue
	sync.Mutex
	m map[string]*QueueItem
}

func (o *MatterQueue) Len() int {
	return o.pq.Len()
}

func (o *MatterQueue) Push(item QueueItem) (exist bool) {
	o.Lock()
	defer o.Unlock()

	id := item.GetUniqueId()
	if _, ok := o.m[id]; ok {
		return true
	}

	heap.Push(o.pq, item)
	o.m[id] = &item
	return false
}

func (o *MatterQueue) Get(index int) QueueItem {
	n := o.pq.Len()
	if n <= index {
		return nil
	}

	return (*o.pq)[index]
}

func (o *MatterQueue) Top() QueueItem {
	return o.Get(0)
}

func (o *MatterQueue) Remove(uniqId string) QueueItem {
	o.Lock()
	defer o.Unlock()

	old, ok := o.m[uniqId]
	if !ok {
		return nil
	}

	item := heap.Remove(o.pq, (*old).GetIndex())
	delete(o.m, uniqId)
	return item.(QueueItem)
}

func (o *MatterQueue) clean() {
	o.Lock()
	defer o.Unlock()

	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	o.pq = &pq
	o.m = make(map[string]*QueueItem)
}

func NewBidQueue() *MatterQueue {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	queue := MatterQueue{
		pq: &pq,
		m:  make(map[string]*QueueItem),
	}
	return &queue
}

type MatterResult struct {
	Symbol        string          `json:"symbol"`
	OrderId       string          `json:"order_id"`
	TradeQuantity decimal.Decimal `json:"trade_quantity"`
	TradePrice    decimal.Decimal `json:"trade_price"`
	TradeAmount   decimal.Decimal `json:"trade_amount"`
	TradeTime     int64           `json:"trade_time"`
}

type Matter struct {
	Symbol         string
	ChTradeResult  chan MatterResult
	ChCancelResult chan struct{}

	priceDigit    int
	quantityDigit int
	latestPrice   decimal.Decimal
	Duration      time.Duration
	OrderQueue    *MatterQueue

	sync.Mutex
}

func (t *Matter) PushNewOrder(item QueueItem) {
	t.handlerNewOrder(item)
}

func (t *Matter) handlerNewOrder(newOrder QueueItem) {
	t.Lock()
	defer t.Unlock()

	t.OrderQueue.Push(newOrder)

}

func (t *Matter) CleanAll() {
	t.Lock()
	defer t.Unlock()

	t.OrderQueue.clean()
}

func (t *Matter) GetLatestPrice() decimal.Decimal {
	return t.OrderQueue.Top().GetPrice()
}

func (t *Matter) GetTop() QueueItem {
	return t.OrderQueue.Top()
}

func (t *Matter) SendMatterResultNotify(item QueueItem) {
	tradelog := MatterResult{}
	tradelog.Symbol = t.Symbol
	tradelog.OrderId = item.GetUniqueId()
	tradelog.TradeQuantity = item.GetQuantity()
	tradelog.TradePrice = item.GetPrice()
	tradelog.TradeTime = time.Now().UnixNano()
	tradelog.TradeAmount = tradelog.TradeQuantity.Mul(tradelog.TradePrice)

	t.latestPrice = tradelog.TradePrice

	if Debug {
		logrus.Infof("%s tradelog: %+v", t.Symbol, tradelog)
	}

	t.ChTradeResult <- tradelog
}

func NewMatter(symbol string, priceDigit, quantityDigit int, timeToCancel time.Duration) *Matter {
	t := &Matter{
		Symbol:         symbol,
		ChTradeResult:  make(chan MatterResult),
		ChCancelResult: make(chan struct{}),
		Duration:       timeToCancel,

		priceDigit:    priceDigit,
		quantityDigit: quantityDigit,

		OrderQueue: NewBidQueue(),
	}

	return t
}
