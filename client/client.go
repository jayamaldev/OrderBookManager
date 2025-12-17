package client

import (
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
)

var conn *websocket.Conn
var lastUpdateIds map[string]int
var depthUpdateEvent = "depthUpdate"
var snapshotURL = "https://api.binance.com/api/v3/depth?symbol=%s&limit=50"
var depthStr = "%s@depth"
var mainCurrencyPair string
var updateIdChan chan int
var bufferedEvents chan dtos.EventUpdate
var listSubscriptions chan []string
var firstEntryMap map[string]bool
var uniqueReqId int
var mutex sync.Mutex
var listSubscReqId int
var SUBSCRIBE, UNSUBSCRIBE, LIST_SUBSCRIPTIONS = "SUBSCRIBE", "UNSUBSCRIBE", "LIST_SUBSCRIPTIONS"
var WSS_STREAM, BINANCE_URL, WS_CONTEXT_ROOT = "wss", "stream.binance.com:9443", "/ws"

type SUbscriptionRequest struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int      `json:"id"`
}

type ListSubscriptionRequest struct {
	Method string `json:"method"`
	Id     int    `json:"id"`
}

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

func ConnectToWebSocket() {
	u := url.URL{
		Scheme: WSS_STREAM,
		Host:   BINANCE_URL,
		Path:   WS_CONTEXT_ROOT,
	}
	fmt.Printf("connecting to websocket %s\n", u.String())

	firstEntryMap = make(map[string]bool)

	var err error
	conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Websocket connectivity issue", err)
	}
	fmt.Printf("Connected to websocket %s\n", u.String())

	done := make(chan struct{})
	updateIdChan = make(chan int)
	bufferedEvents = make(chan dtos.EventUpdate, 1000)

	go func() {
		defer close(done)

		for {
			var eventUpdate dtos.EventUpdate
			var subscriptionsList dtos.SubscriptionsList

			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error on reading Websocket Message", err)
				return
			}

			err = json.Unmarshal(message, &eventUpdate)
			if err != nil {
				fmt.Println("Error Parsing Json", err)
				break
			}

			err = json.Unmarshal(message, &subscriptionsList)
			if err != nil {
				fmt.Println("Error Parsing Subscriptions List Json", err)
				break
			}

			if eventUpdate.Symbol != "" {
				if firstEntryMap[eventUpdate.Symbol] {
					fmt.Printf("Received first websocket message for currency %s \n", eventUpdate.Symbol)
					firstUpdateId := eventUpdate.FirstUpdateId
					fmt.Printf("first update Id %d for currency %s \n ", firstUpdateId, eventUpdate.Symbol)
					firstEntryMap[eventUpdate.Symbol] = false
					updateIdChan <- firstUpdateId
				}

				bufferedEvents <- eventUpdate
				users.PushEventToUsers(message, eventUpdate.Symbol)
			}

			if subscriptionsList.Id != 0 {
				fmt.Println("admin message received: ", string(message))
				fmt.Println("list subs req id: ", listSubscReqId)
				if subscriptionsList.Id == listSubscReqId {
					fmt.Println("sending subs list to channel ", subscriptionsList.Result)
					listSubscriptions <- subscriptionsList.Result
				}
			}
		}
	}()

	SubscribeToCurrPair(mainCurrencyPair)

	<-done
	fmt.Println("Websocket Client Closed")
}

func SubscribeToCurrPair(currencyPair string) {
	depthRequest := fmt.Sprintf(depthStr, strings.ToLower(currencyPair))
	subscriptionRequest := SUbscriptionRequest{
		Method: SUBSCRIBE,
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

	getMarketDepth(currencyPair)

	go updateEvents(currencyPair, bufferedEvents)
}

func UnsubscribeToCurrPair(currencyPair string) {
	depthRequest := fmt.Sprintf(depthStr, strings.ToLower(currencyPair))
	unsubscriptionRequest := SUbscriptionRequest{
		Method: UNSUBSCRIBE,
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

func ListSubscriptions() []string {
	listSubscReqId = getUniqueReqId()
	listSubscriptions = make(chan []string)

	listSubscriptionRequest := ListSubscriptionRequest{
		Method: LIST_SUBSCRIPTIONS,
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

func getMarketDepth(currPair string) {
	resp, err := http.Get(fmt.Sprintf(snapshotURL, currPair))
	if err != nil {
		log.Printf("Error on Creating New GET Request for curr paid %s\n", currPair)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf(fmt.Sprintf("Error on Closing Response for curr pair %s", currPair), err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Fatal(fmt.Sprintf("Error on Getting Snapshot for curr pair %s", currPair), err)
	}

	var snapshot *dtos.Snapshot
	err = json.NewDecoder(resp.Body).Decode(&snapshot)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cound not parse Response Json for curr pair %s", currPair), err)
	}

	lastUpdateId := snapshot.LastUpdateId
	lastUpdateIds[currPair] = lastUpdateId
	fmt.Printf("lastUpdateId for currency pair %s : %d: \n", currPair, lastUpdateIds[currPair])

	firstUpdateId := <-updateIdChan
	fmt.Printf("last update id for currency pair %s : %d first update id %d \n", currPair, lastUpdateId, firstUpdateId)
	if lastUpdateId > firstUpdateId {
		fmt.Printf("Condition Satisfied for curr pair %s !! \n", currPair)
	} else {
		fmt.Printf("Closing the Application. Re-get snapshot for currency pair %s\n", currPair)
		return
	}

	orderbook.PopulateOrderBook(currPair, snapshot)
}

func updateEvents(currPair string, bufferedEvents <-chan dtos.EventUpdate) {
	fmt.Printf("Event Processor Started for curr pair %s \n", currPair)
	for {
		eventUpdate := <-bufferedEvents
		if eventUpdate.FinalUpdateId > lastUpdateIds[currPair] {
			//process this
			if eventUpdate.EventType == depthUpdateEvent {
				formattedText := fmt.Sprintf("processing event for curr pair: %s %d %d %s", currPair, eventUpdate.FirstUpdateId, eventUpdate.FinalUpdateId, time.Now())
				fmt.Println(formattedText)
				for _, bidEntry := range eventUpdate.Bids {
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

				for _, askEntry := range eventUpdate.Asks {
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

				// update last update id of the orderbook
				lastUpdateIds[currPair] = eventUpdate.FinalUpdateId
			} else {
				fmt.Printf("Event type for curr pair %s %s \n", currPair, eventUpdate.EventType)
			}
		}
	}
}
