package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/api"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/internal/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedirectEndpoint(t *testing.T) {
	t.Parallel()

	const (
		seededCode = "abc1234"
		seededURL  = "https://example.com"
	)

	testCases := []struct {
		name string

		setupRedis func(ctx context.Context, t *testing.T) *redis.Client
		requestURL string

		expectedStatus   int
		expectedLocation string // for 302
		expectedErrBody  string // for 4xx/5xx
	}{
		{
			name: "success 302",
			setupRedis: func(ctx context.Context, t *testing.T) *redis.Client {
				client := testutil.InitMockRedis(t)
				err := client.Set(ctx, seededCode, seededURL, time.Hour).Err()
				assert.NoError(t, err)
				return client
			},
			requestURL:       "/v1/links/redirect/" + seededCode,
			expectedStatus:   http.StatusFound,
			expectedLocation: seededURL,
		},
		{
			name: "not found 404",
			setupRedis: func(ctx context.Context, t *testing.T) *redis.Client {
				return testutil.InitMockRedis(t)
			},
			requestURL:      "/v1/links/redirect/notexist",
			expectedStatus:  http.StatusNotFound,
			expectedErrBody: "code not found",
		},
		{
			name: "redis closed internal server error",
			setupRedis: func(ctx context.Context, t *testing.T) *redis.Client {
				return testutil.InitClosedRedis(t)
			},
			requestURL:      "/v1/links/redirect/" + seededCode,
			expectedStatus:  http.StatusInternalServerError,
			expectedErrBody: "internal server error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			cfg, err := api.NewConfig()
			if err != nil {
				t.Fatal(err)
			}
			svcCfg := service.Config{
				ServiceName: "bookmark_service",
				InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
			}
			gin.SetMode(gin.TestMode)
			testAPI := api.NewEngine(cfg, &svcCfg, tc.setupRedis(ctx, t))

			req := httptest.NewRequest(http.MethodGet, tc.requestURL, nil)
			recorder := httptest.NewRecorder()
			testAPI.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusFound {
				assert.Equal(t, tc.expectedLocation, recorder.Header().Get("Location"))
				return
			}

			var body map[string]string
			assert.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
			assert.Equal(t, tc.expectedErrBody, body["error"])
		})
	}
}
