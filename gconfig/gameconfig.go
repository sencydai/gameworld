package gconfig

import (
	"github.com/sencydai/utils/log"
	"sync"
)

type gameConfig struct {
	LogPath       string
	LogLevel      log.LogLevel
	Port          int
	ServerId      int
	MaxConnection int
	ConfigPath    string
	Database      string

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
