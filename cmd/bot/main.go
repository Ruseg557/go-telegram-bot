package main

import (
	"github.com/Ruseg557/go-telegram-bot/internal/config"
	"log"
)

func main() {
	// Загружаем конфигурациюп
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	log.Println("Конфигурация успешно загружена!")
	log.Printf("Токен бота: %s...", cfg.BotToken[:10])
	log.Printf("Путь к модели: %s", cfg.ModelPath)
	log.Printf("Режим отладки: %v", cfg.DebugMode)

	// TODO: здесь будет инициализация бота
	log.Println("Бот готов к запуску...")
}
