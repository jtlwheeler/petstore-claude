package repository_test

import (
	"context"
	"testing"

	"github.com/jtlwheeler/petstore/internal/models"
)

func TestUserCRUD(t *testing.T) {
	ctx := context.Background()

	user := models.User{
		Username:   "crud_user",
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@example.com",
		Password:   "secret",
		Phone:      "555-0100",
		UserStatus: 1,
	}

	created, err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
	t.Cleanup(func() { userRepo.Delete(ctx, created.Username) }) //nolint:errcheck

	got, err := userRepo.GetByUsername(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if got.Username != user.Username {
		t.Errorf("Username: got %q, want %q", got.Username, user.Username)
	}
	if got.Email != user.Email {
		t.Errorf("Email: got %q, want %q", got.Email, user.Email)
	}

	updated := models.User{
		Username:  "crud_user",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Password:  "newpassword",
		Phone:     "555-0200",
	}
	if err := userRepo.Update(ctx, user.Username, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	gotUpdated, err := userRepo.GetByUsername(ctx, "crud_user")
	if err != nil {
		t.Fatalf("GetByUsername after update: %v", err)
	}
	if gotUpdated.FirstName != "Jane" {
		t.Errorf("FirstName: got %q, want Jane", gotUpdated.FirstName)
	}
	if gotUpdated.Email != "jane@example.com" {
		t.Errorf("Email: got %q, want jane@example.com", gotUpdated.Email)
	}

	if err := userRepo.Delete(ctx, "crud_user"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = userRepo.GetByUsername(ctx, "crud_user")
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestUserCreateBatch(t *testing.T) {
	ctx := context.Background()

	users := []models.User{
		{Username: "batch_user_1", Password: "p1"},
		{Username: "batch_user_2", Password: "p2"},
		{Username: "batch_user_3", Password: "p3"},
	}

	created, err := userRepo.CreateBatch(ctx, users)
	if err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}
	for _, u := range created {
		t.Cleanup(func() { userRepo.Delete(ctx, u.Username) }) //nolint:errcheck
	}

	if len(created) != len(users) {
		t.Fatalf("expected %d users, got %d", len(users), len(created))
	}
	for i, u := range created {
		if u.ID == 0 {
			t.Errorf("user[%d] has zero ID", i)
		}
		if u.Username != users[i].Username {
			t.Errorf("user[%d] Username: got %q, want %q", i, u.Username, users[i].Username)
		}
	}

	for _, u := range users {
		if _, err := userRepo.GetByUsername(ctx, u.Username); err != nil {
			t.Errorf("GetByUsername %q: %v", u.Username, err)
		}
	}
}

func TestUserLogin(t *testing.T) {
	ctx := context.Background()

	user := models.User{
		Username: "login_user",
		Password: "correctpassword",
	}
	_, err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { userRepo.Delete(ctx, user.Username) }) //nolint:errcheck

	got, err := userRepo.Login(ctx, user.Username, user.Password)
	if err != nil {
		t.Fatalf("Login with valid credentials: %v", err)
	}
	if got.Username != user.Username {
		t.Errorf("Username: got %q, want %q", got.Username, user.Username)
	}
}

func TestUserLogin_InvalidPassword(t *testing.T) {
	ctx := context.Background()

	user := models.User{
		Username: "login_invalid_user",
		Password: "realpassword",
	}
	_, err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { userRepo.Delete(ctx, user.Username) }) //nolint:errcheck

	_, err = userRepo.Login(ctx, user.Username, "wrongpassword")
	if err == nil {
		t.Fatal("expected error for invalid password, got nil")
	}
}

func TestUserGetByUsername_NotFound(t *testing.T) {
	_, err := userRepo.GetByUsername(context.Background(), "nonexistent_user_xyz")
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
