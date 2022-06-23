package app

import (
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	te "github.com/yzimhao/trading_engine"
	"github.com/yzimhao/utilgo"
)

var (
	config   *viper.Viper
	rdc      *redis.Client
	klinerdc *redis.Client
)

func Start(configPath string) {
	config = utilgo.ViperInit(configPath)
	rdc = redis.NewClient(&redis.Options{
		Addr:     config.GetString("trading_engine.redis.host"),
		Password: config.GetString("trading_engine.redis.password"),
		DB:       config.GetInt("trading_engine.redis.db"),
	})

	klinerdc = redis.NewClient(&redis.Options{
		Addr:     config.GetString("kline.redis.host"),
		Password: config.GetString("kline.redis.password"),
		DB:       config.GetInt("kline.redis.db"),
	})

	symbol := "eurusd"
	price_digit := 5
	qty_digit := 2
	tradingEngineStart(symbol, price_digit, qty_digit)
}

func tradingEngineStart(symbol string, pdig, qdig int) {
	pair := te.NewTradePair(strings.ToLower(symbol), pdig, qdig)

	//new order
	go getNewOrder(pair)

	//cancel order
	go cancelOrder(pair)

	//todo depth

	//publish msg
	publishMsg(pair)
	logrus.Info("trading engine done.")
}

func str2decimal(a string) decimal.Decimal {
	d, _ := decimal.NewFromString(a)
	return d
}

func str2Int64(a string) int64 {
	i, _ := strconv.ParseInt(a, 10, 64)
	return i
}
