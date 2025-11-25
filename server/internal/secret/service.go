package secret

import (
	"time"

	"github.com/Adigezalov/goph-keeper/internal/logger"
)

type RealtimeService interface {
	NotifySecretCreated(userID int, secretID string, excludeSessionID string) error
	NotifySecretUpdated(userID int, secretID string, excludeSessionID string) error
	NotifySecretDeleted(userID int, secretID string, excludeSessionID string) error
}

type Service struct {
	repo            Repository
	realtimeService RealtimeService
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) SetRealtimeService(realtimeService RealtimeService) {
	s.realtimeService = realtimeService
}

type SyncResponse struct {
	Secrets    []*Secret `json:"secrets"`
	ServerTime time.Time `json:"server_time"`
}

func (s *Service) CreateSecret(userID int, req *CreateSecretRequest, excludeSessionID string) (*Secret, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	secret := &Secret{
		UserID:     userID,
		Login:      req.Login,
		Password:   req.Password,
		Metadata:   req.Metadata,
		BinaryData: req.BinaryData,
		Version:    1,
	}

	if err := s.repo.CreateSecret(secret); err != nil {
		return nil, WrapError(err, "не удалось создать секрет")
	}

	if s.realtimeService != nil {
		logger.Log.WithFields(map[string]interface{}{
			"user_id":         userID,
			"secret_id":       secret.ID,
			"exclude_session": excludeSessionID,
		}).Info("[Secret] Отправка события создания секрета")
		if err := s.realtimeService.NotifySecretCreated(userID, secret.ID, excludeSessionID); err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"user_id":   userID,
				"secret_id": secret.ID,
				"error":     err.Error(),
			}).Error("[Secret] Ошибка отправки события через WebSocket")
		} else {
			logger.Log.WithFields(map[string]interface{}{
				"user_id":   userID,
				"secret_id": secret.ID,
			}).Info("[Secret] Событие успешно отправлено через WebSocket")
		}
	} else {
		logger.Warn("[Secret] RealtimeService не настроен, событие не отправлено")
	}

	return secret, nil
}

func (s *Service) GetSecret(id string, userID int) (*Secret, error) {
	secret, err := s.repo.GetSecretByID(id, userID)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (s *Service) GetAllSecrets(userID int) ([]*Secret, error) {
	secrets, err := s.repo.GetSecretsByUserID(userID)
	if err != nil {
		return nil, err
	}

	if secrets == nil {
		secrets = []*Secret{}
	}

	return secrets, nil
}

func (s *Service) UpdateSecret(id string, userID int, req *UpdateSecretRequest, excludeSessionID string) (*Secret, error) {
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	secret, err := s.repo.GetSecretByID(id, userID)
	if err != nil {
		return nil, err
	}

	if secret.Version != req.Version {
		return nil, ErrVersionConflict
	}

	secret.Login = req.Login
	secret.Password = req.Password
	secret.Metadata = req.Metadata
	secret.BinaryData = req.BinaryData

	if err := s.repo.UpdateSecret(secret); err != nil {
		return nil, WrapError(err, "не удалось обновить секрет")
	}

	if s.realtimeService != nil {
		if err := s.realtimeService.NotifySecretUpdated(userID, secret.ID, excludeSessionID); err != nil {
		}
	}

	return secret, nil
}

func (s *Service) DeleteSecret(id string, userID int, excludeSessionID string) error {
	if err := s.repo.SoftDeleteSecret(id, userID); err != nil {
		return err
	}

	if s.realtimeService != nil {
		if err := s.realtimeService.NotifySecretDeleted(userID, id, excludeSessionID); err != nil {
		}
	}

	return nil
}

func (s *Service) GetSecretsForSync(userID int, since *time.Time) (*SyncResponse, error) {
	var secrets []*Secret
	var err error

	if since == nil {
		secrets, err = s.repo.GetSecretsByUserID(userID)
	} else {
		secrets, err = s.repo.GetSecretsModifiedSince(userID, *since)
	}

	if err != nil {
		return nil, WrapError(err, "не удалось получить секреты для синхронизации")
	}

	if secrets == nil {
		secrets = []*Secret{}
	}

	response := &SyncResponse{
		Secrets:    secrets,
		ServerTime: time.Now(),
	}

	return response, nil
}

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
