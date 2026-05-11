package repository

import (
	"context"
	"testing"

	redisPkg "github.com/jaimesHub/bookmark-management/pkg/redis"
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
		verifyFunc    func(ctx context.Context, r *redis.Client)
	}{
		{
			name: "normal case",

			setupMock: func() *redis.Client {
				mock := redisPkg.InitMockRedis(t)
				return mock
			},

			expectedError: nil,
			verifyFunc: func(ctx context.Context, r *redis.Client) {
				url, err := r.Get(ctx, "1234567").Result()
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
				tc.verifyFunc(ctx, redisMock)
			}
		})
	}
}

// GetURL
