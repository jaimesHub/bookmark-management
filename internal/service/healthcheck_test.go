package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		inputConfig Config

		expectedMessage     string
		expectedServiceName string
		expectedInstanceID  string

		//expectedError error
	}{
		{
			name: "check health success",
			inputConfig: Config{
				ServiceName: "bookmark-management",
				InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
			},
			expectedMessage:     "OK",
			expectedServiceName: "bookmark-management",
			expectedInstanceID:  "550e8400-e29b-41d4-a716-446655440000",
			//expectedError:       nil,
		},
		{
			name: "check health success with custom config",
			inputConfig: Config{
				ServiceName: "test-management",
				InstanceID:  "123456789",
			},
			expectedMessage:     "OK",
			expectedServiceName: "test-management",
			expectedInstanceID:  "123456789",
			//expectedError:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testSvc := NewHealthCheck(&tc.inputConfig)

			res, err := testSvc.Check()

			assert.Equal(t, tc.expectedMessage, res.Message)
			assert.Equal(t, tc.expectedServiceName, res.ServiceName)
			assert.Equal(t, tc.expectedInstanceID, res.InstanceID)
			//assert.ErrorIs(t, err, tc.expectedError)
			assert.NoError(t, err)
		})
	}
}
