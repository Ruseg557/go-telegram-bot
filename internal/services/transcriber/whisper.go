package transcriber

type Service struct {
	modelPath string
}

func New(modelPath string) *Service {
	return &Service{modelPath: modelPath}
}

func (s *Service) Transcribe(filepath string) (string, error) {
	return "тест", nil
}
