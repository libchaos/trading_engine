## 交易系统撮合引擎
<p align="left">
    <img src="https://img.shields.io/github/stars/yzimhao/trading_engine?style=social">
    <img src="https://img.shields.io/github/forks/yzimhao/trading_engine?style=social">
	<img src="https://img.shields.io/github/issues/yzimhao/trading_engine">
	<img src="https://img.shields.io/github/repo-size/yzimhao/trading_engine">
	<img src="https://img.shields.io/github/license/yzimhao/trading_engine">
</p>


## 适用场景
  买卖双方自由报价，需要按照价格优先、时间优先的原则撮合成交，如：证券交易、虚拟货币交易等。

## 支持的功能
* 限价委托 -- 用户指定一个价格，只有当撮合引擎找到同样价格或者最优的价格才会执行交易
* 市价委托 -- 市价委托会忽略价格因素，最大限度的完成指定数量或者金额的成交，市价单优先级最高，流动性充足的市场，可以保证成交，流动性不足时，剩余未能成交的部分会撤单。
  * 市价 按数量
  * 市价 按成交金额
* 取消订单
* 委托深度
* 最新成交价格

 功能体验 => [http://132.226.14.192:8080/demo](http://132.226.14.192:8080/demo)

## 接入方式
 1. `go package`
 ```
 go get github.com/libchaos/trading_engine
 ```
 具体详细使用方法参考 [Readme](https://github.com/yzimhao/trading_engine#readme)


  2. 独立程序
 使用消息中间件接入，开发准备中...
 ```
 ...
 ```

### Support or Contact

[需求建议](https://github.com/yzimhao/trading_engine#%E9%9C%80%E6%B1%82%E8%AE%A8%E8%AE%BA%E8%81%94%E7%B3%BB)
