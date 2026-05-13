package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestShortenHandler_ShortenURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		setupRequest     func(ctx *gin.Context)
		setupMockService func(ctx *gin.Context) *mocks.ShortenService

		expectedStatus int
		expectedBody   string // JSON string
	}{
		{
			name: "success",

			setupRequest: func(ctx *gin.Context) {
				ctx.Request = httptest.NewRequest(
					http.MethodPost, "/v1/links/shorten",
					strings.NewReader(`{"url": "https://example.com", "exp": 3600}`),
				)
				ctx.Request.Header.Set("Content-Type", "application/json")
			},
			setupMockService: func(ctx *gin.Context) *mocks.ShortenService {
				serviceMock := mocks.NewShortenService(t)
				serviceMock.On(
					"ShortenURL",
					ctx.Request.Context(),
					"https://example.com",
					3600*time.Second,
				).Return("abc1234", nil)
				return serviceMock
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"code": "abc1234"}`,
		},
		{
			name: "bad request",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request = httptest.NewRequest(
					http.MethodPost, "/v1/links/shorten",
					strings.NewReader(`{"url": "https://example.com"}`),
				)
				ctx.Request.Header.Set("Content-Type", "application/json")
			},
			setupMockService: func(ctx *gin.Context) *mocks.ShortenService {
				return mocks.NewShortenService(t)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "internal server error",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request = httptest.NewRequest(
					http.MethodPost, "/v1/links/shorten",
					strings.NewReader(`{"url": "https://example.com", "exp": 3600}`),
				)
				ctx.Request.Header.Set("Content-Type", "application/json")
			},
			setupMockService: func(ctx *gin.Context) *mocks.ShortenService {
				serviceMock := mocks.NewShortenService(t)
				serviceMock.On(
					"ShortenURL",
					ctx.Request.Context(),
					"https://example.com", 3600*time.Second,
				).Return("", errors.New("redis connection failed"))
				return serviceMock
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error": "redis connection failed"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			tc.setupRequest(ctx)

			mockSvc := tc.setupMockService(ctx)
			testHandler := NewShorten(mockSvc)

			testHandler.ShortenURL(ctx)

			if tc.expectedStatus == http.StatusCreated {
				assert.Equal(t, tc.expectedStatus, rec.Code)
				assert.JSONEq(t, tc.expectedBody, rec.Body.String())
			}

			if tc.expectedStatus == http.StatusBadRequest {
				var body map[string]string
				assert.Equal(t, tc.expectedStatus, rec.Code)
				assert.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
				assert.NotEmpty(t, body["error"])
			}

			if tc.expectedStatus == http.StatusInternalServerError {
				var body map[string]string
				assert.Equal(t, tc.expectedStatus, rec.Code)
				assert.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
				assert.Equal(t, "redis connection failed", body["error"])
			}
		})
	}
}
