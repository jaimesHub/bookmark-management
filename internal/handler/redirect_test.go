package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRedirectHandler_Redirect(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		paramCode string

		setupMockService func(ctx *gin.Context, code string) *mocks.ShortenService

		expectedStatus   int
		expectedLocation string // for 302
		expectedErrBody  string // for 4xx/5xx
	}{
		{
			name:      "success",
			paramCode: "abc1234",
			setupMockService: func(ctx *gin.Context, code string) *mocks.ShortenService {
				m := mocks.NewShortenService(t)
				m.On("GetOriginalURL", ctx.Request.Context(), code).
					Return("https://example.com", nil)
				return m
			},
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://example.com",
		},
		{
			name:      "not found",
			paramCode: "missing",
			setupMockService: func(ctx *gin.Context, code string) *mocks.ShortenService {
				m := mocks.NewShortenService(t)
				m.On("GetOriginalURL", ctx.Request.Context(), code).
					Return("", service.ErrCodeNotFound)
				return m
			},
			expectedStatus:  http.StatusNotFound,
			expectedErrBody: "code not found",
		},
		{
			name:      "internal server error",
			paramCode: "abc1234",
			setupMockService: func(ctx *gin.Context, code string) *mocks.ShortenService {
				m := mocks.NewShortenService(t)
				m.On("GetOriginalURL", ctx.Request.Context(), code).
					Return("", errors.New("redis connection failed"))
				return m
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedErrBody: "internal server error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = httptest.NewRequest(
				http.MethodGet,
				"/v1/links/redirect/"+tc.paramCode,
				nil,
			)
			ctx.Params = gin.Params{{Key: "code", Value: tc.paramCode}}

			mockSvc := tc.setupMockService(ctx, tc.paramCode)
			testHandler := NewRedirect(mockSvc)

			testHandler.Redirect(ctx)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusFound {
				assert.Equal(t, tc.expectedLocation, rec.Header().Get("Location"))
				return
			}

			var body map[string]string
			assert.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			assert.Equal(t, tc.expectedErrBody, body["error"])
		})
	}
}
