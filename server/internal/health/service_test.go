package health

import "testing"

func TestService_CheckHealth(t *testing.T) {
	service := NewService()

	isHealthy := service.CheckHealth()

	if !isHealthy {
		t.Error("ожидали, что сервер здоров (true), но получили false")
	}
}

func TestNewService(t *testing.T) {
	service := NewService()

	if service == nil {
		t.Error("NewService должен возвращать не nil")
	}
}
