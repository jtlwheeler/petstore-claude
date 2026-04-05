package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jtlwheeler/petstore/internal/models"
	"github.com/jtlwheeler/petstore/internal/repository"
)

// StoreHandler handles HTTP requests for store endpoints.
type StoreHandler struct {
	repo *repository.OrderRepository
}

// NewStoreHandler creates a new StoreHandler.
func NewStoreHandler(repo *repository.OrderRepository) *StoreHandler {
	return &StoreHandler{repo: repo}
}

// GetInventory handles GET /store/inventory
func (h *StoreHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	inventory, err := h.repo.GetInventory(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, inventory)
}

// PlaceOrder handles POST /store/order
func (h *StoreHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.repo.Create(r.Context(), order)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, created)
}

// GetOrderByID handles GET /store/order/{orderId}
func (h *StoreHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "orderId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	order, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, order)
}

// DeleteOrder handles DELETE /store/order/{orderId}
func (h *StoreHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "orderId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "order deleted"})
}
