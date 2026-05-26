package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	BotToken   string
	Debug      bool
	DevMode    bool // DEV_MODE=true — bypasses Telegram initData HMAC (local dev only)
	MaxPlayers int
	DataDir    string
	HTTPAddr   string // e.g. ":8080"
	WebAppURL  string // HTTPS URL where frontend is hosted (for CORS)
}

func Load() Config {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is required")
	}

	debug := os.Getenv("DEBUG") == "true"
	devMode := os.Getenv("DEV_MODE") == "true"

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

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		// Railway injects PORT; fall back to 8080 for local dev
		if port := os.Getenv("PORT"); port != "" {
			httpAddr = ":" + port
		} else {
			httpAddr = ":8080"
		}
	}

	return Config{
		BotToken:   token,
		Debug:      debug,
		DevMode:    devMode,
		MaxPlayers: maxPlayers,
		DataDir:    dataDir,
		HTTPAddr:   httpAddr,
		WebAppURL:  os.Getenv("WEB_APP_URL"),
	}
}
