package notificationmocks

import (
	"github.com/stretchr/testify/mock"
)

type MockNotificationClient struct {
	mock.Mock
}

func (m *MockNotificationClient) SendEmailNotification(email, msg string) error {
	args := m.Called(email, msg)
	return args.Error(0)
}

func (m *MockNotificationClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
