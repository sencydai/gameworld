package data

import (
	"github.com/sencydai/gameworld/log"
	t "github.com/sencydai/gameworld/typedefine"
)

var (
	accounts = make(map[int]*t.Account)
)

func AppendAccount(account *t.Account) {
	if !account.IsClose() {
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

func GetAccount(accountId int) *t.Account {
	return accounts[accountId]
}

func GetAccountCount() int {
	return len(accounts)
}
