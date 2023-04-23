package trading_engine

import (
	"github.com/shopspring/decimal"
)

type Order struct {
	orderId    string
	price      decimal.Decimal
	quantity   decimal.Decimal
	createTime int64
	index      int

	priceType PriceType
	amount    decimal.Decimal
}

func (o *Order) GetIndex() int {
	return o.index
}

func (o *Order) SetIndex(index int) {
	o.index = index
}

func (o *Order) SetQuantity(qnt decimal.Decimal) {
	o.quantity = qnt
}

func (o *Order) SetAmount(amount decimal.Decimal) {
	o.amount = amount
}

func (o *Order) GetUniqueId() string {
	return o.orderId
}

func (o *Order) GetPrice() decimal.Decimal {
	return o.price
}

func (o *Order) GetQuantity() decimal.Decimal {
	return o.quantity
}

func (o *Order) GetCreateTime() int64 {
	return o.createTime
}

func (o *Order) GetPriceType() PriceType {
	return o.priceType
}
func (o *Order) GetAmount() decimal.Decimal {
	return o.amount
}

// 这个方法留在具体的 ask/bid 队列中实现
// func (o *Order) Less() {}

type AskItem struct {
	Order
}

func (a *AskItem) Less(other QueueItem) bool {
	//价格优先，时间优先原则
	//价格低的在最上面
	return (a.price.Cmp(other.(*AskItem).price) == -1) || (a.price.Cmp(other.(*AskItem).price) == 0 && a.createTime < other.(*AskItem).createTime)
}

func (a *AskItem) GetOrderSide() OrderSide {
	return OrderSideSell
}

type BidItem struct {
	Order
}

func (a *BidItem) Less(other QueueItem) bool {
	//价格优先，时间优先原则
	//价格高的在最上面
	return (a.price.Cmp(other.(*BidItem).price) == 1) || (a.price.Cmp(other.(*BidItem).price) == 0 && a.createTime < other.(*BidItem).createTime)
}

func (a *BidItem) GetOrderSide() OrderSide {
	return OrderSideBuy
}

func NewAskItem(pt PriceType, uniqId string, price, quantity, amount decimal.Decimal, createTime int64) *AskItem {
	return &AskItem{
		Order: Order{
			orderId:    uniqId,
			price:      price,
			quantity:   quantity,
			createTime: createTime,
			priceType:  pt,
			amount:     amount,
		},
	}
}

// 限价卖单
func NewAskLimitItem(uniq string, price, quantity decimal.Decimal, createTime int64) *AskItem {
	return NewAskItem(PriceTypeLimit, uniq, price, quantity, decimal.Zero, createTime)
}

// 市价-按数量卖出
func NewAskMarketQtyItem(uniq string, quantity decimal.Decimal, createTime int64) *AskItem {
	return NewAskItem(PriceTypeMarketQuantity, uniq, decimal.Zero, quantity, decimal.Zero, createTime)
}

// 市价 按金额卖出订单时，需要用户持有交易物的数量，在撮合时候防止超卖
func NewAskMarketAmountItem(uniq string, amount, maxHoldQty decimal.Decimal, createTime int64) *AskItem {
	return NewAskItem(PriceTypeMarketAmount, uniq, decimal.Zero, maxHoldQty, amount, createTime)
}

func NewBidItem(pt PriceType, uniqId string, price, quantity, amount decimal.Decimal, createTime int64) *BidItem {
	return &BidItem{
		Order: Order{
			orderId:    uniqId,
			price:      price,
			quantity:   quantity,
			createTime: createTime,
			priceType:  pt,
			amount:     amount,
		}}
}

// 限价买单
func NewBidLimitItem(uniq string, price, quantity decimal.Decimal, createTime int64) *BidItem {
	return NewBidItem(PriceTypeLimit, uniq, price, quantity, decimal.Zero, createTime)
}

// 市价-按数量买单,需要用户可用资金来限制最大买入量
func NewBidMarketQtyItem(uniq string, quantity, maxAmount decimal.Decimal, createTime int64) *BidItem {
	return NewBidItem(PriceTypeMarketQuantity, uniq, decimal.Zero, quantity, maxAmount, createTime)
}

// 市价-按金额买单
func NewBidMarketAmountItem(uniq string, amount decimal.Decimal, createTime int64) *BidItem {
	return NewBidItem(PriceTypeMarketAmount, uniq, decimal.Zero, decimal.Zero, amount, createTime)
}
