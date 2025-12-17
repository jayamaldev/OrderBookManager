package server

import (
	"fmt"
	"log"
	"net/http"
	"order-book-manager/orderbook"
	"order-book-manager/users"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var wg *sync.WaitGroup
var wgCount int
var SUBSCRIBE, UNSUBSCRIBE = "SUB", "UNSUB"

func CreateServer() *http.Server {
	return &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}
}

func RunWSServer(server *http.Server) {
	wg = new(sync.WaitGroup)

	http.HandleFunc("/ws", websocketHandler)

	go func() {
		fmt.Println("Websocket Server started on :8080")
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println("Error on websocket Server: ", err)
		}
	}()
	wg.Wait()
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error Upgrading Websocket: ", err)
		return
	}

	wgCount++
	wg.Add(wgCount)
	go handleConnection(conn, wgCount)
}

func handleConnection(conn *websocket.Conn, number int) {
	defer func() {
		wg.Done()
		users.RemoveUser(conn)
		err := conn.Close()
		if err != nil {
			log.Println("Error on Closing the Connection", err)
		}
	}()

	users.AddUser(conn)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error Reading Message. ", err)
			break
		}

		msgArgs := strings.Split(string(message), " ")

		fmt.Printf("Message Received: %s from %d \n", string(message), number)

		switch string(msgArgs[0]) {
		case SUBSCRIBE:
			currPair := msgArgs[1]
			fmt.Printf("Order Book Subscription Requested for curr pair %s\n", currPair)
			err := conn.WriteMessage(websocket.TextMessage, orderbook.GetOrderBook(currPair))
			if err != nil {
				fmt.Println("Error Writing Message: ", err)
			}

			users.SubUser(conn, currPair, true)
		case UNSUBSCRIBE:
			currPair := msgArgs[1]
			fmt.Printf("Order Book Unsubscription Requested for curr pair %s\n", currPair)
			users.SubUser(conn, currPair, false)
		default:
			fmt.Println("Unknown command received")
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Println("Error Writing Message: ", err)
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all connections
}
