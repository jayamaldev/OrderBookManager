// order book handler manages the order book using the snapshot and the market depth updates sent by the binance

package orderbook

import (
	"encoding/json"
	"fmt"
	"log"
	"order-book-manager/dtos"
	"strconv"
	"sync"

	tree "github.com/emirpasic/gods/trees/redblacktree"
)

var mutex sync.Mutex

type OrderBook struct {
	Bids *tree.Tree
	Asks *tree.Tree
}

var orderBookMap map[string]OrderBook

func InitOrderBook() {
	orderBookMap = make(map[string]OrderBook)
}

func InitOrderBookForCurrency(currPair string) {
	orderBookMap[currPair] = OrderBook{
		Bids: tree.NewWith(bidComparator),
		Asks: tree.NewWith(askComparator),
	}
}

func RemoveOrderBookForCurrency(currPair string) {
	delete(orderBookMap, currPair)
}

// populate order book for the given currency using the snapshot
func PopulateOrderBook(currPair string, snapshot *dtos.Snapshot) {
	mutex.Lock()
	defer mutex.Unlock()

	orderBook := orderBookMap[currPair]

	// process bids
	for _, bidEntry := range snapshot.Bids {
		priceVal, err := strconv.ParseFloat(bidEntry[0], 64)
		if err != nil {
			log.Println("Error on Parsing Bid Entry Price")
		}

		qtyVal, err := strconv.ParseFloat(bidEntry[1], 64)
		if err != nil {
			log.Println("Error on Parsing Bid Entry Qty")
		}

		orderBook.Bids.Put(priceVal, qtyVal)
	}

	// process asks
	for _, askEntry := range snapshot.Asks {
		priceVal, err := strconv.ParseFloat(askEntry[0], 64)
		if err != nil {
			log.Println("Error on Parsing Ask Entry Price")
		}

		qtyVal, err := strconv.ParseFloat(askEntry[1], 64)
		if err != nil {
			log.Println("Error on Parsing Ask Entry Qty")
		}

		orderBook.Asks.Put(priceVal, qtyVal)
	}

	PrintBids(currPair)
	PrintAsks(currPair)
}

func UpdateBids(currPair string, price, qty float64) {
	mutex.Lock()
	defer mutex.Unlock()

	orderBook, OK := orderBookMap[currPair]
	if OK {
		orderBook.Bids.Put(price, qty)
	}
}

func UpdateAsks(currPair string, price, qty float64) {
	mutex.Lock()
	defer mutex.Unlock()

	orderBook, OK := orderBookMap[currPair]
	if OK {
		orderBook.Asks.Put(price, qty)
	}
}

func PrintBids(currPair string) {
	bidArr, _ := orderBookMap[currPair].Bids.ToJSON()
	fmt.Println(string(bidArr))
	fmt.Println()
}

func PrintAsks(currPair string) {
	askArr, _ := orderBookMap[currPair].Asks.ToJSON()
	fmt.Println(string(askArr))
	fmt.Println()
}

func askComparator(a, b interface{}) int {
	aFloat := a.(float64)
	bFloat := b.(float64)

	if aFloat < bFloat {
		return -1
	}
	if aFloat > bFloat {
		return 1
	}
	return 0
}

func bidComparator(a, b interface{}) int {
	aFloat := a.(float64)
	bFloat := b.(float64)

	if aFloat < bFloat {
		return 1
	}
	if aFloat > bFloat {
		return -1
	}
	return 0
}

// marshal the order book to a json string to sent to the subscribed client
func GetOrderBook(currPair string) []byte {
	mutex.Lock()
	defer mutex.Unlock()

	orderBook := orderBookMap[currPair]
	jsonStr, err := json.Marshal(orderBook)
	if err != nil {
		log.Println("error on parsing order book to json", err)
	}

	return jsonStr
}
