package repository

import (
	"context"
	"testing"

	"github.com/jaimesHub/bookmark-management/internal/testutil"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// StoreURL
func TestUrlStorage_StoreURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		setupMock func() *redis.Client

		expectedError error
		verifyFunc    func(ctx context.Context, repo *urlStorage)
	}{
		{
			name: "normal case",

			setupMock: func() *redis.Client {
				mock := testutil.InitMockRedis(t)
				return mock
			},

			expectedError: nil,
			verifyFunc: func(ctx context.Context, repo *urlStorage) {
				url, err := repo.GetURL(ctx, "1234567")
				assert.Nil(t, err)
				assert.Equal(t, url, "https://google.com")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			redisMock := tc.setupMock()
			testRepo := NewUrlStorage(redisMock)

			err := testRepo.StoreURL(ctx, "1234567", "https://google.com")
			assert.Equal(t, tc.expectedError, err)
			if err == nil {
				tc.verifyFunc(ctx, testRepo)
			}
		})
	}
}

// GetURL
func TestUrlStorage_GetURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		setupMock func() *redis.Client

		code        string
		expectedURL string
		expectedErr error // nil for success, redis.Nil for not found
	}{
		{
			name: "URL exists",
			setupMock: func() *redis.Client {
				mock := testutil.InitMockRedis(t)
				// Pre-populate Redis with test data (using same TTL as production)
				mock.Set(t.Context(), "url_exists", "https://x.com", urlExpTime)
				return mock
			},
			code:        "url_exists",
			expectedURL: "https://x.com",
			expectedErr: nil, // No error expected
		},
		{
			name: "URL not found",
			setupMock: func() *redis.Client {
				mock := testutil.InitMockRedis(t)
				return mock
			},
			code:        "url_not_exists",
			expectedURL: "",
			expectedErr: redis.Nil, // Redis raises error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			redisMock := tc.setupMock()
			testRepo := NewUrlStorage(redisMock)

			url, err := testRepo.GetURL(ctx, tc.code)
			assert.Equal(t, tc.expectedURL, url)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
