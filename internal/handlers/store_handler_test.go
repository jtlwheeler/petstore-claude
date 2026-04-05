package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/jtlwheeler/petstore/internal/models"
)

func TestPlaceOrder(t *testing.T) {
	order := models.Order{
		PetID:    42,
		Quantity: 3,
		Status:   "placed",
		Complete: false,
	}

	got := doPlaceOrder(t, order)
	if got.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if got.PetID != order.PetID {
		t.Errorf("PetID: got %d, want %d", got.PetID, order.PetID)
	}
	if got.Quantity != order.Quantity {
		t.Errorf("Quantity: got %d, want %d", got.Quantity, order.Quantity)
	}
	if got.Status != order.Status {
		t.Errorf("Status: got %q, want %q", got.Status, order.Status)
	}
}

func TestGetOrderByID(t *testing.T) {
	created := doPlaceOrder(t, models.Order{PetID: 10, Quantity: 1, Status: "placed"})

	resp, err := http.Get(fmt.Sprintf("%s/api/v3/store/order/%d", testServer.URL, created.ID))
	if err != nil {
		t.Fatalf("GET /store/order/{id}: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got models.Order
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID: got %d, want %d", got.ID, created.ID)
	}
}

func TestGetOrderByID_NotFound(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v3/store/order/999999999", testServer.URL))
	if err != nil {
		t.Fatalf("GET /store/order/999999999: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteOrder(t *testing.T) {
	created := doPlaceOrder(t, models.Order{PetID: 5, Quantity: 1, Status: "placed"})

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v3/store/order/%d", testServer.URL, created.ID), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /store/order/{id}: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	getResp, err := http.Get(fmt.Sprintf("%s/api/v3/store/order/%d", testServer.URL, created.ID))
	if err != nil {
		t.Fatalf("GET after delete: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getResp.StatusCode)
	}
}

func TestGetInventory(t *testing.T) {
	doAddPet(t, models.Pet{Name: "Inv_H_Available", PhotoUrls: []string{"u"}, Status: "available"})
	doAddPet(t, models.Pet{Name: "Inv_H_Sold", PhotoUrls: []string{"u"}, Status: "sold"})

	resp, err := http.Get(testServer.URL + "/api/v3/store/inventory")
	if err != nil {
		t.Fatalf("GET /store/inventory: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var inventory map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&inventory); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := inventory["available"]; !ok {
		t.Error("expected 'available' key in inventory")
	}
	if _, ok := inventory["sold"]; !ok {
		t.Error("expected 'sold' key in inventory")
	}
}

// doPlaceOrder is a test helper that POSTs an order and returns the created resource.
func doPlaceOrder(t *testing.T, order models.Order) models.Order {
	t.Helper()
	body, _ := json.Marshal(order)
	resp, err := http.Post(testServer.URL+"/api/v3/store/order", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /store/order: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /store/order: expected 200, got %d", resp.StatusCode)
	}
	var created models.Order
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created order: %v", err)
	}
	return created
}
