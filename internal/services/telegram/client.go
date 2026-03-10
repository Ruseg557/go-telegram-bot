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

func (b *Bot) UserName() string {
	return b.api.Self.UserName
}

func (b *Bot) Start() error {

	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 60

	updates := b.api.GetUpdatesChan(upd)

	go b.handleUpdates(updates)

	return nil
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		var text string
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				text = "Привет, я бот для преобразования голосовых сообщений и аудио файлов в текст. " +
					"А возможно и не только :) используй /help для помощи"
			case "help":
				text = "Для преобразования аудио в текст просто отправь мне запись"
			default:
				text = "Извини, но я не знаю такой комманды("
			}
		} else if update.Message.Voice != nil {
			text = "Скоро научусь обрабатывать голосовые сообщения и аудио)"
		} else if update.Message.Text != "" {
			text = "Отправь аудио или голосовое и я его обработаю"
		} else {
			text = "Не умею работать с таким типом файлов(("
		}

		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			text)
		_, err := b.api.Send(msg)
		if err != nil {
			log.Println("Ошибка отправки сообщения:", err)
		}

		if update.Message.IsCommand() || update.Message.Text != "" {
			log.Printf("[%s]: %s", update.Message.From.UserName, update.Message.Text)
		} else {
			log.Printf("[%s]: Аудио/видео/файл", update.Message.From.UserName)
		}
	}
}
