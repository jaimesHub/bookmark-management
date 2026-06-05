package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/model"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_Register(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	// validBody — reusable JSON body satisfy mọi binding tag (T2 RegisterRequest):
	//   username alphanum min=3 max=32 ✓ "alice"
	//   email format max=254          ✓ "alice@example.com"
	//   password min=8 max=72         ✓ "Passw0rd!"
	//   display_name min=1 max=64     ✓ "Alice"
	const validBody = `{"username":"alice","email":"alice@example.com","password":"Passw0rd!","display_name":"Alice"}`

	testCases := []struct {
		name string

		body             string
		setupMockService func(t *testing.T) *mocks.UserService

		expectedStatus    int
		expectedBodyAsset func(t *testing.T, body []byte)
	}{
		// ─── TC-L6-H001 — Happy path 201 ───────────────────────────
		{
			name: "happy path → 201 with PDF-locked message",
			body: validBody,
			setupMockService: func(t *testing.T) *mocks.UserService {
				m := mocks.NewUserService(t)
				m.On("Register", mock.Anything, mock.MatchedBy(func(req model.RegisterRequest) bool {
					return req.Username == "alice" &&
						req.Email == "alice@example.com" &&
						req.DisplayName == "Alice"
				})).Return(&model.User{
					ID:          "uid-001",
					Username:    "alice",
					Email:       "alice@example.com",
					DisplayName: "Alice",
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime,
				}, nil).Once()
				return m
			},
			expectedStatus: http.StatusCreated,
			expectedBodyAsset: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "Register an user successfully!", resp["message"])
				data, ok := resp["data"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "uid-001", data["id"])
				assert.Equal(t, "alice@example.com", data["email"])
				assert.NotContains(t, data, "password_hash") // json:"-" works
				assert.NotContains(t, data, "password")
			},
		},

		// ─── TC-L6-H002 — Body không phải JSON → 400 ───────────────
		{
			name: "body not JSON → 400 invalid request",
			body: `not json at all`,
			setupMockService: func(t *testing.T) *mocks.UserService {
				return mocks.NewUserService(t) // svc không được gọi
			},
			expectedStatus:    http.StatusBadRequest,
			expectedBodyAsset: assertInvalidRequest,
		},

		// ─── TC-L6-H003 — Thiếu username → 400 ─────────────────────
		{
			name:              "missing username → 400",
			body:              `{"email":"a@b.com","password":"Passw0rd!","display_name":"A"}`,
			setupMockService:  func(t *testing.T) *mocks.UserService { return mocks.NewUserService(t) },
			expectedStatus:    http.StatusBadRequest,
			expectedBodyAsset: assertInvalidRequest,
		},

		// ─── TC-L6-H004 — Thiếu email → 400 ────────────────────────
		{
			name:              "missing email → 400",
			body:              `{"username":"alice","password":"Passw0rd!","display_name":"Alice"}`,
			setupMockService:  func(t *testing.T) *mocks.UserService { return mocks.NewUserService(t) },
			expectedStatus:    http.StatusBadRequest,
			expectedBodyAsset: assertInvalidRequest,
		},

		// ─── TC-L6-H005 — Thiếu password → 400 ─────────────────────
		{
			name:              "missing password → 400",
			body:              `{"username":"alice","email":"a@b.com","display_name":"Alice"}`,
			setupMockService:  func(t *testing.T) *mocks.UserService { return mocks.NewUserService(t) },
			expectedStatus:    http.StatusBadRequest,
			expectedBodyAsset: assertInvalidRequest,
		},

		// ─── TC-L6-H006 — Thiếu display_name → 400 ─────────────────
		{
			name:              "missing display_name → 400",
			body:              `{"username":"alice","email":"a@b.com","password":"Passw0rd!"}`,
			setupMockService:  func(t *testing.T) *mocks.UserService { return mocks.NewUserService(t) },
			expectedStatus:    http.StatusBadRequest,
			expectedBodyAsset: assertInvalidRequest,
		},

		// ─── TC-L6-H007 — Email format sai → 400 ───────────────────
		{
			name:              "email format invalid → 400",
			body:              `{"username":"alice","email":"not-valid-email","password":"Passw0rd!","display_name":"Alice"}`,
			setupMockService:  func(t *testing.T) *mocks.UserService { return mocks.NewUserService(t) },
			expectedStatus:    http.StatusBadRequest,
			expectedBodyAsset: assertInvalidRequest,
		},

		// ─── TC-L6-H008 — Password < 8 ký tự → 400 ─────────────────
		{
			name:              "password too short (5 chars) → 400",
			body:              `{"username":"alice","email":"a@b.com","password":"short","display_name":"Alice"}`,
			setupMockService:  func(t *testing.T) *mocks.UserService { return mocks.NewUserService(t) },
			expectedStatus:    http.StatusBadRequest,
			expectedBodyAsset: assertInvalidRequest,
		},

		// ─── TC-L6-H009 — Username < 3 ký tự → 400 ─────────────────
		{
			name:              "username too short (2 chars) → 400",
			body:              `{"username":"ab","email":"a@b.com","password":"Passw0rd!","display_name":"Alice"}`,
			setupMockService:  func(t *testing.T) *mocks.UserService { return mocks.NewUserService(t) },
			expectedStatus:    http.StatusBadRequest,
			expectedBodyAsset: assertInvalidRequest,
		},

		// ─── TC-L6-H010 — Username chứa space (không alphanum) → 400 ─
		{
			name:              "username has space (not alphanum) → 400",
			body:              `{"username":"al ice","email":"a@b.com","password":"Passw0rd!","display_name":"Alice"}`,
			setupMockService:  func(t *testing.T) *mocks.UserService { return mocks.NewUserService(t) },
			expectedStatus:    http.StatusBadRequest,
			expectedBodyAsset: assertInvalidRequest,
		},

		// ─── TC-L6-H011 — ErrEmailAlreadyExists → 409 ──────────────
		{
			name: "ErrEmailAlreadyExists → 409 with specific message",
			body: validBody,
			setupMockService: func(t *testing.T) *mocks.UserService {
				m := mocks.NewUserService(t)
				m.On("Register", mock.Anything, mock.Anything).
					Return(nil, service.ErrEmailAlreadyExists).Once()
				return m
			},
			expectedStatus: http.StatusConflict,
			expectedBodyAsset: func(t *testing.T, body []byte) {
				var resp map[string]string
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "email already registered", resp["error"])
			},
		},

		// ─── TC-L6-H012 — ErrUsernameAlreadyExists → 409 ───────────
		{
			name: "ErrUsernameAlreadyExists → 409 with specific message",
			body: validBody,
			setupMockService: func(t *testing.T) *mocks.UserService {
				m := mocks.NewUserService(t)
				m.On("Register", mock.Anything, mock.Anything).
					Return(nil, service.ErrUsernameAlreadyExists).Once()
				return m
			},
			expectedStatus: http.StatusConflict,
			expectedBodyAsset: func(t *testing.T, body []byte) {
				var resp map[string]string
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "username already taken", resp["error"])
			},
		},

		// ─── TC-L6-H013 — Generic 5xx KHÔNG leak ───────────────────
		{
			name: "generic service error → 500 + no leak (CLAUDE.md MANDATORY)",
			body: validBody,
			setupMockService: func(t *testing.T) *mocks.UserService {
				m := mocks.NewUserService(t)
				m.On("Register", mock.Anything, mock.Anything).
					Return(nil, errors.New("database exploded — secret table name")).Once()
				return m
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBodyAsset: func(t *testing.T, body []byte) {
				var resp map[string]string
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "internal server error", resp["error"])
				// MANDATORY: raw err KHÔNG được leak ra HTTP response
				assert.NotContains(t, string(body), "database exploded")
				assert.NotContains(t, string(body), "secret table name")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context() // CLAUDE.md MANDATORY: ctx once at top-level

			rec := httptest.NewRecorder()
			gctx, _ := gin.CreateTestContext(rec)
			gctx.Request = httptest.NewRequest(
				http.MethodPost,
				"/v1/users/register",
				bytes.NewReader([]byte(tc.body)),
			).WithContext(ctx) // inject test ctx vào gin's request
			gctx.Request.Header.Set("Content-Type", "application/json")

			mockSvc := tc.setupMockService(t)
			h := NewUserHandler(mockSvc)

			h.Register(gctx)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			tc.expectedBodyAsset(t, rec.Body.Bytes())
		})
	}
}

// assertInvalidRequest — DRY assertion cho 400 cases (H002-H010).
// Body shape: {"error": "invalid request"} (handler trả generic).
func assertInvalidRequest(t *testing.T, body []byte) {
	t.Helper()
	var resp map[string]string
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Equal(t, "invalid request", resp["error"])
}
