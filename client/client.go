/*
package client acts as the client for the binance server.
This calls the Binance API and manages the order book and market depth updates
client will manage subscriptions for multiple currency pairs
*/

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"order-book-manager/dtos"
	"order-book-manager/orderbook"
	"order-book-manager/users"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

var conn *websocket.Conn
var lastUpdateIds map[string]int

var mainCurrencyPair string
var updateIdChan chan int
var bufferedEvents chan dtos.EventUpdate
var listSubscriptions chan []string
var firstEntryMap map[string]bool
var uniqueReqId int
var mutex sync.Mutex
var listSubscReqId int

var snapshotURL = "https://api.binance.com/api/v3/depth?symbol=%s&limit=50"
var depthStr = "%s@depth"

const (
	subscribe, unsubscribe, listSubscriptionsConst = "SUBSCRIBE", "UNSUBSCRIBE", "LIST_SUBSCRIPTIONS"
	wssStream, binanceUrl, wsContextRoot           = "wss", "stream.binance.com:9443", "/ws"
	depthUpdateEvent                               = "depthUpdate"
)

// Subscription/Unsubscription Request to the Binance to Subscribe for a Currecy Pair
type SubscriptionRequest struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int      `json:"id"`
}

// Subscription List Request to get the Subscribed Currecy Pairs list from Binance
type ListSubscriptionRequest struct {
	Method string `json:"method"`
	Id     int    `json:"id"`
}

// Function to get Unique request ID to send with Binance requests
func getUniqueReqId() int {
	mutex.Lock()
	defer mutex.Unlock()

	uniqueReqId++
	return uniqueReqId
}

func SetCurrencyPair(currPair string) {
	mainCurrencyPair = currPair
	lastUpdateIds = make(map[string]int)
}

/*
Setting up initial websocket connectivity with Binance and subscribe to the default currency pair.
This starts the message reader as well
*/
func ConnectToWebSocket() error {
	u := url.URL{
		Scheme: wssStream,
		Host:   binanceUrl,
		Path:   wsContextRoot,
	}

	g, ctx := errgroup.WithContext(context.Background())
	fmt.Printf("connecting to websocket %s\n", u.String())

	firstEntryMap = make(map[string]bool)

	var err error
	conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Websocket connectivity issue", err)
		return err
	}
	fmt.Printf("Connected to websocket %s\n", u.String())

	done := make(chan struct{})
	updateIdChan = make(chan int)
	bufferedEvents = make(chan dtos.EventUpdate, 1000)

	defer close(bufferedEvents)
	defer close(updateIdChan)

	g.Go(func() error {
		defer close(done)
		return readAndProcessWSMessages()
	})

	err = SubscribeToCurrPair(ctx, mainCurrencyPair)
	if err != nil {
		return err
	}

	go updateEvents()
	<-done
	fmt.Println("Websocket Client Closed")

	if err := g.Wait(); err != nil {
		log.Println("Error on Websocket Client")
		return err
	}

	return nil
}

func readAndProcessWSMessages() error {
	for {
		var eventUpdate dtos.EventUpdate
		var subscriptionsList dtos.SubscriptionsList

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error on reading Websocket Message", err)
			return err
		}

		// market depth update
		err = json.Unmarshal(message, &eventUpdate)
		if err != nil {
			fmt.Println("Error Parsing Json", err)
			break
		}

		// subscription list response
		err = json.Unmarshal(message, &subscriptionsList)
		if err != nil {
			fmt.Println("Error Parsing Subscriptions List Json", err)
			break
		}

		// process only market depth updates
		processMarketDepthUpdate(eventUpdate, message)

		// process list of subscription responses
		processSubscriptionList(subscriptionsList, message)
	}
	return nil
}

// process only market depth updates
func processMarketDepthUpdate(eventUpdate dtos.EventUpdate, message []byte) {
	if eventUpdate.Symbol == "" {
		// not an market depth update message
		return
	}

	if firstEntryMap[eventUpdate.Symbol] {
		firstUpdateId := eventUpdate.FirstUpdateId
		fmt.Printf("first update Id %d for currency %s \n ", firstUpdateId, eventUpdate.Symbol)
		firstEntryMap[eventUpdate.Symbol] = false
		updateIdChan <- firstUpdateId
	}

	// write to bufferedEvents channel so the Event Processor goroutine will read from channel
	bufferedEvents <- eventUpdate

	// push the event to subscribed users
	users.PushEventToUsers(message, eventUpdate.Symbol)
}

// process list of subscription responses
func processSubscriptionList(subscriptionsList dtos.SubscriptionsList, message []byte) {
	if subscriptionsList.Id == 0 {
		// not an admin message
		return
	}
	fmt.Println("admin message received: ", string(message))
	if subscriptionsList.Id == listSubscReqId {
		fmt.Println("sending subs list to channel ", subscriptionsList.Result)
		listSubscriptions <- subscriptionsList.Result
	}
}

