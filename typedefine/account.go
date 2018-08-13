package typedefine

import (
	"sync"
	"time"

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

	writer map[int][]byte
	start  int
	end    int
	wLock  sync.Mutex
}

func NewAccount(conn *websocket.Conn) *Account {
	account := &Account{conn: conn, writer: make(map[int][]byte)}
	go func() {
		write := account.conn.WriteMessage
		datas := account.writer
		bm := websocket.BinaryMessage
		loopTime := time.Millisecond * 25
		timeout := time.Millisecond
		for {
			select {
			case <-time.After(loopTime):
				if account.IsClose() {
					return
				}

				account.wLock.Lock()

				tick := time.Now()
				for account.start < account.end {
					if write(bm, datas[account.start]) != nil {
						break
					}
					delete(datas, account.start)
					account.start++
					if time.Since(tick) > timeout {
						break
					}
				}

				account.wLock.Unlock()
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
	account.conn.Close()
}

func (account *Account) IsClose() bool {
	account.closeMu.RLock()
	defer account.closeMu.RUnlock()

	return account.closed
}

func (account *Account) Reply(data []byte) {
	account.wLock.Lock()
	defer account.wLock.Unlock()

	account.writer[account.end] = data
	account.end++
}
