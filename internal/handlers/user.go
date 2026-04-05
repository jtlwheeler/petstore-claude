package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jtlwheeler/petstore/internal/models"
	"github.com/jtlwheeler/petstore/internal/repository"
)

// UserHandler handles HTTP requests for user endpoints.
type UserHandler struct {
	repo *repository.UserRepository
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// CreateUser handles POST /user
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.repo.Create(r.Context(), user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, created)
}

// CreateUsersWithList handles POST /user/createWithList
func (h *UserHandler) CreateUsersWithList(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.repo.CreateBatch(r.Context(), users)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, created)
}

// LoginUser handles GET /user/login
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	if username == "" || password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	_, err := h.repo.Login(r.Context(), username, password)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusBadRequest, "invalid username/password")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid username/password")
		return
	}

	expiresAfter := time.Now().UTC().Add(time.Hour)
	w.Header().Set("X-Rate-Limit", "1000")
	w.Header().Set("X-Expires-After", expiresAfter.Format(time.RFC3339))
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged in"})
}

// LogoutUser handles GET /user/logout
func (h *UserHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// GetUserByName handles GET /user/{username}
func (h *UserHandler) GetUserByName(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := h.repo.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// UpdateUser handles PUT /user/{username}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.repo.Update(r.Context(), username, user); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user updated"})
}

// DeleteUser handles DELETE /user/{username}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	if err := h.repo.Delete(r.Context(), username); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}
