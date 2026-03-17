package telegram

import (
	"fmt"
	"github.com/Ruseg557/go-telegram-bot/internal/config"
	"github.com/Ruseg557/go-telegram-bot/internal/services/transcriber"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Bot struct {
	config      *config.Config
	api         *tgbotapi.BotAPI
	transcriber *transcriber.Service
}

// NewBot Создаем бота
func NewBot(cfg *config.Config) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	tr := transcriber.New(cfg.ModelPath, cfg.WhisperExecutable)

	return &Bot{api: bot, config: cfg, transcriber: tr}, nil
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

// downloadAndTranscribe скачивает и транскрибирует аудио(голосовые, файлы)
func (b *Bot) downloadAndTranscribe(message *tgbotapi.Message, fileID, fileName string) string {
	if err := os.MkdirAll("temp", os.ModePerm); err != nil {
		log.Println("Ошибка создания папки temp:", err)
	}

	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Println("Ошибка обработки аудио:", err)
		return "Возникла ошибка обработки аудио: " + err.Error()
	}

	fileURL := file.Link(b.api.Token)

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
		return "Ошибка создания временного файла"
	}

	_, err = io.Copy(fileTemp, response.Body)
	if err != nil {
		fileTemp.Close()
		os.Remove(fileName)
		log.Println("Ошибка копирования:", err)
		return "Возникла ошибка"
	}
	fileTemp.Close()

	logFileInfo(message, fileName)

	text, err := b.transcriber.Transcribe(fileName)
	if err != nil {
		log.Println("Ошибка распознавания:", err)
		os.Remove(fileName)
		return "Не удалось распознать речь, попробуй ещё раз"
	}
	os.Remove(fileName)

	txtFile := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".txt"
	os.Remove(txtFile)

	return "Распознанный текст:\n\n" + text
}

// handleVoice обрабатыает голосовые сообщения
func (b *Bot) handleVoice(message *tgbotapi.Message) string {
	fileID := message.Voice.FileID

	fileName := filepath.Join("temp", fileID+".ogg")

	go b.sendProcessingMessage(message.Chat.ID, message.MessageID)

	return b.downloadAndTranscribe(message, fileID, fileName)
}

// handleAudio обрабатывает аудио сообщения (музыкальные файлы)
func (b *Bot) handleAudio(message *tgbotapi.Message) string {
	fileID := message.Audio.FileID

	fileName := filepath.Join("temp", fileID+".ogg")

	go b.sendProcessingMessage(message.Chat.ID, message.MessageID)

	return b.downloadAndTranscribe(message, fileID, fileName)
}

// handleDocument обрабатывает файлы
func (b *Bot) handleDocument(message *tgbotapi.Message) string {
	mimeType := message.Document.MimeType

	supportedAudio := map[string]bool{
		"audio/mpeg":     true,
		"audio/mp3":      true,
		"audio/ogg":      true,
		"audio/wav":      true,
		"audio/x-wav":    true,
		"audio/x-m4a":    true,
		"audio/aac":      true,
		"audio/flac":     true,
		"audio/vnd.wave": true,
	}
	if !supportedAudio[mimeType] {
		return fmt.Sprintf("Неподдерживаемый тип файла: %s. Отправь аудио в формате MP3, OGG, WAV, M4A или FLAC", mimeType)
	}

	fileID := message.Document.FileID
	ext := filepath.Ext(message.Document.FileName)
	if ext == "" {
		switch mimeType {
		case "audio/mpeg", "audio/mp3":
			ext = ".mp3"
		case "audio/ogg":
			ext = ".ogg"
		case "audio/wav", "audio/x-wav":
			ext = ".wav"
		case "audio/x-m4a":
			ext = ".m4a"
		default:
			ext = ".audio"
		}
	}
	fileName := filepath.Join("temp", fileID+ext)

	go b.sendProcessingMessage(message.Chat.ID, message.MessageID)

	return b.downloadAndTranscribe(message, fileID, fileName)
}

func (b *Bot) sendProcessingMessage(chatID int64, replyToID int) {
	msg := tgbotapi.NewMessage(chatID, "Аудио получено и в обработке...")
	msg.ReplyToMessageID = replyToID
	_, err := b.api.Send(msg)
	if err != nil {
		log.Println("Ошибка отправки сообщения о начале обработки аудио")
	}
}

// logFileInfo логирует информацию о файле
func logFileInfo(message *tgbotapi.Message, fileName string) {
	switch {
	case message.Voice != nil:
		log.Printf("Голосовое сообщение от %s сохранено: %s (длительность: %dс, размер: %d байт)",
			message.From.UserName, fileName, message.Voice.Duration, message.Voice.FileSize)
	case message.Audio != nil:
		log.Printf("Аудиофайл от %s сохранено: %s (название: %s, длительность: %dс)",
			message.From.UserName, fileName, message.Audio.Title, message.Audio.Duration)
	case message.Document != nil:
		log.Printf("Документ от %s сохранен: %s (имя: %s, MIME: %s)",
			message.From.UserName, fileName, message.Document.FileName, message.Document.MimeType)
	}
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
		} else if update.Message.Document != nil {
			text = b.handleDocument(update.Message)
		} else if update.Message.Audio != nil {
			text = b.handleAudio(update.Message)
		} else if update.Message.Text != "" {
			text = "Отправь аудио или голосовое и я его обработаю"
		} else {
			text = "Не умею работать с таким типом файлов(("
		}

		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			text)
		msg.ReplyToMessageID = update.Message.MessageID
		_, err := b.api.Send(msg)
		if err != nil {
			log.Println("Ошибка отправки сообщения:", err)
		}

		if update.Message.IsCommand() || update.Message.Text != "" {
			log.Printf("[%s]: %s", update.Message.From.UserName, update.Message.Text)
		}
	}
}
