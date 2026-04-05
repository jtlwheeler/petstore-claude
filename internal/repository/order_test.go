package repository_test

import (
	"context"
	"testing"

	"github.com/jtlwheeler/petstore/internal/models"
)

func TestOrderCRUD(t *testing.T) {
	ctx := context.Background()

	order := models.Order{
		PetID:    1,
		Quantity: 2,
		Status:   "placed",
		Complete: false,
	}

	created, err := orderRepo.Create(ctx, order)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
	t.Cleanup(func() { orderRepo.Delete(ctx, created.ID) }) //nolint:errcheck

	got, err := orderRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
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
	if got.Complete != order.Complete {
		t.Errorf("Complete: got %v, want %v", got.Complete, order.Complete)
	}

	if err := orderRepo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = orderRepo.GetByID(ctx, created.ID)
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestOrderGetByID_NotFound(t *testing.T) {
	_, err := orderRepo.GetByID(context.Background(), 999999999)
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestOrderDelete_NotFound(t *testing.T) {
	err := orderRepo.Delete(context.Background(), 999999999)
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetInventory(t *testing.T) {
	ctx := context.Background()

	availablePet := models.Pet{Name: "Inventory_Available", PhotoUrls: []string{"u"}, Status: "available"}
	pendingPet := models.Pet{Name: "Inventory_Pending", PhotoUrls: []string{"u"}, Status: "pending"}

	a, err := petRepo.Create(ctx, availablePet)
	if err != nil {
		t.Fatalf("Create available pet: %v", err)
	}
	t.Cleanup(func() { petRepo.Delete(ctx, a.ID) }) //nolint:errcheck

	p, err := petRepo.Create(ctx, pendingPet)
	if err != nil {
		t.Fatalf("Create pending pet: %v", err)
	}
	t.Cleanup(func() { petRepo.Delete(ctx, p.ID) }) //nolint:errcheck

	inventory, err := orderRepo.GetInventory(ctx)
	if err != nil {
		t.Fatalf("GetInventory: %v", err)
	}

	if _, ok := inventory["available"]; !ok {
		t.Error("expected 'available' key in inventory")
	}
	if _, ok := inventory["pending"]; !ok {
		t.Error("expected 'pending' key in inventory")
	}
	if inventory["available"] < 1 {
		t.Errorf("expected at least 1 available pet, got %d", inventory["available"])
	}
	if inventory["pending"] < 1 {
		t.Errorf("expected at least 1 pending pet, got %d", inventory["pending"])
	}
}
