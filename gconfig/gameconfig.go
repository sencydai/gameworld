package gconfig

import (
	"github.com/sencydai/gameworld/log"
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
)

func CloseGame() {
	GameConfig.gameClose = true
}

func IsGameClose() bool {
	return GameConfig.gameClose
}
