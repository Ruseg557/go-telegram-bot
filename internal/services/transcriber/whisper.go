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

// convertToWav конвертирует OGG в WAV через FFmpeg
func (s *Service) convertToWav(oggPath string) (string, error) {
	wavPath := strings.TrimSuffix(oggPath, ".ogg") + ".wav"

	// ffmpeg -i input.ogg -ar 16000 -ac 1 -c:a pcm_s16le output.wav
	cmd := exec.Command("ffmpeg",
		"-i", oggPath,
		"-ar", "16000", // частота 16kHz
		"-ac", "1", // моно
		"-c:a", "pcm_s16le", // 16-bit PCM
		"-y", // перезаписывать без вопросов
		wavPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ошибка конвертации FFmpeg: %w\n%s", err, stderr.String())
	}

	return wavPath, nil
}

// Transcribe Вызывает whisper.cpp для распознавания аудио
func (s *Service) Transcribe(filePath string) (string, error) {
	log.Printf("Запуск распознавания: %s", filePath)

	// Проверяем, что файл существует
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("файл не найден: %s", filePath)
	}

	// Конвертируем OGG в WAV
	wavPath, err := s.convertToWav(filePath)
	if err != nil {
		return "", fmt.Errorf("конвертация не удалась: %w", err)
	}

	// Удаляем WAV после обработки
	defer func() {
		if err := os.Remove(wavPath); err != nil {
			log.Printf("Не удалось удалить WAV файл %s: %v", wavPath, err)
		}
	}()

	// Подготавливаем команду
	// -f - входной файл
	// -m - модель
	// -l ru - русский язык
	// --no-timestamps - убрать временные метки из вывода
	cmd := exec.Command(s.executable,
		"-f", wavPath,
		"-m", s.modelPath,
		"-l", "ru",
		"--no-timestamps",
	)

	// Захватываем stdout и stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Запускаем
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ошибка whisper: %w\n%s", err, stderr.String())
	}

	// Получаем текст из stdout
	text := strings.TrimSpace(stdout.String())
	if text == "" {
		text = strings.TrimSpace(stderr.String())
	}

	if text == "" {
		return "", fmt.Errorf("whisper не вернул текст")
	}

	log.Printf("Распознано: %s", text)
	return text, nil
}
