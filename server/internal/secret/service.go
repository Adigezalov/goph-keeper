package secret

import (
	"log"
	"time"
)

// RealtimeService интерфейс для отправки событий в реальном времени (опционально)
type RealtimeService interface {
	NotifySecretCreated(userID int, secretID string, excludeSessionID string) error
	NotifySecretUpdated(userID int, secretID string, excludeSessionID string) error
	NotifySecretDeleted(userID int, secretID string, excludeSessionID string) error
}

// Service содержит бизнес-логику для работы с секретами
type Service struct {
	repo            Repository
	realtimeService RealtimeService // Опционально, может быть nil
}

// NewService создает новый экземпляр Service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// SetRealtimeService устанавливает сервис для отправки событий в реальном времени
func (s *Service) SetRealtimeService(realtimeService RealtimeService) {
	s.realtimeService = realtimeService
}

// SyncResponse представляет ответ для синхронизации
type SyncResponse struct {
	Secrets    []*Secret `json:"secrets"`     // Все измененные секреты (created/updated/deleted)
	ServerTime time.Time `json:"server_time"` // Текущее время сервера для следующего запроса
}

// CreateSecret создает новый секрет для пользователя
// excludeSessionID - опциональный ID WebSocket сессии, которую нужно исключить из рассылки
func (s *Service) CreateSecret(userID int, req *CreateSecretRequest, excludeSessionID string) (*Secret, error) {
	// Валидация
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Создаем секрет
	secret := &Secret{
		UserID:     userID,
		Login:      req.Login,
		Password:   req.Password,
		Metadata:   req.Metadata,
		BinaryData: req.BinaryData,
		Version:    1, // Начальная версия
	}

	if err := s.repo.CreateSecret(secret); err != nil {
		return nil, WrapError(err, "не удалось создать секрет")
	}

	// Отправляем событие через WebSocket (если сервис настроен)
	if s.realtimeService != nil {
		log.Printf("[Secret] Отправка события создания секрета: userID=%d, secretID=%s, excludeSessionID=%s", userID, secret.ID, excludeSessionID)
		if err := s.realtimeService.NotifySecretCreated(userID, secret.ID, excludeSessionID); err != nil {
			log.Printf("[Secret] Ошибка отправки события через WebSocket: %v", err)
			// Логируем ошибку, но не прерываем выполнение
			// Это graceful degradation - если WebSocket не работает, HTTP API продолжает работать
		} else {
			log.Printf("[Secret] Событие успешно отправлено через WebSocket")
		}
	} else {
		log.Printf("[Secret] RealtimeService не настроен, событие не отправлено")
	}

	return secret, nil
}

// GetSecret получает секрет по ID
func (s *Service) GetSecret(id string, userID int) (*Secret, error) {
	secret, err := s.repo.GetSecretByID(id, userID)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// GetAllSecrets получает все активные секреты пользователя
func (s *Service) GetAllSecrets(userID int) ([]*Secret, error) {
	secrets, err := s.repo.GetSecretsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Если секретов нет, возвращаем пустой slice вместо nil
	if secrets == nil {
		secrets = []*Secret{}
	}

	return secrets, nil
}

// UpdateSecret обновляет существующий секрет
// excludeSessionID - опциональный ID WebSocket сессии, которую нужно исключить из рассылки
func (s *Service) UpdateSecret(id string, userID int, req *UpdateSecretRequest, excludeSessionID string) (*Secret, error) {
	// Валидация
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Получаем текущий секрет для проверки прав доступа
	secret, err := s.repo.GetSecretByID(id, userID)
	if err != nil {
		return nil, err
	}

	// Проверяем версию (оптимистическая блокировка)
	if secret.Version != req.Version {
		return nil, ErrVersionConflict
	}

	// Обновляем данные
	secret.Login = req.Login
	secret.Password = req.Password
	secret.Metadata = req.Metadata
	secret.BinaryData = req.BinaryData
	// Version будет инкрементирован в repository

	if err := s.repo.UpdateSecret(secret); err != nil {
		return nil, WrapError(err, "не удалось обновить секрет")
	}

	// Отправляем событие через WebSocket (если сервис настроен)
	if s.realtimeService != nil {
		if err := s.realtimeService.NotifySecretUpdated(userID, secret.ID, excludeSessionID); err != nil {
			// Логируем ошибку, но не прерываем выполнение
		}
	}

	return secret, nil
}

// DeleteSecret выполняет мягкое удаление секрета
// excludeSessionID - опциональный ID WebSocket сессии, которую нужно исключить из рассылки
func (s *Service) DeleteSecret(id string, userID int, excludeSessionID string) error {
	if err := s.repo.SoftDeleteSecret(id, userID); err != nil {
		return err
	}

	// Отправляем событие через WebSocket (если сервис настроен)
	if s.realtimeService != nil {
		if err := s.realtimeService.NotifySecretDeleted(userID, id, excludeSessionID); err != nil {
			// Логируем ошибку, но не прерываем выполнение
		}
	}

	return nil
}

// GetSecretsForSync получает все секреты для синхронизации
// Если since == nil, возвращает все секреты (первая синхронизация)
// Если since != nil, возвращает только измененные после указанного времени
func (s *Service) GetSecretsForSync(userID int, since *time.Time) (*SyncResponse, error) {
	var secrets []*Secret
	var err error

	if since == nil {
		// Первая синхронизация - получаем все активные секреты
		secrets, err = s.repo.GetSecretsByUserID(userID)
	} else {
		// Инкрементальная синхронизация - только измененные
		secrets, err = s.repo.GetSecretsModifiedSince(userID, *since)
	}

	if err != nil {
		return nil, WrapError(err, "не удалось получить секреты для синхронизации")
	}

	// Если секретов нет, возвращаем пустой slice
	if secrets == nil {
		secrets = []*Secret{}
	}

	response := &SyncResponse{
		Secrets:    secrets,
		ServerTime: time.Now(), // Текущее время сервера для следующего запроса
	}

	return response, nil
}

// validateCreateRequest валидирует запрос на создание секрета
func (s *Service) validateCreateRequest(req *CreateSecretRequest) error {
	if req == nil {
		return ErrRequestRequired
	}

	if req.Login == "" {
		return ErrLoginRequired
	}

	if req.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

// validateUpdateRequest валидирует запрос на обновление секрета
func (s *Service) validateUpdateRequest(req *UpdateSecretRequest) error {
	if req == nil {
		return ErrRequestRequired
	}

	if req.Login == "" {
		return ErrLoginRequired
	}

	if req.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}
