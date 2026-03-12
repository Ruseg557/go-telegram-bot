package transcriber

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Service struct {
	modelPath  string
	executable string
}

func New(modelPath, executable string) *Service {
	return &Service{modelPath: modelPath, executable: executable}
}

// Transcribe Вызывает whisper.cpp для распознавания аудио
func (s *Service) Transcribe(filePath string) (string, error) {
	log.Printf("Запуск распознавания: %s", filePath)

	// Проверяем, что файл существует
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("файл не найден: %s", filePath)
	}

	// Подготавливаем команду
	// -f - входной файл
	// -m - модель
	// -otxt - вывод в текстовый файл (чтобы легче было прочитать)
	// -of - имя выходного файла (без расширения)

	outputBase := strings.TrimSuffix(filePath, ".ogg")

	cmd := exec.Command(s.executable,
		"-f", filePath,
		"-m", s.modelPath,
		"-otxt",
		"-of", outputBase,
	)

	// Захватываем stderr для отладки
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Запускаем
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ошибка whisper: %w\n%s", err, stderr.String())
	}

	// Читаем результат из txt файла
	resultFile := outputBase + ".txt"
	textBytes, err := os.ReadFile(resultFile)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать результат: %w", err)
	}

	text := strings.TrimSpace(string(textBytes))
	if text == "" {
		text = "[Распознано, но текст пуст]"
	}

	log.Printf("✅ Распознано: %s", text)
	return text, nil
}
