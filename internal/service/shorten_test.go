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
				// Mock StoreURL succeed
				m.On("StoreURL",
					mock.Anything,                 // ctx
					mock.AnythingOfType("string"), // code (random)
					"https://example.com",         // url
					24*time.Hour,                  // exp
				).Return(nil)
			},
			expectedCodeLength: 7,
			expectedError:      nil,
		},
		{
			name:     "repository error",
			inputURL: "https://example.com",
			inputExp: 24 * time.Hour,
			mockSetup: func(m *mocks.URLStorage) {
				m.On("StoreURL",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(errors.New("redis connection failed"))
			},
			expectedCodeLength: 0, // empty code on error
			expectedError:      errors.New("redis connection failed"),
		},
		{
			name:     "shorten URL with custom TTL",
			inputURL: "https://google.com",
			inputExp: 1 * time.Hour,
			mockSetup: func(m *mocks.URLStorage) {
				m.On("StoreURL",
					mock.Anything,
					mock.AnythingOfType("string"),
					"https://google.com",
					1*time.Hour,
				).Return(nil)
			},
			expectedCodeLength: 7,
			expectedError:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			// 1. Create mock
			mockRepo := mocks.NewURLStorage(t)
			tc.mockSetup(mockRepo)

			// 2. Create service with mock
			svc := service.NewShortenService(mockRepo)

			// 3. Execute
			code, err := svc.ShortenURL(ctx, tc.inputURL, tc.inputExp)

			// 4. Assert
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				assert.Empty(t, code)

			} else {
				assert.Regexp(t, "^[a-zA-Z0-9]+$", code) // code should be alphanumeric
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCodeLength, len(code))
			}
		})
	}
}
