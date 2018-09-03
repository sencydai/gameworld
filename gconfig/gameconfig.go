package gconfig

import (
	"sync"

	"github.com/sencydai/gameworld/log"
)

type gameConfig struct {
	LogPath       string
	LogLevel      log.LogLevel
	Port          int
	CertFile      string
	KeyFile       string
	ServerId      int
	MaxConnection uint
	RealMax       uint
	ConfigPath    string
	Database      string
	CrossUrl      string

	lock sync.RWMutex

	gameClose bool
}

var (
	GameConfig = gameConfig{}
	ServerIdML int64
)

func SetMaxCount(count uint) {
	GameConfig.lock.Lock()
	defer GameConfig.lock.Unlock()

	GameConfig.MaxConnection = count
	if GameConfig.RealMax > GameConfig.MaxConnection {
		GameConfig.RealMax = GameConfig.MaxConnection
	}
}

func SetRealCount(count uint) {
	GameConfig.lock.Lock()
	defer GameConfig.lock.Unlock()
	GameConfig.RealMax = count
	if GameConfig.RealMax > GameConfig.MaxConnection {
		GameConfig.RealMax = GameConfig.MaxConnection
	}
}

func GetMaxCount() uint {
	GameConfig.lock.RLock()
	GameConfig.lock.RUnlock()
	return GameConfig.MaxConnection
}

func GetRealCount() uint {
	GameConfig.lock.RLock()
	GameConfig.lock.RUnlock()
	return GameConfig.RealMax
}

func ReduceRealCount() {
	GameConfig.lock.Lock()
	defer GameConfig.lock.Unlock()

	if GameConfig.RealMax <= 100 {
		return
	}

	GameConfig.RealMax -= 1
}

func AddMaxCount() {
	GameConfig.lock.Lock()
	defer GameConfig.lock.Unlock()
	if GameConfig.RealMax >= GameConfig.MaxConnection {
		return
	}
	GameConfig.RealMax++
}

func CloseGame() {
	GameConfig.gameClose = true
}

func IsGameClose() bool {
	return GameConfig.gameClose
}
