package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/jtlwheeler/petstore/internal/models"
)

func TestCreateUser(t *testing.T) {
	user := models.User{
		Username:  "handler_create_user",
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		Password:  "pass123",
	}

	got := doCreateUser(t, user)
	if got.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if got.Username != user.Username {
		t.Errorf("Username: got %q, want %q", got.Username, user.Username)
	}
}

func TestGetUserByName(t *testing.T) {
	created := doCreateUser(t, models.User{
		Username: "handler_get_user",
		Email:    "get@example.com",
		Password: "pass",
	})

	resp, err := http.Get(fmt.Sprintf("%s/api/v3/user/%s", testServer.URL, created.Username))
	if err != nil {
		t.Fatalf("GET /user/{username}: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got models.User
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Username != created.Username {
		t.Errorf("Username: got %q, want %q", got.Username, created.Username)
	}
	if got.Email != created.Email {
		t.Errorf("Email: got %q, want %q", got.Email, created.Email)
	}
}

func TestUpdateUser(t *testing.T) {
	created := doCreateUser(t, models.User{
		Username: "handler_update_user",
		Email:    "before@example.com",
		Password: "pass",
	})

	updated := models.User{
		Username: "handler_update_user",
		Email:    "after@example.com",
		Password: "newpass",
	}
	body, _ := json.Marshal(updated)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v3/user/%s", testServer.URL, created.Username), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /user/{username}: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	getResp, err := http.Get(fmt.Sprintf("%s/api/v3/user/%s", testServer.URL, created.Username))
	if err != nil {
		t.Fatalf("GET after update: %v", err)
	}
	defer getResp.Body.Close()

	var got models.User
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Email != "after@example.com" {
		t.Errorf("Email after update: got %q, want after@example.com", got.Email)
	}
}

func TestDeleteUser(t *testing.T) {
	created := doCreateUser(t, models.User{
		Username: "handler_delete_user",
		Password: "pass",
	})

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v3/user/%s", testServer.URL, created.Username), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /user/{username}: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	getResp, err := http.Get(fmt.Sprintf("%s/api/v3/user/%s", testServer.URL, created.Username))
	if err != nil {
		t.Fatalf("GET after delete: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getResp.StatusCode)
	}
}

func TestCreateUsersWithList(t *testing.T) {
	users := []models.User{
		{Username: "handler_list_user_1", Password: "p1"},
		{Username: "handler_list_user_2", Password: "p2"},
		{Username: "handler_list_user_3", Password: "p3"},
	}

	body, _ := json.Marshal(users)
	resp, err := http.Post(testServer.URL+"/api/v3/user/createWithList", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /user/createWithList: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	for _, u := range users {
		getResp, err := http.Get(fmt.Sprintf("%s/api/v3/user/%s", testServer.URL, u.Username))
		if err != nil {
			t.Fatalf("GET user %q: %v", u.Username, err)
		}
		getResp.Body.Close()
		if getResp.StatusCode != http.StatusOK {
			t.Errorf("user %q: expected 200, got %d", u.Username, getResp.StatusCode)
		}
	}
}

func TestLoginUser(t *testing.T) {
	user := models.User{
		Username: "handler_login_user",
		Password: "loginpass",
	}
	doCreateUser(t, user)

	resp, err := http.Get(fmt.Sprintf("%s/api/v3/user/login?username=%s&password=%s", testServer.URL, user.Username, user.Password))
	if err != nil {
		t.Fatalf("GET /user/login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("X-Rate-Limit") != "1000" {
		t.Errorf("X-Rate-Limit: got %q, want 1000", resp.Header.Get("X-Rate-Limit"))
	}
	if resp.Header.Get("X-Expires-After") == "" {
		t.Error("expected X-Expires-After header to be set")
	}
}

// doCreateUser is a test helper that POSTs a user and returns the created resource.
func doCreateUser(t *testing.T, user models.User) models.User {
	t.Helper()
	body, _ := json.Marshal(user)
	resp, err := http.Post(testServer.URL+"/api/v3/user", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /user: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /user: expected 200, got %d", resp.StatusCode)
	}
	var created models.User
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created user: %v", err)
	}
	return created
}
