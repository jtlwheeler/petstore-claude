package models

import "time"

// Category represents a pet category.
type Category struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Tag represents a pet tag.
type Tag struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Pet represents a pet in the store.
type Pet struct {
	ID        int64     `json:"id,omitempty"`
	Name      string    `json:"name"`
	Category  *Category `json:"category,omitempty"`
	PhotoUrls []string  `json:"photoUrls"`
	Tags      []Tag     `json:"tags,omitempty"`
	Status    string    `json:"status,omitempty"`
}

// Order represents a purchase order.
type Order struct {
	ID       int64      `json:"id,omitempty"`
	PetID    int64      `json:"petId,omitempty"`
	Quantity int32      `json:"quantity,omitempty"`
	ShipDate *time.Time `json:"shipDate,omitempty"`
	Status   string     `json:"status,omitempty"`
	Complete bool       `json:"complete"`
}

// User represents a user of the petstore.
type User struct {
	ID         int64  `json:"id,omitempty"`
	Username   string `json:"username,omitempty"`
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	Email      string `json:"email,omitempty"`
	Password   string `json:"password,omitempty"`
	Phone      string `json:"phone,omitempty"`
	UserStatus int32  `json:"userStatus,omitempty"`
}

// ApiResponse represents a generic API response.
type ApiResponse struct {
	Code    int32  `json:"code,omitempty"`
	Type    string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
}
