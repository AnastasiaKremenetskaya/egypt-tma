package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/anterekhova/egypt-tma/config"
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

	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		logger.Fatalf("bot api: %v", err)
	}
	api.Debug = cfg.Debug
	logger.Printf("authorised as @%s", api.Self.UserName)

	st := store.NewMemoryStore()
	b := bot.New(api, st, bank, logger)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := api.GetUpdatesChan(u)
	logger.Println("polling for updates...")

	for update := range updates {
		go b.HandleUpdate(update)
	}
}
