package repository

import (
	"testing"

	"github.com/jaimesHub/bookmark-management/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPingRepo_Ping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setupRepo func(t *testing.T) *pingRepo
		expectErr bool
	}{
		{
			name: "ping success",
			setupRepo: func(t *testing.T) *pingRepo {
				return NewPingRepo(testutil.InitMockRedis(t))
			},
			expectErr: false,
		},
		{
			name: "ping fail - redis unavailable",
			setupRepo: func(t *testing.T) *pingRepo {
				return NewPingRepo(testutil.InitClosedRedis(t))
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			repo := tc.setupRepo(t)
			err := repo.Ping(ctx)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
