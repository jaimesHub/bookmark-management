package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/internal/service/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortenURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		inputURL string
		inputExp time.Duration

		mockSetup func(ctx context.Context, m *mocks.URLStorage)

		expectedCodeLength int
		expectedError      error
	}{
		{
			name:     "shorten URL success",
			inputURL: "https://example.com",
			inputExp: 24 * time.Hour,
			mockSetup: func(ctx context.Context, m *mocks.URLStorage) {
				m.On("StoreURLIfNotExists",
					ctx,
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
			mockSetup: func(ctx context.Context, m *mocks.URLStorage) {
				m.On("StoreURLIfNotExists",
					ctx, mock.Anything, mock.Anything, mock.Anything,
				).Return(false, errors.New("redis connection failed"))
			},
			expectedCodeLength: 0,
			expectedError:      errors.New("redis connection failed"),
		},
		{
			name:     "shorten URL with custom TTL",
			inputURL: "https://google.com",
			inputExp: 1 * time.Hour,
			mockSetup: func(ctx context.Context, m *mocks.URLStorage) {
				m.On("StoreURLIfNotExists",
					ctx,
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
			mockSetup: func(ctx context.Context, m *mocks.URLStorage) {
				m.On("StoreURLIfNotExists",
					ctx,
					mock.AnythingOfType("string"),
					"https://example.com",
					24*time.Hour,
				).Return(false, nil).Once() // lần 1: collision
				m.On("StoreURLIfNotExists",
					ctx,
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
			tc.mockSetup(ctx, mockRepo)

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

func TestShortenService_GetOriginalURL(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("redis connection failed")

	testCases := []struct {
		name string

		inputCode string

		mockSetup func(ctx context.Context, m *mocks.URLStorage)

		expectedURL    string
		expectedErrIs  error  // for errors.Is comparison; nil if not applicable
		expectedErrSub string // substring of error message; "" if not applicable
	}{
		{
			name:      "success",
			inputCode: "abc1234",
			mockSetup: func(ctx context.Context, m *mocks.URLStorage) {
				m.On("GetURL", ctx, "abc1234").Return("https://example.com", nil)
			},
			expectedURL: "https://example.com",
		},
		{
			name:      "not found",
			inputCode: "missing",
			mockSetup: func(ctx context.Context, m *mocks.URLStorage) {
				m.On("GetURL", ctx, "missing").Return("", redis.Nil)
			},
			expectedURL:   "",
			expectedErrIs: service.ErrCodeNotFound,
		},
		{
			name:      "repo error wrapped",
			inputCode: "abc1234",
			mockSetup: func(ctx context.Context, m *mocks.URLStorage) {
				m.On("GetURL", ctx, "abc1234").Return("", repoErr)
			},
			expectedURL:    "",
			expectedErrIs:  repoErr,
			expectedErrSub: "get original url",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			mockRepo := mocks.NewURLStorage(t)
			tc.mockSetup(ctx, mockRepo)

			svc := service.NewShortenService(mockRepo)

			url, err := svc.GetOriginalURL(ctx, tc.inputCode)

			assert.Equal(t, tc.expectedURL, url)

			if tc.expectedErrIs == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tc.expectedErrIs)
			if tc.expectedErrSub != "" {
				assert.Contains(t, err.Error(), tc.expectedErrSub)
			}
		})
	}
}
