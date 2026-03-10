package telegram

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
		return nil, err
	}
	return &Bot{api: bot, config: cfg}, nil
}

func (b Bot) UserName() string {
	return b.api.Self.UserName
}

func (b *Bot) Start() error {

	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 60

	updates := b.api.GetUpdatesChan(upd)

	go b.handleUpdates(updates)

	return nil
}

func (b Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s]: %s", update.Message.From.UserName, update.Message.Text)
		//TODO: обработка разных типов сообщений
	}
}
