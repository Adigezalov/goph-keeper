package health

import "testing"

func TestService_CheckHealth(t *testing.T) {
	// Arrange
	service := NewService()

	// Act
	isHealthy := service.CheckHealth()

	// Assert
	if !isHealthy {
		t.Error("ожидали, что сервер здоров (true), но получили false")
	}
}

func TestNewService(t *testing.T) {
	// Act
	service := NewService()

	// Assert
	if service == nil {
		t.Error("NewService должен возвращать не nil")
	}
}
