package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	te "github.com/yzimhao/trading_engine"
)

type Order struct {
	OrderId  string `json:"order_id"`
	Side     string `json:"side"` // buy、sell
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
	Amount   string `json:"amount"`

	MaxHoldAmount string `json:"max_hold_amount"` //扣除交易产生的手续费之后剩余的最大资金量
	MaxHoldQty    string `json:"max_hold_qty"`    //最大持有的交易物数量

	CreateTime int64 `json:"create_time"` //精确到纳秒
}

type NewOrderMsgBody struct {
	PriceType string `json:"price_type"` //limit、market_qty、market_amount
	Order     Order  `json:"order"`
}

type CancelOrderMsgBody struct {
	Side    string `json:"side"`
	OrderId string `json:"order_id"`
}

func getNewOrder(pair *te.TradePair) {
	ctx := context.Background()
	sub := rdc.Subscribe(ctx, fmt.Sprintf("push_new_order.%s", pair.Symbol))
	defer sub.Close()
	for {
		msg, err := sub.ReceiveMessage(ctx)
		if err != nil {
			logrus.Errorf("sub.ReceiveMessage error: %v", err)
			continue
		}
		logrus.Debugf("sub.ReceiveMessage: %v", msg)
		if msg.Channel == fmt.Sprintf("push_new_order.%s", pair.Symbol) {
			var item NewOrderMsgBody

			err := json.Unmarshal([]byte(msg.Payload), &item)
			if err != nil {
				logrus.Errorf("json.Unmarshal error: %v payload: %s", err, msg.Payload)
				continue
			}

			order := item.Order
			price_type := strings.ToLower(item.PriceType)
			order_type := strings.ToLower(order.Side)
			if price_type == "limit" {
				if order_type == "buy" {
					pair.ChNewOrder <- te.NewBidLimitItem(order.OrderId, str2decimal(order.Price), str2decimal(order.Quantity), order.CreateTime)
				} else if order_type == "sell" {
					pair.ChNewOrder <- te.NewAskLimitItem(order.OrderId, str2decimal(order.Price), str2decimal(order.Quantity), order.CreateTime)
				}
			} else if price_type == "market_qty" {
				if order_type == "buy" {
					maxHoldAmount := str2decimal(order.MaxHoldAmount)
					pair.ChNewOrder <- te.NewBidMarketQtyItem(order.OrderId, str2decimal(order.Quantity), maxHoldAmount, order.CreateTime)
				} else if order_type == "sell" {
					pair.ChNewOrder <- te.NewAskMarketQtyItem(order.OrderId, str2decimal(order.Quantity), order.CreateTime)
				}
			} else if price_type == "market_amount" {
				if order_type == "buy" {
					pair.ChNewOrder <- te.NewBidMarketAmountItem(order.OrderId, str2decimal(order.Amount), order.CreateTime)
				} else if order_type == "sell" {
					maxHoldQty := str2decimal(order.MaxHoldQty)
					pair.ChNewOrder <- te.NewAskMarketAmountItem(order.OrderId, str2decimal(order.Amount), maxHoldQty, order.CreateTime)
				}
			}
		}
	}
}

func cancelOrder(pair *te.TradePair) {
	ctx := context.Background()

	sub := rdc.Subscribe(ctx, fmt.Sprintf("cancel_order.%s", pair.Symbol))
	defer sub.Close()
	for {
		msg, err := sub.ReceiveMessage(ctx)
		if err != nil {
			logrus.Errorf("sub.ReceiveMessage error: %v", err)
			continue
		}
		logrus.Debugf("sub.ReceiveMessage: %v", msg)

		if msg.Channel == fmt.Sprintf("cancel_order.%s", pair.Symbol) {
			var item CancelOrderMsgBody

			err := json.Unmarshal([]byte(msg.Payload), &item)
			if err != nil {
				logrus.Errorf("json.Unmarshal error: %v payload: %s", err, msg.Payload)
				continue
			}

			if item.Side == "buy" {
				pair.CancelOrder(te.OrderSideBuy, item.OrderId)
			} else if item.Side == "sell" {
				pair.CancelOrder(te.OrderSideSell, item.OrderId)
			}
		}
	}
}

func publishMsg(pair *te.TradePair) {
	//
	config.SetDefault("kline.redis.trade_log_subscribe_key", "list:trade_log")

	ctx := context.Background()
	for {
		select {
		case log, ok := <-pair.ChTradeResult:
			if ok {
				msg, _ := json.Marshal(log)
				klinerdc.LPush(ctx, config.GetString("kline.redis.trade_log_subscribe_key"), msg)
			}
		case cancelOrderId := <-pair.ChCancelResult:
			rdc.LPush(ctx, fmt.Sprintf("cancel_result.%s", pair.Symbol), cancelOrderId)

		default:
			time.Sleep(time.Duration(100) * time.Millisecond)
		}

	}
}
