package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jtlwheeler/petstore/internal/models"
	"github.com/jtlwheeler/petstore/internal/repository"
)

// PetHandler handles HTTP requests for pet endpoints.
type PetHandler struct {
	repo *repository.PetRepository
}

// NewPetHandler creates a new PetHandler.
func NewPetHandler(repo *repository.PetRepository) *PetHandler {
	return &PetHandler{repo: repo}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}

// AddPet handles POST /pet
func (h *PetHandler) AddPet(w http.ResponseWriter, r *http.Request) {
	var pet models.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if pet.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	created, err := h.repo.Create(r.Context(), pet)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, created)
}

// UpdatePet handles PUT /pet
func (h *PetHandler) UpdatePet(w http.ResponseWriter, r *http.Request) {
	var pet models.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if pet.ID == 0 {
		writeError(w, http.StatusBadRequest, "pet ID is required")
		return
	}

	updated, err := h.repo.Update(r.Context(), pet)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "pet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// GetPetByID handles GET /pet/{petId}
func (h *PetHandler) GetPetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "petId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid pet ID")
		return
	}

	pet, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "pet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pet)
}

// UpdatePetWithForm handles POST /pet/{petId}
func (h *PetHandler) UpdatePetWithForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "petId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid pet ID")
		return
	}

	pet, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "pet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if name := r.URL.Query().Get("name"); name != "" {
		pet.Name = name
	}
	if status := r.URL.Query().Get("status"); status != "" {
		pet.Status = status
	}

	updated, err := h.repo.Update(r.Context(), pet)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "pet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// DeletePet handles DELETE /pet/{petId}
func (h *PetHandler) DeletePet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "petId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid pet ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "pet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "pet deleted"})
}

// FindByStatus handles GET /pet/findByStatus
func (h *PetHandler) FindByStatus(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		writeError(w, http.StatusBadRequest, "status query param is required")
		return
	}

	pets, err := h.repo.FindByStatus(r.Context(), status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pets)
}

// FindByTags handles GET /pet/findByTags
func (h *PetHandler) FindByTags(w http.ResponseWriter, r *http.Request) {
	rawTags := r.URL.Query()["tags"]
	if len(rawTags) == 0 {
		writeError(w, http.StatusBadRequest, "tags query param is required")
		return
	}

	// Support comma-separated values too
	var tags []string
	for _, t := range rawTags {
		for _, part := range strings.Split(t, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				tags = append(tags, part)
			}
		}
	}

	pets, err := h.repo.FindByTags(r.Context(), tags)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pets)
}

// UploadImage handles POST /pet/{petId}/uploadImage
func (h *PetHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "petId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid pet ID")
		return
	}

	// Verify pet exists
	if _, err := h.repo.GetByID(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "pet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil || len(data) == 0 {
		writeError(w, http.StatusBadRequest, "no file uploaded")
		return
	}

	dir := filepath.Join(".", "uploads", fmt.Sprintf("%d", id))
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create upload directory")
		return
	}

	filename := r.URL.Query().Get("additionalMetadata")
	if filename == "" {
		filename = "upload.bin"
	}
	filePath := filepath.Join(dir, filepath.Base(filename))

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to write file")
		return
	}

	resp, err := h.repo.AddPhotoURL(r.Context(), id, filePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
