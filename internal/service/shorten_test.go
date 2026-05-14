package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortenURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		inputURL string
		inputExp time.Duration

		// Mock behavior
		mockSetup func(*mocks.URLStorage)

		// Expectations
		expectedCodeLength int
		expectedError      error
	}{
		{
			name:     "shorten URL success",
			inputURL: "https://example.com",
			inputExp: 24 * time.Hour,
			mockSetup: func(m *mocks.URLStorage) {
				m.On("StoreURLIfNotExists",
					mock.Anything,
					mock.AnythingOfType("string"),
					"https://example.com",
					24*time.Hour,
				).Return(true, nil)
			},
			expectedCodeLength: 7,
			expectedError:      nil,
		},
		{
			name:     "repository error",
			inputURL: "https://example.com",
			inputExp: 24 * time.Hour,
			mockSetup: func(m *mocks.URLStorage) {
				m.On("StoreURLIfNotExists",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(false, errors.New("redis connection failed"))
			},
			expectedCodeLength: 0,
			expectedError:      errors.New("redis connection failed"),
		},
		{
			name:     "shorten URL with custom TTL",
			inputURL: "https://google.com",
			inputExp: 1 * time.Hour,
			mockSetup: func(m *mocks.URLStorage) {
				m.On("StoreURLIfNotExists",
					mock.Anything,
					mock.AnythingOfType("string"),
					"https://google.com",
					1*time.Hour,
				).Return(true, nil)
			},
			expectedCodeLength: 7,
			expectedError:      nil,
		},
		{
			name:     "code collision retry success",
			inputURL: "https://example.com",
			inputExp: 24 * time.Hour,
			mockSetup: func(m *mocks.URLStorage) {
				m.On("StoreURLIfNotExists",
					mock.Anything,
					mock.AnythingOfType("string"),
					"https://example.com",
					24*time.Hour,
				).Return(false, nil).Once() // lần 1: collision
				m.On("StoreURLIfNotExists",
					mock.Anything,
					mock.AnythingOfType("string"),
					"https://example.com",
					24*time.Hour,
				).Return(true, nil).Once() // lần 2: success
			},
			expectedCodeLength: 7,
			expectedError:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			mockRepo := mocks.NewURLStorage(t)
			tc.mockSetup(mockRepo)

			svc := service.NewShortenService(mockRepo)

			code, err := svc.ShortenURL(ctx, tc.inputURL, tc.inputExp)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				assert.Empty(t, code)

			} else {
				assert.Regexp(t, "^[a-zA-Z0-9]+$", code)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCodeLength, len(code))
			}
		})
	}
}