// Subscribe to Currency Pair
func SubscribeToCurrPair(ctx context.Context, currencyPair string) error {
	depthRequest := fmt.Sprintf(depthStr, strings.ToLower(currencyPair))
	subscriptionRequest := SubscriptionRequest{
		Method: subscribe,
		Params: []string{depthRequest},
		Id:     getUniqueReqId(),
	}

	fmt.Println("Subscription currency pair: ", currencyPair)
	orderbook.InitOrderBookForCurrency(currencyPair)
	firstEntryMap[currencyPair] = true

	subsRequest, err := json.Marshal(subscriptionRequest)
	fmt.Println("Subscription Request: ", string(subsRequest))
	if err != nil {
		fmt.Println("Error on parsing subscription request", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, subsRequest)
	if err != nil {
		fmt.Println("error on sending subscription request")
	}

	return getMarketDepth(ctx, currencyPair)
}

// UnSubscribe to Currency Pair
func UnsubscribeToCurrPair(currencyPair string) {
	depthRequest := fmt.Sprintf(depthStr, strings.ToLower(currencyPair))
	unsubscriptionRequest := SubscriptionRequest{
		Method: unsubscribe,
		Params: []string{depthRequest},
		Id:     getUniqueReqId(),
	}

	fmt.Println("Unsubscription currency pair: ", currencyPair)
	orderbook.RemoveOrderBookForCurrency(currencyPair)

	unsubsRequest, err := json.Marshal(unsubscriptionRequest)
	fmt.Println("UnSubscription Request: ", string(unsubsRequest))
	if err != nil {
		fmt.Println("Error on parsing unsubscription request", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, unsubsRequest)
	if err != nil {
		fmt.Println("error on sending unsubscription request")
	}
	fmt.Println("Unsubscribed for ", currencyPair)
}

// List of Subscribtions
func ListSubscriptions() []string {
	listSubscReqId = getUniqueReqId()
	listSubscriptions = make(chan []string)

	listSubscriptionRequest := ListSubscriptionRequest{
		Method: listSubscriptionsConst,
		Id:     listSubscReqId,
	}

	listsubsRequest, err := json.Marshal(listSubscriptionRequest)
	fmt.Println("List Subscription Request: ", string(listsubsRequest))
	if err != nil {
		fmt.Println("Error on parsing list subscription request", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, listsubsRequest)
	if err != nil {
		fmt.Println("error on sending unsubscription request")
	}

	subscriptionsList := <-listSubscriptions
	return subscriptionsList
}

func CloseConnection() {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error on writing close request to websocket")
	}
}

// get the market depth for a currency pair and populate the order book
func getMarketDepth(ctx context.Context, currPair string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(snapshotURL, currPair), nil)
	if err != nil {
		log.Printf("Error on Creating New GET Request for curr paid %s\n", currPair)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf(fmt.Sprintf("Error on Closing Response for curr pair %s", currPair), err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Fatal(fmt.Sprintf("Error on Getting Snapshot for curr pair %s", currPair), err)
		return err
	}

	var snapshot *dtos.Snapshot
	err = json.NewDecoder(resp.Body).Decode(&snapshot)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cound not parse Response Json for curr pair %s", currPair), err)
		return err
	}

	lastUpdateId := snapshot.LastUpdateId
	lastUpdateIds[currPair] = lastUpdateId

	firstUpdateId := <-updateIdChan
	fmt.Printf("last update id for currency pair %s : %d first update id %d \n", currPair, lastUpdateId, firstUpdateId)
	if lastUpdateId > firstUpdateId {
		fmt.Printf("Condition Satisfied for curr pair %s !! \n", currPair)
	} else {
		fmt.Printf("Closing the Application. Re-get snapshot for currency pair %s\n", currPair)
		return err
	}

	orderbook.PopulateOrderBook(currPair, snapshot)
	return nil
}

// event processor to process market depth updates
func updateEvents() {
	fmt.Println("Event Processor Started")
	for {
		eventUpdate := <-bufferedEvents
		currPair := eventUpdate.Symbol

		if eventUpdate.FinalUpdateId <= lastUpdateIds[currPair] {
			// not an eligible event to process
			continue
		}

		if eventUpdate.EventType != depthUpdateEvent {
			// Order Book Manager do not need to process these events
			fmt.Printf("Event type for curr pair %s %s \n", currPair, eventUpdate.EventType)
			continue
		}

		formattedText := fmt.Sprintf("processing event for curr pair: %s %d %d %s", currPair, eventUpdate.FirstUpdateId, eventUpdate.FinalUpdateId, time.Now())
		fmt.Println(formattedText)

		// process bids
		processBids(eventUpdate.Bids, currPair)

		// process asks
		processAsks(eventUpdate.Asks, currPair)

		// update last update id of the orderbook
		lastUpdateIds[currPair] = eventUpdate.FinalUpdateId
	}
}

// process bids and update order book from snapshot
func processBids(bids [][]string, currPair string) {
	for _, bidEntry := range bids {
		priceVal, err := strconv.ParseFloat(bidEntry[0], 64)
		if err != nil {
			log.Printf("Error on Parsing Bid Entry Price for curr pair %s \n", currPair)
		}

		qtyVal, err := strconv.ParseFloat(bidEntry[1], 64)
		if err != nil {
			log.Printf("Error on Parsing Bid Entry Quantity for curr pair %s \n", currPair)
		}

		orderbook.UpdateBids(currPair, priceVal, qtyVal)
	}
}

// process asks and update order book from snapshot
func processAsks(asks [][]string, currPair string) {
	for _, askEntry := range asks {
		priceVal, err := strconv.ParseFloat(askEntry[0], 64)
		if err != nil {
			log.Printf("Error on Parsing Ask Entry Price for curr pair %s \n", currPair)
		}

		qtyVal, err := strconv.ParseFloat(askEntry[1], 64)
		if err != nil {
			log.Printf("Error on Parsing Ask Entry Quantity for curr pair %s \n", currPair)
		}

		orderbook.UpdateAsks(currPair, priceVal, qtyVal)
	}
}
