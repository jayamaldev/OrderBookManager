// user handler package will manage the users and subscriptions they have
package users

import (
	"fmt"
	"log"
	"slices"

	"github.com/gorilla/websocket"
)

type User struct {
	CurrPairs []string
	Conn      *websocket.Conn
}

var usersList map[*websocket.Conn]*User

func InitUserList() {
	usersList = make(map[*websocket.Conn]*User)
}

func AddUser(conn *websocket.Conn) {
	user := &User{
		Conn: conn,
	}

	usersList[conn] = user
}

// manage subscriptions and unsubscriptions
func SubUser(conn *websocket.Conn, currPair string, subscribe bool) {
	subUser := usersList[conn]
	if subscribe {
		subUser.CurrPairs = append(subUser.CurrPairs, currPair)
	} else {
		for id, curr := range subUser.CurrPairs {
			if curr == currPair {
				subUser.CurrPairs = slices.Delete(subUser.CurrPairs, id, id+1)
			}
		}
	}
	fmt.Println("User Subscription List ", subUser.CurrPairs)
}

func RemoveUser(conn *websocket.Conn) {
	delete(usersList, conn)
}

// push market depth event to the subscribed users (subscribed to the given currency pair)
func PushEventToUsers(message []byte, currPar string) {
	for _, user := range usersList {
		if slices.Contains(user.CurrPairs, currPar) {
			err := user.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error on Writing to Websocket", err)
			}
		}
	}
}
