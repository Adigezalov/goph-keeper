package health

// Service содержит бизнес-логику для health check
type Service struct {
	// В будущем можно добавить зависимости для проверки БД, кеша и т.д.
}

// NewService создает новый экземпляр Service
func NewService() *Service {
	return &Service{}
}

// CheckHealth проверяет доступность сервера
// Возвращает true, если сервер работает нормально
func (s *Service) CheckHealth() bool {
	// Базовая проверка - сервер запущен и отвечает
	// В будущем здесь можно добавить проверки:
	// - подключения к БД
	// - состояния кеша
	// - доступности внешних сервисов
	return true
}
