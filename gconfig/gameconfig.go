package gconfig

import (
	"sync"

	"github.com/sencydai/utils/log"
)

type gameConfig struct {
	LogPath       string
	LogLevel      log.LogLevel
	Port          int
	ServerId      int
	MaxConnection int
	ConfigPath    string
	Database      string
	CrossUrl      string

	gameClose bool
}

var (
	GameConfig = gameConfig{}
	ServerIdML int64
	closeMu    sync.RWMutex
)

func CloseGame() {
	closeMu.Lock()
	defer closeMu.Unlock()

	GameConfig.gameClose = true
}

func IsGameClose() bool {
	closeMu.RLock()
	defer closeMu.RUnlock()
	return GameConfig.gameClose
}
