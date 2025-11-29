package emailworker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// MockEmailSender для тестирования
type MockEmailSender struct {
	mu          sync.Mutex
	calls       []EmailCall
	shouldFail  bool
	failCount   int
	currentFail int
}

type EmailCall struct {
	To   string
	Code string
	Time time.Time
}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		calls: make([]EmailCall, 0),
	}
}

func (m *MockEmailSender) SendVerificationCode(toEmail, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, EmailCall{
		To:   toEmail,
		Code: code,
		Time: time.Now(),
	})

	if m.shouldFail {
		if m.failCount > 0 {
			m.currentFail++
			if m.currentFail <= m.failCount {
				return errors.New("mock error: failed to send email")
			}
		} else {
			return errors.New("mock error: failed to send email")
		}
	}

	return nil
}

func (m *MockEmailSender) GetCalls() []EmailCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]EmailCall{}, m.calls...)
}

func (m *MockEmailSender) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func (m *MockEmailSender) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make([]EmailCall, 0)
	m.currentFail = 0
}

func (m *MockEmailSender) SetShouldFail(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = fail
}

func (m *MockEmailSender) SetFailCount(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failCount = count
	m.currentFail = 0
}

func TestWorker_SuccessfulEmailSend(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 10, 3, 100*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	// Отправляем email
	worker.SendEmail("test@example.com", "123456")

	// Ждем обработки
	time.Sleep(200 * time.Millisecond)

	// Проверяем
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

func TestWorker_MultipleEmails(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 10, 3, 100*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	// Отправляем несколько email
	emails := []struct {
		to   string
		code string
	}{
		{"user1@example.com", "111111"},
		{"user2@example.com", "222222"},
		{"user3@example.com", "333333"},
	}

	for _, email := range emails {
		worker.SendEmail(email.to, email.code)
	}

	// Ждем обработки всех задач
	time.Sleep(500 * time.Millisecond)

	// Проверяем
	calls := mockSender.GetCalls()
	if len(calls) != 3 {
		t.Fatalf("Expected 3 calls, got %d", len(calls))
	}

	for i, email := range emails {
		if calls[i].To != email.to {
			t.Errorf("Call %d: expected to='%s', got '%s'", i, email.to, calls[i].To)
		}
		if calls[i].Code != email.code {
			t.Errorf("Call %d: expected code='%s', got '%s'", i, email.code, calls[i].Code)
		}
	}
}

func TestWorker_RetryOnFailure(t *testing.T) {
	mockSender := NewMockEmailSender()
	// Первые 2 попытки будут неудачными, третья успешная
	mockSender.SetFailCount(2)
	mockSender.SetShouldFail(true)

	worker := NewWorker(mockSender, 10, 3, 100*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	// Отправляем email
	worker.SendEmail("test@example.com", "123456")

	// Ждем всех попыток (3 попытки + задержки)
	time.Sleep(1 * time.Second)

	// Проверяем, что было 3 попытки
	calls := mockSender.GetCalls()
	if len(calls) != 3 {
		t.Fatalf("Expected 3 attempts, got %d", len(calls))
	}

	// Все попытки должны быть для одного и того же email
	for i, call := range calls {
		if call.To != "test@example.com" {
			t.Errorf("Attempt %d: expected to='test@example.com', got '%s'", i, call.To)
		}
		if call.Code != "123456" {
			t.Errorf("Attempt %d: expected code='123456', got '%s'", i, call.Code)
		}
	}
}

func TestWorker_MaxRetriesExceeded(t *testing.T) {
	mockSender := NewMockEmailSender()
	mockSender.SetShouldFail(true) // Всегда неудачно

	worker := NewWorker(mockSender, 10, 2, 50*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	// Отправляем email
	worker.SendEmail("test@example.com", "123456")

	// Ждем всех попыток
	time.Sleep(500 * time.Millisecond)

	// Проверяем, что было максимум 3 попытки (1 начальная + 2 retry)
	calls := mockSender.GetCalls()
	if len(calls) > 3 {
		t.Fatalf("Expected max 3 attempts, got %d", len(calls))
	}
}

func TestWorker_QueueLength(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 10, 3, 1*time.Second) // Большая задержка
	worker.Start()
	defer worker.Stop()

	// Отправляем несколько задач быстро
	worker.SendEmail("user1@example.com", "111111")
	worker.SendEmail("user2@example.com", "222222")
	worker.SendEmail("user3@example.com", "333333")

	// Проверяем длину очереди сразу
	queueLen := worker.QueueLength()
	if queueLen < 0 || queueLen > 3 {
		t.Errorf("Expected queue length between 0 and 3, got %d", queueLen)
	}
}

func TestWorker_GracefulShutdown(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 10, 3, 50*time.Millisecond)
	worker.Start()

	// Отправляем несколько задач
	worker.SendEmail("user1@example.com", "111111")
	worker.SendEmail("user2@example.com", "222222")
	worker.SendEmail("user3@example.com", "333333")

	// Небольшая пауза
	time.Sleep(100 * time.Millisecond)

	// Останавливаем worker
	worker.Stop()

	// Проверяем, что все задачи обработаны
	calls := mockSender.GetCalls()
	if len(calls) != 3 {
		t.Errorf("Expected 3 processed emails after shutdown, got %d", len(calls))
	}
}

func TestWorker_FullQueue(t *testing.T) {
	mockSender := NewMockEmailSender()
	// Маленькая очередь и большая задержка для заполнения
	worker := NewWorker(mockSender, 2, 3, 5*time.Second)
	worker.Start()
	defer worker.Stop()

	// Пытаемся отправить больше задач, чем вмещает очередь
	for i := 0; i < 10; i++ {
		worker.SendEmail("test@example.com", "123456")
	}

	// Проверяем, что задачи были добавлены (не все, так как очередь мала)
	queueLen := worker.QueueLength()
	if queueLen > 2 {
		t.Errorf("Queue should not exceed capacity of 2, got %d", queueLen)
	}
}

func TestWorker_ConcurrentSends(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 100, 3, 10*time.Millisecond)
	worker.Start()
	defer worker.Stop()

	// Отправляем задачи конкурентно
	var wg sync.WaitGroup
	numGoroutines := 10
	emailsPerGoroutine := 5

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < emailsPerGoroutine; j++ {
				email := "user@example.com"
				code := "123456"
				worker.SendEmail(email, code)
			}
		}(i)
	}

	wg.Wait()

	// Ждем обработки
	time.Sleep(1 * time.Second)

	// Проверяем, что все задачи обработаны
	expectedCalls := numGoroutines * emailsPerGoroutine
	actualCalls := mockSender.CallCount()

	if actualCalls != expectedCalls {
		t.Errorf("Expected %d calls, got %d", expectedCalls, actualCalls)
	}
}

func TestWorker_StopBeforeStart(t *testing.T) {
	mockSender := NewMockEmailSender()
	worker := NewWorker(mockSender, 10, 3, 100*time.Millisecond)

	// Пытаемся остановить до запуска
	// Не должно паниковать
	worker.Stop()

	// Теперь запускаем и останавливаем нормально
	worker = NewWorker(mockSender, 10, 3, 100*time.Millisecond)
	worker.Start()
	worker.Stop()
}
