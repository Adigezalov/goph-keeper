package emailworker

import (
	"testing"
	"time"
)

func TestEmailWorkerService_GenerateVerificationCode(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 10, 3, 100*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	service := NewEmailWorkerService(worker)

	// Генерируем несколько кодов
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := service.GenerateVerificationCode()

		// Проверяем формат (6 цифр)
		if len(code) != 6 {
			t.Errorf("Expected code length 6, got %d for code '%s'", len(code), code)
		}

		// Проверяем, что все символы - цифры
		for _, ch := range code {
			if ch < '0' || ch > '9' {
				t.Errorf("Expected only digits, got '%c' in code '%s'", ch, code)
			}
		}

		codes[code] = true
	}

	// Проверяем, что коды разные (уникальность)
	// Вероятность коллизии крайне мала для 100 кодов из миллиона возможных
	if len(codes) < 90 {
		t.Errorf("Expected at least 90 unique codes out of 100, got %d", len(codes))
	}
}

func TestEmailWorkerService_SendEmail(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 10, 3, 100*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	service := NewEmailWorkerService(worker)

	// Отправляем email через сервис
	service.SendEmail("test@example.com", "123456")

	// Ждем обработки
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что задача была обработана
	calls := mockSender.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].To != "test@example.com" {
		t.Errorf("Expected to='test@example.com', got '%s'", calls[0].To)
	}

	if calls[0].Code != "123456" {
		t.Errorf("Expected code='123456', got '%s'", calls[0].Code)
	}
}

func TestEmailWorkerService_Integration(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 10, 3, 100*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	service := NewEmailWorkerService(worker)

	// Симулируем полный флоу регистрации
	email := "newuser@example.com"
	code := service.GenerateVerificationCode()

	// Отправляем код
	service.SendEmail(email, code)

	// Ждем обработки
	time.Sleep(200 * time.Millisecond)

	// Проверяем результат
	calls := mockSender.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].To != email {
		t.Errorf("Expected to='%s', got '%s'", email, calls[0].To)
	}

	if calls[0].Code != code {
		t.Errorf("Expected code='%s', got '%s'", code, calls[0].Code)
	}

	// Проверяем, что код 6-значный
	if len(calls[0].Code) != 6 {
		t.Errorf("Expected code length 6, got %d", len(calls[0].Code))
	}
}

func TestEmailWorkerService_MultipleUsers(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 50, 3, 50*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	service := NewEmailWorkerService(worker)

	// Симулируем регистрацию нескольких пользователей
	users := []string{
		"user1@example.com",
		"user2@example.com",
		"user3@example.com",
		"user4@example.com",
		"user5@example.com",
	}

	codes := make(map[string]string)
	for _, email := range users {
		code := service.GenerateVerificationCode()
		codes[email] = code
		service.SendEmail(email, code)
	}

	// Ждем обработки всех задач
	time.Sleep(500 * time.Millisecond)

	// Проверяем результаты
	calls := mockSender.GetCalls()
	if len(calls) != len(users) {
		t.Fatalf("Expected %d calls, got %d", len(users), len(calls))
	}

	// Проверяем, что каждый пользователь получил свой код
	callsMap := make(map[string]string)
	for _, call := range calls {
		callsMap[call.To] = call.Code
	}

	for email, expectedCode := range codes {
		actualCode, found := callsMap[email]
		if !found {
			t.Errorf("Email '%s' not found in calls", email)
			continue
		}
		if actualCode != expectedCode {
			t.Errorf("For email '%s': expected code '%s', got '%s'", email, expectedCode, actualCode)
		}
	}
}

func TestEmailWorkerService_NonBlocking(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 10, 3, 100*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	service := NewEmailWorkerService(worker)

	// Измеряем время вызова SendEmail
	start := time.Now()
	service.SendEmail("test@example.com", "123456")
	duration := time.Since(start)

	// Вызов должен быть быстрым (неблокирующим)
	// Даже если обработка займет 100мс, сам вызов должен быть мгновенным
	if duration > 10*time.Millisecond {
		t.Errorf("SendEmail should be non-blocking, took %v", duration)
	}
}
