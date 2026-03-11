package telegram

import (
	"github.com/Ruseg557/go-telegram-bot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Bot struct {
	config *config.Config
	api    *tgbotapi.BotAPI
}

// NewBot Дает боту информацию о токене
func NewBot(cfg *config.Config) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}
	return &Bot{api: bot, config: cfg}, nil
}

// UserName Возвращает имя бота
func (b *Bot) UserName() string {
	return b.api.Self.UserName
}

// Start запускает канал для сообщений
func (b *Bot) Start() error {

	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 60

	updates := b.api.GetUpdatesChan(upd)

	go b.handleUpdates(updates)

	return nil
}

// TODO: Поднять локальный сервер. Убрать ограничение в 50 МБ

// handleVoice обрабатывает голосовые сообщения
func (b *Bot) handleVoice(message *tgbotapi.Message) string {
	if err := os.MkdirAll("temp", os.ModePerm); err != nil {
		log.Println("Ошибка создания папки temp:", err)
	}

	fileID := message.Voice.FileID

	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Println("Ошибка обработки аудио:", err)
		return "Возникла ошибка обработки аудио: " + err.Error()
	}

	fileURL := file.Link(b.api.Token)

	fileName := filepath.Join("temp", fileID+".ogg")

	response, err := http.Get(fileURL)
	if err != nil {
		log.Println("Ошибка скачивания файла:", err)
		return "Ошибка скачивания файла"
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Println("HTTP ошибка:", response.Status)
		return "Ошибка скачивания с сервера"
	}

	fileTemp, err := os.Create(fileName)
	if err != nil {
		log.Println("Ошибка создания файла:", err)
	}
	defer fileTemp.Close()

	_, err = io.Copy(fileTemp, response.Body)
	if err != nil {
		log.Println("Ошибка копирования тела ответа в файл:", err)
		return "Возникла ошибка"
	}

	log.Printf("Голосовое сообщение от %s сохранено: %s (длительность: %dс, размер: %d Мбайт)",
		message.From.UserName,
		fileName,
		message.Voice.Duration,
		message.Voice.FileSize/1024/1024)

	// TODO: Обработка аудио
	return "Скоро научусь обрабатывать"
}

// handleUpdates обрабатывает сообщения
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
			text = b.handleVoice(update.Message)
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
		}
	}
}
