package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/api"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/internal/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestShortenEndpoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupRedis     func(*testing.T) *redis.Client
		setupTestHTTP  func(api api.Engine) *httptest.ResponseRecorder
		expectedStatus int
	}{
		{
			name:       "success",
			setupRedis: testutil.InitMockRedis,
			setupTestHTTP: func(api api.Engine) *httptest.ResponseRecorder {
				req := httptest.NewRequest(
					http.MethodPost, "/v1/links/shorten",
					strings.NewReader(`{"url": "https://example.com", "exp": 3600}`),
				)
				req.Header.Set("Content-Type", "application/json")
				respRecorder := httptest.NewRecorder()

				api.ServeHTTP(respRecorder, req)
				return respRecorder
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:       "bad request",
			setupRedis: testutil.InitMockRedis,
			setupTestHTTP: func(api api.Engine) *httptest.ResponseRecorder {
				req := httptest.NewRequest(
					http.MethodPost, "/v1/links/shorten",
					strings.NewReader(`{"url": "https://example.com"}`),
				)
				req.Header.Set("Content-Type", "application/json")
				respRecorder := httptest.NewRecorder()

				api.ServeHTTP(respRecorder, req)
				return respRecorder
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "not found router",
			setupRedis: testutil.InitMockRedis,
			setupTestHTTP: func(api api.Engine) *httptest.ResponseRecorder {
				req := httptest.NewRequest(
					http.MethodPost, "/v1/links/unknown-route", nil)
				respRecorder := httptest.NewRecorder()

				api.ServeHTTP(respRecorder, req)
				return respRecorder
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "redis closed internal server error",
			setupRedis: testutil.InitClosedRedis,
			setupTestHTTP: func(api api.Engine) *httptest.ResponseRecorder {
				req := httptest.NewRequest(
					http.MethodPost, "/v1/links/shorten",
					strings.NewReader(`{"url": "https://example.com", "exp": 3600}`),
				)
				req.Header.Set("Content-Type", "application/json")
				respRecorder := httptest.NewRecorder()

				api.ServeHTTP(respRecorder, req)
				return respRecorder
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := api.NewConfig()
			if err != nil {
				t.Fatal(err)
			}
			svcCfg := service.Config{
				ServiceName: "bookmark_service",
				InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
			}
			gin.SetMode(gin.TestMode)
			testAPI := api.NewEngine(cfg, &svcCfg, tc.setupRedis(t))

			recorder := tc.setupTestHTTP(testAPI)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusCreated {
				var body map[string]string
				assert.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
				assert.Len(t, body["code"], 7)
				assert.Regexp(t, "^[a-zA-Z0-9]+$", body["code"])
				assert.Equal(t, "Shorten URL generated successfully!", body["message"])
			}

			if tc.expectedStatus == http.StatusBadRequest {
				var body map[string]string
				assert.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
				assert.Equal(t, "invalid request", body["error"])
			}

			if tc.expectedStatus == http.StatusInternalServerError {
				var body map[string]string
				assert.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
				assert.Equal(t, "internal server error", body["error"])
			}
		})
	}
}
