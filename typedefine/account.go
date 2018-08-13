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

	datas [][]byte
	start int
	end   int
	wLock sync.Mutex
}

func NewAccount(conn *websocket.Conn) *Account {
	account := &Account{conn: conn, datas: make([][]byte, 0)}
	go func() {
		write := account.conn.WriteMessage
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
				if len(account.datas) == 0 {
					account.wLock.Unlock()
					break
				}

				tick := time.Now()
				var index int
				for _, data := range account.datas {
					if write(bm, data) != nil {
						break
					}
					index++
					if time.Since(tick) > timeout {
						break
					}
				}
				account.datas = account.datas[index:]

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

	account.datas = append(account.datas, data)
}
