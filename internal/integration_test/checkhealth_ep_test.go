package integration_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/api"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestCheckHealthEndpoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupTestHTTP func(api api.Engine) *httptest.ResponseRecorder

		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "Normal case",
			setupTestHTTP: func(api api.Engine) *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodGet, "/health-check", nil)
				respRecorder := httptest.NewRecorder()

				api.ServeHTTP(respRecorder, req)
				return respRecorder
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"message":"OK", "service_name": "bookmark_service", "instance_id": "550e8400-e29b-41d4-a716-446655440000"}`,
		},
		{
			name: "Not found case",
			setupTestHTTP: func(api api.Engine) *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodPost, "/unknown-route", nil)
				respRecorder := httptest.NewRecorder()

				api.ServeHTTP(respRecorder, req)
				return respRecorder
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "404 page not found",
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

	gin.SetMode(gin.TestMode)
	testAPI := api.NewEngine(cfg, &svcCfg, nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := tc.setupTestHTTP(testAPI)

			assert.Equal(t, tc.expectedStatusCode, recorder.Code)

			if tc.expectedStatusCode == http.StatusOK {
				assert.JSONEq(t, tc.expectedResponseBody, recorder.Body.String())
			}

			if tc.expectedStatusCode == http.StatusNotFound {
				assert.Equal(t, tc.expectedResponseBody, strings.TrimSpace(recorder.Body.String()))
			}
		})
	}
}
