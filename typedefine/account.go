package typedefine

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sencydai/gameworld/log"
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
	account := &Account{conn: conn, writer: make(chan []byte, 64)}
	go func() {
		write := account.conn.WriteMessage
		bm := websocket.BinaryMessage
		for data := range account.writer {
			write(bm, data)
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
	account.conn.Close()
	close(account.writer)
}

func (account *Account) IsClose() bool {
	account.closeMu.RLock()
	defer account.closeMu.RUnlock()

	return account.closed
}

func (account *Account) Reply(data []byte) {
	account.closeMu.Lock()
	defer account.closeMu.Unlock()
	if account.closed {
		return
	}

	select {
	case account.writer <- data:
	case <-time.After(time.Second):
		account.closed = true
		account.conn.Close()
		close(account.writer)

		log.Warnf("server is very busy now,disconnect account %d", account.AccountId)
	}
}
