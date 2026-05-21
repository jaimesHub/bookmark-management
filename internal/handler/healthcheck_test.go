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

func TestHealthCheckHandler_CheckHealth(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		setupRequest     func(ctx *gin.Context)
		setupMockService func(ctx *gin.Context) *mocks.HealthCheck

		expectedStatus   int
		expectedResponse service.Response
	}{
		{
			name: "success",

			setupRequest: func(ctx *gin.Context) {
				ctx.Request = httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
			},
			setupMockService: func(ctx *gin.Context) *mocks.HealthCheck {
				serviceMock := mocks.NewHealthCheck(t)
				serviceMock.On("Check").Return(service.Response{
					Message:     "OK",
					ServiceName: "bookmark_service",
					InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
					Hostname:    "test-host-01",
				}, nil)
				return serviceMock
			},
			expectedStatus: http.StatusOK,
			expectedResponse: service.Response{
				Message:     "OK",
				ServiceName: "bookmark_service",
				InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
				Hostname:    "test-host-01",
			},
		},
		{
			name: "internal server error",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request = httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
			},
			setupMockService: func(ctx *gin.Context) *mocks.HealthCheck {
				serviceMock := mocks.NewHealthCheck(t)
				serviceMock.On("Check").Return(service.Response{}, errors.New("Internal Server Error!"))
				return serviceMock
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: service.Response{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			tc.setupRequest(ctx)

			mockSvc := tc.setupMockService(ctx)
			testHandler := NewHealthCheck(mockSvc)

			testHandler.CheckHealth(ctx)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp service.Response

				err := json.NewDecoder(rec.Body).Decode(&resp)

				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResponse.Message, resp.Message)
				assert.Equal(t, tc.expectedResponse.ServiceName, resp.ServiceName)
				assert.Equal(t, tc.expectedResponse.InstanceID, resp.InstanceID)
				assert.Equal(t, tc.expectedResponse.Hostname, resp.Hostname)
			}

			if tc.expectedStatus == http.StatusInternalServerError {
				assert.JSONEq(t, `{"error":"Internal Server Error!"}`, rec.Body.String())
			}
		})
	}
}
