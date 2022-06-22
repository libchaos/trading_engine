module example

go 1.15

replace github.com/yzimhao/trading_engine => ../../trading_engine

require (
	github.com/gin-gonic/gin v1.8.1
	github.com/go-playground/validator/v10 v10.11.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/pelletier/go-toml/v2 v2.0.2 // indirect
	github.com/shopspring/decimal v1.3.1
	github.com/yzimhao/trading_engine v0.0.3
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/net v0.0.0-20220621193019-9d032be2e588 // indirect
	golang.org/x/sys v0.0.0-20220615213510-4f61da869c0c // indirect
)
