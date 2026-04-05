package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jtlwheeler/petstore/internal/repository"
)

// SetupRoutes wires all handlers to routes and returns the HTTP handler.
func SetupRoutes(
	pool *pgxpool.Pool,
	petRepo *repository.PetRepository,
	orderRepo *repository.OrderRepository,
	userRepo *repository.UserRepository,
) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	health := &healthHandler{pool: pool}
	r.Get("/readyz", health.readyz)

	petHandler := NewPetHandler(petRepo)
	storeHandler := NewStoreHandler(orderRepo)
	userHandler := NewUserHandler(userRepo)

	r.Route("/api/v3", func(r chi.Router) {
		// Pet routes — static paths first, then parameterized
		r.Put("/pet", petHandler.UpdatePet)
		r.Post("/pet", petHandler.AddPet)
		r.Get("/pet/findByStatus", petHandler.FindByStatus)
		r.Get("/pet/findByTags", petHandler.FindByTags)
		r.Get("/pet/{petId}", petHandler.GetPetByID)
		r.Post("/pet/{petId}", petHandler.UpdatePetWithForm)
		r.Delete("/pet/{petId}", petHandler.DeletePet)
		r.Post("/pet/{petId}/uploadImage", petHandler.UploadImage)

		// Store routes
		r.Get("/store/inventory", storeHandler.GetInventory)
		r.Post("/store/order", storeHandler.PlaceOrder)
		r.Get("/store/order/{orderId}", storeHandler.GetOrderByID)
		r.Delete("/store/order/{orderId}", storeHandler.DeleteOrder)

		// User routes — static paths first
		r.Post("/user", userHandler.CreateUser)
		r.Post("/user/createWithList", userHandler.CreateUsersWithList)
		r.Get("/user/login", userHandler.LoginUser)
		r.Get("/user/logout", userHandler.LogoutUser)
		r.Get("/user/{username}", userHandler.GetUserByName)
		r.Put("/user/{username}", userHandler.UpdateUser)
		r.Delete("/user/{username}", userHandler.DeleteUser)
	})

	return r
}
