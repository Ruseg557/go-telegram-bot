package main

import (
	"github.com/Ruseg557/go-telegram-bot/internal/config"
	"github.com/Ruseg557/go-telegram-bot/internal/services/telegram"
	"log"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	// Создаем бота
	bot, err := telegram.NewBot(cfg)
	if err != nil {
		log.Fatal("Ошибка создания бота:", err)
	}

	// Проверка
	log.Printf("Бот авторизован как %s", bot.UserName())

	// TODO: обработка сообщений
	log.Println("Бот готов к работе!")
}
