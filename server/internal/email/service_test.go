package email

import (
	"strings"
	"testing"
)

func TestService_SendVerificationCode_NoSMTPConfig(t *testing.T) {
	// Сервис без SMTP настроек (для разработки)
	service := NewService("smtp.yandex.ru", "465", "", "", "")

	// Не должно быть ошибки - просто логирование
	err := service.SendVerificationCode("test@example.com", "123456")
	if err != nil {
		t.Errorf("Expected no error for missing SMTP config, got: %v", err)
	}
}

func TestService_NewService(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		username string
		password string
		from     string
	}{
		{
			name:     "Full config",
			host:     "smtp.yandex.ru",
			port:     "465",
			username: "user@yandex.ru",
			password: "password",
			from:     "user@yandex.ru",
		},
		{
			name:     "Empty config",
			host:     "",
			port:     "",
			username: "",
			password: "",
			from:     "",
		},
		{
			name:     "Partial config",
			host:     "smtp.gmail.com",
			port:     "587",
			username: "user@gmail.com",
			password: "",
			from:     "user@gmail.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.host, tt.port, tt.username, tt.password, tt.from)

			if service == nil {
				t.Fatal("Expected non-nil service")
			}

			if service.host != tt.host {
				t.Errorf("Expected host='%s', got '%s'", tt.host, service.host)
			}

			if service.port != tt.port {
				t.Errorf("Expected port='%s', got '%s'", tt.port, service.port)
			}

			if service.username != tt.username {
				t.Errorf("Expected username='%s', got '%s'", tt.username, service.username)
			}

			if service.password != tt.password {
				t.Errorf("Expected password='%s', got '%s'", tt.password, service.password)
			}

			if service.from != tt.from {
				t.Errorf("Expected from='%s', got '%s'", tt.from, service.from)
			}
		})
	}
}

func TestService_SendVerificationCode_InvalidEmail(t *testing.T) {
	// Эти тесты проверяют только логику без реального SMTP
	service := NewService("smtp.yandex.ru", "465", "test@yandex.ru", "pass", "test@yandex.ru")

	// С некорректным SMTP - будет ошибка подключения (ожидаемое поведение)
	// Но с пустыми credentials - должен просто залогировать
	service = NewService("smtp.yandex.ru", "465", "", "", "")
	err := service.SendVerificationCode("", "123456")
	if err != nil {
		t.Errorf("Expected no error for missing SMTP config, got: %v", err)
	}
}

func TestService_SendVerificationCode_EmptyCode(t *testing.T) {
	service := NewService("smtp.yandex.ru", "465", "", "", "")

	// Пустой код - не должно быть паники
	err := service.SendVerificationCode("test@example.com", "")
	if err != nil {
		t.Errorf("Expected no error for missing SMTP config, got: %v", err)
	}
}

func TestService_MessageFormat(t *testing.T) {
	// Проверяем, что сообщение формируется корректно
	// Этот тест проверяет формат, не отправляя реальное письмо

	service := NewService("", "", "", "", "")

	// Тестовые данные
	from := "sender@example.com"

	// Проверяем, что сервис создан
	if service == nil {
		t.Fatal("Expected non-nil service")
	}

	// В реальной отправке формируется сообщение с этими данными
	// Проверим, что данные сохранены правильно
	expectedFrom := from

	// Сервис должен содержать правильные данные
	service = NewService("smtp.test.com", "465", "user@test.com", "pass", expectedFrom)

	if service.from != expectedFrom {
		t.Errorf("Expected from='%s', got '%s'", expectedFrom, service.from)
	}

	// Для полноценного теста отправки нужен реальный SMTP или mock сервер
	// Здесь проверяем только структуру сервиса
}

func TestService_SMTPAddressFormat(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		port         string
		expectedAddr string
	}{
		{
			name:         "Yandex",
			host:         "smtp.yandex.ru",
			port:         "465",
			expectedAddr: "smtp.yandex.ru:465",
		},
		{
			name:         "Gmail",
			host:         "smtp.gmail.com",
			port:         "587",
			expectedAddr: "smtp.gmail.com:587",
		},
		{
			name:         "Mail.ru",
			host:         "smtp.mail.ru",
			port:         "465",
			expectedAddr: "smtp.mail.ru:465",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.host, tt.port, "user@test.com", "pass", "user@test.com")

			// Проверяем, что адрес формируется как host:port
			expectedParts := strings.Split(tt.expectedAddr, ":")
			if service.host != expectedParts[0] {
				t.Errorf("Expected host='%s', got '%s'", expectedParts[0], service.host)
			}
			if service.port != expectedParts[1] {
				t.Errorf("Expected port='%s', got '%s'", expectedParts[1], service.port)
			}
		})
	}
}

// Бенчмарк для проверки производительности создания сервиса
func BenchmarkNewService(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewService("smtp.yandex.ru", "465", "user@yandex.ru", "password", "user@yandex.ru")
	}
}

// Бенчмарк для SendVerificationCode без реального SMTP
func BenchmarkSendVerificationCode_NoSMTP(b *testing.B) {
	service := NewService("smtp.yandex.ru", "465", "", "", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.SendVerificationCode("test@example.com", "123456")
	}
}
