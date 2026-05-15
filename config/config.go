package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	BotToken   string
	Debug      bool
	MaxPlayers int
	DataDir    string
}

func Load() Config {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is required")
	}

	debug := os.Getenv("DEBUG") == "true"

	maxPlayers := 8
	if s := os.Getenv("MAX_PLAYERS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 1 {
			maxPlayers = n
		}
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}

	return Config{
		BotToken:   token,
		Debug:      debug,
		MaxPlayers: maxPlayers,
		DataDir:    dataDir,
	}
}
