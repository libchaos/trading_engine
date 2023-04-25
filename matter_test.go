package trading_engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

var bids = NewMatter("北大街", 2, 0, time.Second*1)

func TestBid(t *testing.T) {
	bids.CleanAll()
	price := 1000.0
	for i := 0; i < 10; i++ {
		price += 1.0
		bids.PushNewOrder(NewBidLimitItem(uuid.NewString(), d(price), d(1000), time.Now().Unix()))
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println(bids.OrderQueue.pq)
	for _, item := range *bids.OrderQueue.pq {
		fmt.Println(item.(*BidItem).Order.orderId, item.(*BidItem).Order.price)
	}
	fmt.Printf("latest price: %#v\n", bids.GetTop())

}

func TestAsk(t *testing.T) {
	bids.CleanAll()
	price := 1000.0
	for i := 0; i < 10; i++ {
		price += 1.0
		bids.PushNewOrder(NewAskLimitItem(uuid.NewString(), d(price), d(1000), time.Now().Unix()))
		fmt.Println("latest price: ", bids.GetLatestPrice())
		time.Sleep(10 * time.Millisecond)
	}
}

func TestSendRandomResult(t *testing.T) {

	go func() {
		bids.CleanAll()
		price := 1000.0
		for i := 0; i < 10; i++ {
			price += 1.0
			bids.PushNewOrder(NewAskLimitItem(uuid.NewString(), d(price), d(1000), time.Now().Unix()))
			fmt.Println("latest price: ", bids.GetLatestPrice())
			time.Sleep(10 * time.Millisecond)
		}

	}()

	go func() {

		<-time.After(bids.Duration)
		bids.ChCancelResult <- struct{}{}

	}()

	select {
	case <-bids.ChCancelResult:
		go bids.SendMatterResultNotify(bids.GetRandomOne())
		bids.CleanAll()
		fmt.Println(<-bids.ChTradeResult)

	case <-time.After(2 * time.Second):
		fmt.Println("time out...")

	}

}

func TestSendTopResult(t *testing.T) {

	go func() {
		bids.CleanAll()
		price := 1000.0
		for i := 0; i < 10; i++ {
			price += 1.0
			bids.PushNewOrder(NewAskLimitItem(uuid.NewString(), d(price), d(1000), time.Now().Unix()))
			fmt.Println("latest price: ", bids.GetLatestPrice())
			time.Sleep(10 * time.Millisecond)
		}

	}()

	go func() {

		<-time.After(bids.Duration)
		bids.ChCancelResult <- struct{}{}

	}()

	select {
	case <-bids.ChCancelResult:
		go bids.SendMatterResultNotify(bids.GetTop())
		bids.CleanAll()
		fmt.Println(<-bids.ChTradeResult)

	case <-time.After(2 * time.Second):
		fmt.Println("time out...")

	}

}
