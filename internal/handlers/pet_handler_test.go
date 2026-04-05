package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/jtlwheeler/petstore/internal/models"
)

func TestAddPet(t *testing.T) {
	pet := models.Pet{
		Name:      "Handler_AddDog",
		PhotoUrls: []string{"https://example.com/add.jpg"},
		Status:    "available",
	}

	got := doAddPet(t, pet)
	if got.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if got.Name != pet.Name {
		t.Errorf("Name: got %q, want %q", got.Name, pet.Name)
	}
	if got.Status != pet.Status {
		t.Errorf("Status: got %q, want %q", got.Status, pet.Status)
	}
}

func TestGetPetByID(t *testing.T) {
	created := doAddPet(t, models.Pet{
		Name:      "Handler_GetDog",
		PhotoUrls: []string{"https://example.com/get.jpg"},
		Status:    "pending",
		Category:  &models.Category{Name: "Dogs"},
		Tags:      []models.Tag{{Name: "handler-test-tag"}},
	})

	resp, err := http.Get(fmt.Sprintf("%s/api/v3/pet/%d", testServer.URL, created.ID))
	if err != nil {
		t.Fatalf("GET /pet/{id}: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got models.Pet
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID: got %d, want %d", got.ID, created.ID)
	}
	if got.Name != created.Name {
		t.Errorf("Name: got %q, want %q", got.Name, created.Name)
	}
}

func TestGetPetByID_NotFound(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v3/pet/999999999", testServer.URL))
	if err != nil {
		t.Fatalf("GET /pet/999999999: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdatePet(t *testing.T) {
	created := doAddPet(t, models.Pet{
		Name:      "Handler_UpdateDog_Before",
		PhotoUrls: []string{"https://example.com/before.jpg"},
		Status:    "available",
	})

	updated := models.Pet{
		ID:        created.ID,
		Name:      "Handler_UpdateDog_After",
		PhotoUrls: []string{"https://example.com/after.jpg"},
		Status:    "sold",
	}

	body, _ := json.Marshal(updated)
	req, _ := http.NewRequest(http.MethodPut, testServer.URL+"/api/v3/pet", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /pet: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got models.Pet
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != updated.Name {
		t.Errorf("Name: got %q, want %q", got.Name, updated.Name)
	}
	if got.Status != updated.Status {
		t.Errorf("Status: got %q, want %q", got.Status, updated.Status)
	}
}

func TestDeletePet(t *testing.T) {
	created := doAddPet(t, models.Pet{
		Name:      "Handler_DeleteDog",
		PhotoUrls: []string{"https://example.com/del.jpg"},
		Status:    "available",
	})

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v3/pet/%d", testServer.URL, created.ID), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /pet/{id}: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	getResp, err := http.Get(fmt.Sprintf("%s/api/v3/pet/%d", testServer.URL, created.ID))
	if err != nil {
		t.Fatalf("GET after delete: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getResp.StatusCode)
	}
}

func TestFindByStatus(t *testing.T) {
	doAddPet(t, models.Pet{Name: "FindStatus_H_Available", PhotoUrls: []string{"u"}, Status: "available"})
	doAddPet(t, models.Pet{Name: "FindStatus_H_Sold", PhotoUrls: []string{"u"}, Status: "sold"})

	resp, err := http.Get(testServer.URL + "/api/v3/pet/findByStatus?status=available")
	if err != nil {
		t.Fatalf("GET /pet/findByStatus: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var pets []models.Pet
	if err := json.NewDecoder(resp.Body).Decode(&pets); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, p := range pets {
		if p.Status != "available" {
			t.Errorf("expected all pets to have status 'available', got %q for pet %d", p.Status, p.ID)
		}
	}
}

func TestFindByTags(t *testing.T) {
	doAddPet(t, models.Pet{
		Name:      "FindTags_H_Dog",
		PhotoUrls: []string{"u"},
		Status:    "available",
		Tags:      []models.Tag{{Name: "handler-unique-tag-abc"}},
	})

	resp, err := http.Get(testServer.URL + "/api/v3/pet/findByTags?tags=handler-unique-tag-abc")
	if err != nil {
		t.Fatalf("GET /pet/findByTags: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var pets []models.Pet
	if err := json.NewDecoder(resp.Body).Decode(&pets); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(pets) == 0 {
		t.Error("expected at least one pet with tag 'handler-unique-tag-abc'")
	}
}

// doAddPet is a test helper that POSTs a pet and returns the created resource.
func doAddPet(t *testing.T, pet models.Pet) models.Pet {
	t.Helper()
	body, _ := json.Marshal(pet)
	resp, err := http.Post(testServer.URL+"/api/v3/pet", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /pet: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /pet: expected 200, got %d", resp.StatusCode)
	}
	var created models.Pet
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created pet: %v", err)
	}
	return created
}
