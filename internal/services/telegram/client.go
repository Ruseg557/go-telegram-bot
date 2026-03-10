package client

import (
	"github.com/Ruseg557/go-telegram-bot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type Bot struct {
	config *config.Config
	api    *tgbotapi.BotAPI
}

func NewBot(cfg *config.Config) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Panic(err)
	}
	return &Bot{api: bot, config: cfg}, nil
}
