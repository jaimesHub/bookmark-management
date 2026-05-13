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
	"github.com/stretchr/testify/assert"
)

func TestShortenEndpoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		setupTestHTTP func(api api.Engine) *httptest.ResponseRecorder

		expectedStatusCode int
	}{
		{
			name: "success",

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
			expectedStatusCode: http.StatusCreated,
		},
		{
			name: "bad request",

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
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "not found router",

			setupTestHTTP: func(api api.Engine) *httptest.ResponseRecorder {
				req := httptest.NewRequest(
					http.MethodPost, "/v1/links/unknown-route", nil)
				respRecorder := httptest.NewRecorder()

				api.ServeHTTP(respRecorder, req)
				return respRecorder
			},
			expectedStatusCode: http.StatusNotFound,
		},
	}

	cfg, err := api.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	svcCfg := service.Config{
		ServiceName: "bookmark_service",
		InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
	}

	redisClient := testutil.InitMockRedis(t)

	gin.SetMode(gin.TestMode)
	testAPI := api.NewEngine(cfg, &svcCfg, redisClient)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := tc.setupTestHTTP(testAPI)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedStatusCode == http.StatusCreated {
				var body map[string]string
				assert.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
				assert.Len(t, body["code"], 7)
				assert.Regexp(t, "^[a-zA-Z0-9]+$", body["code"])
			}

			if tc.expectedStatusCode == http.StatusBadRequest {
				var body map[string]string
				assert.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
				assert.NotEmpty(t, body["error"])
			}
		})
	}
}
