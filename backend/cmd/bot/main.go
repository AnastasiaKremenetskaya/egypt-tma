package main

import (
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/anterekhova/egypt-tma/config"
	"github.com/anterekhova/egypt-tma/internal/api"
	"github.com/anterekhova/egypt-tma/internal/bot"
	"github.com/anterekhova/egypt-tma/internal/questions"
	"github.com/anterekhova/egypt-tma/internal/store"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	logger := log.New(os.Stdout, "[egypt-tma] ", log.LstdFlags|log.Lshortfile)

	bank, err := questions.Load(cfg.DataDir)
	if err != nil {
		logger.Fatalf("load questions: %v", err)
	}
	logger.Printf("loaded %d maat / %d seth questions", len(bank.Maat), len(bank.Seth))

	tgAPI, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		logger.Fatalf("bot api: %v", err)
	}
	tgAPI.Debug = cfg.Debug
	logger.Printf("authorised as @%s", tgAPI.Self.UserName)

	st := store.NewMemoryStore()
	b := bot.New(tgAPI, st, bank, logger)

	hub := api.NewHub(logger)
	b.SetHub(hub)

	if cfg.DevMode {
		logger.Println("⚠️  DEV_MODE=true — Telegram auth bypass active, do NOT use in production")
	}
	apiServer := api.NewServer(b, hub, cfg.BotToken, cfg.WebAppURL, cfg.DevMode, logger)

	go func() {
		logger.Printf("HTTP API listening on %s", cfg.HTTPAddr)
		if err := http.ListenAndServe(cfg.HTTPAddr, apiServer.Handler()); err != nil {
			logger.Fatalf("http server: %v", err)
		}
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tgAPI.GetUpdatesChan(u)
	logger.Println("polling for updates...")

	for update := range updates {
		go b.HandleUpdate(update)
	}
}
