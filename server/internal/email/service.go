package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/Adigezalov/goph-keeper/internal/logger"
)

type Service struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewService(host, port, username, password, from string) *Service {
	return &Service{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *Service) SendVerificationCode(toEmail, code string) error {
	if s.username == "" || s.password == "" || s.from == "" {
		logger.Warnf("[Email] SMTP не настроен. Код подтверждения для %s: %s", toEmail, code)
		return nil
	}

	subject := "Код подтверждения"
	body := code

	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.from, toEmail, subject, body)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	// Используем TLS для безопасного подключения к Yandex
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к SMTP серверу: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("не удалось создать SMTP клиент: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("не удалось авторизоваться: %w", err)
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("не удалось установить отправителя: %w", err)
	}

	if err := client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("не удалось установить получателя: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("не удалось открыть data writer: %w", err)
	}

	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("не удалось записать сообщение: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("не удалось закрыть data writer: %w", err)
	}

	logger.Infof("[Email] Код подтверждения отправлен на %s", toEmail)
	return nil
}
