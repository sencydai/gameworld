package typedefine

import (
	"sync"

	"github.com/gorilla/websocket"
)

type AccountActor struct {
	ActorId   float64
	ActorName string
	Camp      int
	Sex       int
	Level     int
}

type Account struct {
	AccountId int
	Actor     *Actor
	GmLevel   byte

	conn    *websocket.Conn
	closed  bool
	closeMu sync.RWMutex

	writer chan []byte
}

func NewAccount(conn *websocket.Conn) *Account {
	account := &Account{conn: conn, writer: make(chan []byte, 32)}
	go func() {
		for data := range account.writer {
			if account.conn.WriteMessage(websocket.BinaryMessage, data) != nil {
				break
			}
		}
	}()
	return account
}

func (account *Account) Close() {
	account.closeMu.Lock()
	defer account.closeMu.Unlock()

	if account.closed {
		return
	}
	account.closed = true
	close(account.writer)
	account.conn.Close()
}

func (account *Account) IsClose() bool {
	account.closeMu.RLock()
	defer account.closeMu.RUnlock()

	return account.closed
}

func (account *Account) Reply(data []byte) {
	account.closeMu.RLock()
	defer account.closeMu.RUnlock()

	if account.closed {
		return
	}

	account.writer <- data

	//account.conn.WriteMessage(websocket.BinaryMessage, data)
}
