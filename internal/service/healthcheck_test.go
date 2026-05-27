package service_test

import (
	"errors"
	"testing"

	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		inputConfig service.Config
		mockSetup   func(*mocks.Pinger)

		expectedResponse service.Response
		expectedError    error
	}{
		{
			name: "ping success",
			inputConfig: service.Config{
				ServiceName: "bookmark-management",
				InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
				Hostname:    "test-host-01",
			},
			mockSetup: func(m *mocks.Pinger) {
				m.On("Ping", mock.Anything).Return(nil)
			},
			expectedResponse: service.Response{
				Message:     "OK",
				ServiceName: "bookmark-management",
				InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
				Hostname:    "test-host-01",
			},
			expectedError: nil,
		},
		{
			name: "ping success with empty hostname",
			inputConfig: service.Config{
				ServiceName: "bookmark-management",
				InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
				Hostname:    "",
			},
			mockSetup: func(m *mocks.Pinger) {
				m.On("Ping", mock.Anything).Return(nil)
			},
			expectedResponse: service.Response{
				Message:     "OK",
				ServiceName: "bookmark-management",
				InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
				Hostname:    "",
			},
			expectedError: nil,
		},
		{
			name: "ping failure",
			inputConfig: service.Config{
				ServiceName: "bookmark-management",
				InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
				Hostname:    "test-host-01",
			},
			mockSetup: func(m *mocks.Pinger) {
				m.On("Ping", mock.Anything).Return(errors.New("redis unreachable"))
			},
			expectedResponse: service.Response{},
			expectedError:    errors.New("redis unreachable"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockPinger := mocks.NewPinger(t)
			tc.mockSetup(mockPinger)

			svc := service.NewHealthCheck(&tc.inputConfig, mockPinger)

			res, err := svc.Check()

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				assert.Equal(t, service.Response{}, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, res)
			}
		})
	}
}
