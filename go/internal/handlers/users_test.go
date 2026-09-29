package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/users-api-go/internal/database"
	"github.com/users-api-go/internal/models"
)

func setupTestRouter(t *testing.T) (*chi.Mux, func()) {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := database.NewDB(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create database: %v", err)
	}

	repo := database.NewUserRepository(db)
	handler := NewUserHandler(repo)

	r := chi.NewRouter()
	r.Post("/users", handler.Create)
	r.Get("/users", handler.List)
	r.Get("/users/{id}", handler.GetByID)
	r.Put("/users/{id}", handler.Update)
	r.Patch("/users/{id}", handler.Update)
	r.Delete("/users/{id}", handler.Delete)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return r, cleanup
}

func TestCreateUser(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid user",
			body: map[string]string{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "password123",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "missing name",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "missing email",
			body: map[string]string{
				"name":     "Test User",
				"password": "password123",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "invalid email",
			body: map[string]string{
				"name":     "Test User",
				"email":    "invalid-email",
				"password": "password123",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "short password",
			body: map[string]string{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": "123",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			if !tt.wantError {
				var user models.User
				if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if user.ID == 0 {
					t.Error("Expected user ID to be set")
				}
				if user.Name != tt.body["name"] {
					t.Errorf("got name %q, want %q", user.Name, tt.body["name"])
				}
				if user.Email != tt.body["email"] {
					t.Errorf("got email %q, want %q", user.Email, tt.body["email"])
				}
			}
		})
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	body := map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(body)

	req1 := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusCreated {
		t.Fatalf("First request failed: %d", rec1.Code)
	}

	body["name"] = "Jane Doe"
	bodyBytes, _ = json.Marshal(body)
	req2 := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for duplicate email, got %d", rec2.Code)
	}
}

func TestListUsers(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	for i := 1; i <= 15; i++ {
		body := map[string]string{
			"name":     fmt.Sprintf("User %d", i),
			"email":    fmt.Sprintf("user%d@example.com", i),
			"password": "password123",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}

	tests := []struct {
		name       string
		query      string
		wantCount  int
		wantTotal  int64
		wantPages  int
		wantPage   int
		wantLimit  int
	}{
		{
			name:      "default pagination",
			query:     "",
			wantCount: 10,
			wantTotal: 15,
			wantPages: 2,
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:      "second page",
			query:     "?page=2",
			wantCount: 5,
			wantTotal: 15,
			wantPages: 2,
			wantPage:  2,
			wantLimit: 10,
		},
		{
			name:      "custom limit",
			query:     "?limit=5",
			wantCount: 5,
			wantTotal: 15,
			wantPages: 3,
			wantPage:  1,
			wantLimit: 5,
		},
		{
			name:      "page and limit",
			query:     "?page=2&limit=5",
			wantCount: 5,
			wantTotal: 15,
			wantPages: 3,
			wantPage:  2,
			wantLimit: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.query, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("got status %d, want 200", rec.Code)
			}

			var resp models.UsersListResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if len(resp.Data) != tt.wantCount {
				t.Errorf("got %d users, want %d", len(resp.Data), tt.wantCount)
			}
			if resp.Pagination.Total != tt.wantTotal {
				t.Errorf("got total %d, want %d", resp.Pagination.Total, tt.wantTotal)
			}
			if resp.Pagination.TotalPages != tt.wantPages {
				t.Errorf("got totalPages %d, want %d", resp.Pagination.TotalPages, tt.wantPages)
			}
			if resp.Pagination.Page != tt.wantPage {
				t.Errorf("got page %d, want %d", resp.Pagination.Page, tt.wantPage)
			}
			if resp.Pagination.Limit != tt.wantLimit {
				t.Errorf("got limit %d, want %d", resp.Pagination.Limit, tt.wantLimit)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	body := map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var createdUser models.User
	json.NewDecoder(rec.Body).Decode(&createdUser)

	t.Run("existing user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", createdUser.ID), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", rec.Code)
		}

		var user models.User
		json.NewDecoder(rec.Body).Decode(&user)
		if user.ID != createdUser.ID {
			t.Errorf("got ID %d, want %d", user.ID, createdUser.ID)
		}
		if user.Name != "John Doe" {
			t.Errorf("got name %q, want %q", user.Name, "John Doe")
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/99999", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("got status %d, want 404", rec.Code)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/invalid", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", rec.Code)
		}
	})
}

func TestUpdateUser(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	body := map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var createdUser models.User
	json.NewDecoder(rec.Body).Decode(&createdUser)

	t.Run("PUT full update", func(t *testing.T) {
		updateBody := map[string]string{
			"name":  "Jane Doe",
			"email": "jane@example.com",
		}
		bodyBytes, _ := json.Marshal(updateBody)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d", createdUser.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", rec.Code)
		}

		var user models.User
		json.NewDecoder(rec.Body).Decode(&user)
		if user.Name != "Jane Doe" {
			t.Errorf("got name %q, want %q", user.Name, "Jane Doe")
		}
		if user.Email != "jane@example.com" {
			t.Errorf("got email %q, want %q", user.Email, "jane@example.com")
		}
	})

	t.Run("PATCH partial update", func(t *testing.T) {
		updateBody := map[string]string{
			"name": "John Updated",
		}
		bodyBytes, _ := json.Marshal(updateBody)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%d", createdUser.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", rec.Code)
		}

		var user models.User
		json.NewDecoder(rec.Body).Decode(&user)
		if user.Name != "John Updated" {
			t.Errorf("got name %q, want %q", user.Name, "John Updated")
		}
	})

	t.Run("update non-existent user", func(t *testing.T) {
		updateBody := map[string]string{"name": "Test"}
		bodyBytes, _ := json.Marshal(updateBody)
		req := httptest.NewRequest(http.MethodPut, "/users/99999", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("got status %d, want 404", rec.Code)
		}
	})

	t.Run("update with invalid email", func(t *testing.T) {
		updateBody := map[string]string{"email": "invalid"}
		bodyBytes, _ := json.Marshal(updateBody)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%d", createdUser.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", rec.Code)
		}
	})
}

func TestDeleteUser(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	body := map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var createdUser models.User
	json.NewDecoder(rec.Body).Decode(&createdUser)

	t.Run("delete existing user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", createdUser.ID), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", rec.Code)
		}

		getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", createdUser.ID), nil)
		getRec := httptest.NewRecorder()
		r.ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusNotFound {
			t.Errorf("Expected user to be deleted, got status %d", getRec.Code)
		}
	})

	t.Run("delete non-existent user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/99999", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("got status %d, want 404", rec.Code)
		}
	})
}

func TestPasswordNotReturned(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	body := map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var rawResponse map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&rawResponse)

	if _, exists := rawResponse["password"]; exists {
		t.Error("Password should not be returned in create response")
	}

	var createdUser models.User
	json.Unmarshal(rec.Body.Bytes(), &createdUser)

	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", int(rawResponse["id"].(float64))), nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)

	var getResponse map[string]interface{}
	json.NewDecoder(getRec.Body).Decode(&getResponse)

	if _, exists := getResponse["password"]; exists {
		t.Error("Password should not be returned in get response")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/users", nil)
	listRec := httptest.NewRecorder()
	r.ServeHTTP(listRec, listReq)

	var listResponse map[string]interface{}
	json.NewDecoder(listRec.Body).Decode(&listResponse)

	data := listResponse["data"].([]interface{})
	if len(data) > 0 {
		user := data[0].(map[string]interface{})
		if _, exists := user["password"]; exists {
			t.Error("Password should not be returned in list response")
		}
	}
}
