package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	BotToken  string
	ModelPath string
	DebugMode bool
}

func Load() (*Config, error) {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Println("Предупреждение: .env файл не найден, использую переменные окружения")
	}

	cfg := &Config{
		BotToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		ModelPath: os.Getenv("WHISPER_MODEL_PATH"),
		DebugMode: os.Getenv("DEBUG_MODE") == "true",
	}

	// Проверяем обязательные поля
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Устанавливаем значения по умолчанию
	if cfg.ModelPath == "" {
		cfg.ModelPath = "models/ggml-base.bin"
		log.Printf("Использую путь к модели по умолчанию: %s", cfg.ModelPath)
	}

	return cfg, nil
}

// Для удобства вывода информации о конфиге
func (c *Config) String() string {
	return fmt.Sprintf("BotToken: %s..., ModelPath: %s, DebugMode: %v",
		c.BotToken[:10], c.ModelPath, c.DebugMode)
}
