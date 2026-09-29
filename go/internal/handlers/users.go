package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/users-api-go/internal/database"
	"github.com/users-api-go/internal/models"
)

type UserHandler struct {
	repo *database.UserRepository
}

func NewUserHandler(repo *database.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := validateCreateRequest(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user, err := h.repo.Create(req.Name, req.Email, string(hashedPassword))
	if errors.Is(err, database.ErrDuplicateEmail) {
		respondError(w, http.StatusBadRequest, "Email already exists")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page := parseIntParam(r.URL.Query().Get("page"), 1)
	limit := parseIntParam(r.URL.Query().Get("limit"), 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := h.repo.List(page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	if users == nil {
		users = []models.User{}
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	response := models.UsersListResponse{
		Data: users,
		Pagination: models.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.repo.GetByID(id)
	if errors.Is(err, database.ErrNotFound) {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := validateUpdateRequest(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var passwordHash *string
	if req.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		h := string(hashed)
		passwordHash = &h
	}

	user, err := h.repo.Update(id, req.Name, req.Email, passwordHash)
	if errors.Is(err, database.ErrNotFound) {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	if errors.Is(err, database.ErrDuplicateEmail) {
		respondError(w, http.StatusBadRequest, "Email already exists")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.repo.Delete(id)
	if errors.Is(err, database.ErrNotFound) {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func validateCreateRequest(req models.CreateUserRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email is required")
	}
	if !isValidEmail(req.Email) {
		return errors.New("invalid email format")
	}
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

func validateUpdateRequest(req models.UpdateUserRequest) error {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if req.Email != nil {
		if strings.TrimSpace(*req.Email) == "" {
			return errors.New("email cannot be empty")
		}
		if !isValidEmail(*req.Email) {
			return errors.New("invalid email format")
		}
	}
	if req.Password != nil && len(*req.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, models.ErrorResponse{Error: message})
}
