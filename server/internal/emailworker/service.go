package emailworker

import (
	"fmt"
	"math/rand"
)

// EmailWorkerService обертка для интеграции с user.Service
type EmailWorkerService struct {
	worker *Worker
}

func NewEmailWorkerService(worker *Worker) *EmailWorkerService {
	return &EmailWorkerService{
		worker: worker,
	}
}

func (s *EmailWorkerService) GenerateVerificationCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *EmailWorkerService) SendEmail(toEmail, code string) {
	s.worker.SendEmail(toEmail, code)
}
