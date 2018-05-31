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
	conn      *websocket.Conn
	closed    bool
	mutex     sync.RWMutex

	writer chan []byte
}

func NewAccount(conn *websocket.Conn) *Account {
	return &Account{conn: conn, writer: make(chan []byte, 10)}
}

func (account *Account) IsClosed() bool {
	account.mutex.RLock()
	defer account.mutex.RUnlock()

	return account.closed
}

func (account *Account) Close() {
	account.mutex.Lock()
	defer account.mutex.Unlock()

	if account.closed {
		return
	}
	close(account.writer)
	account.closed = true
	account.conn.Close()
}

func (account *Account) Reply(data []byte) {
	account.mutex.RLock()
	defer account.mutex.RUnlock()

	if account.closed {
		return
	}

	account.writer <- data
}

func (account *Account) GetData() chan []byte {
	return account.writer
}
