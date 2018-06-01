package data

import (
	. "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils/log"
)

var (
	accounts = make(map[int]*Account)
)

func AppendAccount(account *Account) {
	if account.AccountId != 0 {
		accounts[account.AccountId] = account
		log.Infof("account(%d) login", account.AccountId)
	}
}

func RemoveAccount(accountId int) {
	if _, ok := accounts[accountId]; ok {
		delete(accounts, accountId)
		log.Infof("account(%d) logout", accountId)
	}
}

func GetAccount(accountId int) *Account {
	return accounts[accountId]
}

type LoopAccountsHandle func(*Account) bool

func LoopAccounts(handle LoopAccountsHandle) {
	for _, account := range accounts {
		if ok := handle(account); !ok {
			break
		}
	}
}
