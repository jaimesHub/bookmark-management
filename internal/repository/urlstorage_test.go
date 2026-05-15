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

		setupMock func(ctx context.Context) *redis.Client

		expectedError error
		verifyFunc    func(ctx context.Context, repo *urlStorage)
	}{
		{
			name: "normal case",

			setupMock: func(ctx context.Context) *redis.Client {
				return testutil.InitMockRedis(t)
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

			redisMock := tc.setupMock(ctx)
			testRepo := NewUrlStorage(redisMock)

			err := testRepo.StoreURL(ctx, "1234567", "https://google.com", urlExpTime)
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

		setupMock func(ctx context.Context) *redis.Client

		code        string
		expectedURL string
		expectedErr error // nil for success, redis.Nil for not found
	}{
		{
			name: "URL exists",
			setupMock: func(ctx context.Context) *redis.Client {
				mock := testutil.InitMockRedis(t)
				mock.Set(ctx, "url_exists", "https://x.com", urlExpTime)
				return mock
			},
			code:        "url_exists",
			expectedURL: "https://x.com",
			expectedErr: nil,
		},
		{
			name: "URL not found",
			setupMock: func(ctx context.Context) *redis.Client {
				return testutil.InitMockRedis(t)
			},
			code:        "url_not_exists",
			expectedURL: "",
			expectedErr: redis.Nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			redisMock := tc.setupMock(ctx)
			testRepo := NewUrlStorage(redisMock)

			url, err := testRepo.GetURL(ctx, tc.code)
			assert.Equal(t, tc.expectedURL, url)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestUrlStorage_StoreURLIfNotExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupMock      func(ctx context.Context) *redis.Client
		code           string
		expectedStored bool
		expectedErr    error
	}{
		{
			name: "code not exists",
			setupMock: func(ctx context.Context) *redis.Client {
				return testutil.InitMockRedis(t)
			},
			code:           "newcode1",
			expectedStored: true,
			expectedErr:    nil,
		},
		{
			name: "code already exists",
			setupMock: func(ctx context.Context) *redis.Client {
				mock := testutil.InitMockRedis(t)
				mock.Set(ctx, "existcode", "https://x.com", urlExpTime)
				return mock
			},
			code:           "existcode",
			expectedStored: false,
			expectedErr:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			redisMock := tc.setupMock(ctx)
			testRepo := NewUrlStorage(redisMock)

			stored, err := testRepo.StoreURLIfNotExists(ctx, tc.code, "https://example.com", urlExpTime)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedStored, stored)
		})
	}
}
