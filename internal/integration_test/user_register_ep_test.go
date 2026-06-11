package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegisterUser_EndToEnd — full register flow tests qua HTTP layer.
//
// Pattern: shared SQLite + ordered subtests (per Plan § 5.3 + QA note).
//   - I001 (happy) chạy TRƯỚC → insert alice@example.com + username alice
//   - I005 (dup email) + I006 (dup username) depend I001
//   - Other cases (400/edge) independent
//
// Trade-off: shared DB simpler vs full isolation per case (slower).
// QA accept ordering dependency với documented note.
func TestRegisterUser_EndToEnd(t *testing.T) {
	// Setup shared DB + engine ngoài for loop — match Plan § 5.3
	db := SetupTestDB(t)
	engine := BuildEngineWithDB(t, db)

	// Helper: gen 72-byte ASCII password (TC-I010 bcrypt boundary)
	pwd72 := strings.Repeat("A", 71) + "1"

	testCases := []struct {
		name     string
		body     map[string]any
		wantCode int
		verify   func(t *testing.T, body []byte)
	}{
		// ─── TC-L6-I001 — 201 Happy path + full response shape ───────
		{
			name: "I001 — 201 happy path + spec compliance",
			body: map[string]any{
				"username": "alice", "email": "alice@example.com",
				"password": "Passw0rd!", "display_name": "Alice",
			},
			wantCode: http.StatusCreated,
			verify: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))

				// Spec compliance (TC-I015): exact message string
				assert.Equal(t, "Register an user successfully!", resp["message"])

				// Spec compliance (TC-I016): 'data' key present (NOT 'user'/'result')
				data, ok := resp["data"].(map[string]any)
				require.True(t, ok, "response phải có key 'data'")
				_, hasUser := resp["user"]
				assert.False(t, hasUser, "response không được có key 'user'")

				// Field assertions
				assert.Equal(t, "alice@example.com", data["email"])
				assert.Equal(t, "alice", data["username"])
				assert.Equal(t, "Alice", data["display_name"])
				assert.NotEmpty(t, data["created_at"])
				assert.NotEmpty(t, data["updated_at"])

				// Spec compliance (TC-I017): id là UUID format
				id, ok := data["id"].(string)
				require.True(t, ok)
				_, err := uuid.Parse(id)
				assert.NoError(t, err, "id phải parse được thành UUID")

				// Spec compliance (TC-I018): password_hash + Passw0rd! absent
				bodyStr := string(body)
				assert.NotContains(t, bodyStr, "password_hash", "json:\"-\" phải hoạt động end-to-end")
				assert.NotContains(t, bodyStr, "Passw0rd!", "plaintext password KHÔNG được leak")
				assert.NotContains(t, data, "password_hash")
				assert.NotContains(t, data, "password")
			},
		},

		// ─── TC-L6-I002 — 400 Email format sai ─────────────────────
		{
			name: "I002 — 400 email format invalid",
			body: map[string]any{
				"username": "u002", "email": "not-an-email",
				"password": "Passw0rd!", "display_name": "U",
			},
			wantCode: http.StatusBadRequest,
		},

		// ─── TC-L6-I003 — 400 Password quá ngắn ────────────────────
		{
			name: "I003 — 400 password too short (7 chars)",
			body: map[string]any{
				"username": "u003", "email": "u003@example.com",
				"password": "1234567", "display_name": "U",
			},
			wantCode: http.StatusBadRequest,
		},

		// ─── TC-L6-I004 — 400 Username chứa space ──────────────────
		{
			name: "I004 — 400 username has space (not alphanum)",
			body: map[string]any{
				"username": "al ice", "email": "u004@example.com",
				"password": "Passw0rd!", "display_name": "U",
			},
			wantCode: http.StatusBadRequest,
		},

		// ─── TC-L6-I005 — 409 Duplicate email (case-insensitive) ───
		{
			name: "I005 — 409 duplicate email (uppercase normalize)",
			body: map[string]any{
				"username": "alice2", "email": "ALICE@example.com", // normalize → "alice@example.com"
				"password": "Passw0rd!", "display_name": "Alice2",
			},
			wantCode: http.StatusConflict,
			verify: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "email already registered", resp["error"])
			},
		},

		// ─── TC-L6-I006 — 409 Duplicate username ───────────────────
		{
			name: "I006 — 409 duplicate username (alice từ I001)",
			body: map[string]any{
				"username": "alice", "email": "new@example.com",
				"password": "Passw0rd!", "display_name": "New",
			},
			wantCode: http.StatusConflict,
			verify: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "username already taken", resp["error"])
			},
		},

		// ─── TC-L6-I007 — 400 Empty body {} ────────────────────────
		{
			name:     "I007 — 400 empty body",
			body:     map[string]any{},
			wantCode: http.StatusBadRequest,
		},

		// ─── TC-L6-I008 — 400 Empty string fields ──────────────────
		{
			name: "I008 — 400 empty username string",
			body: map[string]any{
				"username": "", "email": "u008@example.com",
				"password": "Passw0rd!", "display_name": "U",
			},
			wantCode: http.StatusBadRequest,
		},

		// ─── TC-L6-I009 — Email mixed-case → lowercase persisted → 201 ─
		// Note: KHÔNG dùng leading/trailing spaces vì Gin binding `email`
		// validator reject trước khi vào service (400). Spaces normalize
		// chỉ có thể test nếu handler normalize TRƯỚC bind — defer Lec-7+.
		// Test này verify case normalization persisted ở email field
		// (I005 chỉ indirect verify qua dup detection).
		{
			name: "I009 — 201 email mixed-case lowercased on persist",
			body: map[string]any{
				"username": "bob", "email": "Bob@Example.com",
				"password": "Passw0rd!", "display_name": "Bob",
			},
			wantCode: http.StatusCreated,
			verify: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "bob@example.com", data["email"], "lowercased on persist")
			},
		},

		// ─── TC-L6-I010 — Password đúng 72 bytes → 201 (bcrypt max) ─
		{
			name: "I010 — 201 password 72 bytes (bcrypt max input)",
			body: map[string]any{
				"username": "u010", "email": "u010@example.com",
				"password": pwd72, "display_name": "U10",
			},
			wantCode: http.StatusCreated,
		},

		// ─── TC-L6-I011 — Password 7 chars → 400 (dưới min=8) ──────
		{
			name: "I011 — 400 password exactly 7 chars (boundary below min)",
			body: map[string]any{
				"username": "u011", "email": "u011@example.com",
				"password": "1234567", "display_name": "U11",
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// KHÔNG t.Parallel() — shared DB ordering dependency (I005/I006 cần I001 trước)
			b, err := json.Marshal(tc.body)
			require.NoError(t, err, "marshal request body")

			req := httptest.NewRequest(http.MethodPost, "/v1/users/register", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)

			require.Equal(t, tc.wantCode, w.Code, "case %q: unexpected status code; body=%s", tc.name, w.Body.String())
			if tc.verify != nil {
				tc.verify(t, w.Body.Bytes())
			}
		})
	}
}
