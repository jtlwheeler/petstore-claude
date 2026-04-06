package repository_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jtlwheeler/petstore/internal/db"
	"github.com/jtlwheeler/petstore/internal/models"
	"github.com/jtlwheeler/petstore/internal/repository"
	"github.com/jtlwheeler/petstore/internal/db/migrations"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	petRepo   *repository.PetRepository
	orderRepo *repository.OrderRepository
	userRepo  *repository.UserRepository
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("petstore_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		panic("starting postgres container: " + err.Error())
	}

	defer func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pgContainer.Terminate(timeoutCtx) //nolint:errcheck
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("getting connection string: " + err.Error())
	}

	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		panic("connecting to database: " + err.Error())
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, migrations.FS); err != nil {
		panic("running migrations: " + err.Error())
	}

	petRepo = repository.NewPetRepository(pool)
	orderRepo = repository.NewOrderRepository(pool)
	userRepo = repository.NewUserRepository(pool)

	os.Exit(m.Run())
}

func isNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}

func TestPetCRUD(t *testing.T) {
	ctx := context.Background()

	pet := models.Pet{
		Name: "TestDog",
		Category: &models.Category{
			Name: "Dogs",
		},
		PhotoUrls: []string{"https://example.com/dog.jpg"},
		Tags: []models.Tag{
			{Name: "friendly"},
			{Name: "small"},
		},
		Status: "available",
	}

	created, err := petRepo.Create(ctx, pet)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
	t.Cleanup(func() { petRepo.Delete(ctx, created.ID) }) //nolint:errcheck

	got, err := petRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != pet.Name {
		t.Errorf("Name: got %q, want %q", got.Name, pet.Name)
	}
	if got.Status != pet.Status {
		t.Errorf("Status: got %q, want %q", got.Status, pet.Status)
	}
	if got.Category == nil {
		t.Fatal("expected category, got nil")
	}
	if got.Category.Name != pet.Category.Name {
		t.Errorf("Category.Name: got %q, want %q", got.Category.Name, pet.Category.Name)
	}
	if len(got.PhotoUrls) != len(pet.PhotoUrls) {
		t.Errorf("PhotoUrls length: got %d, want %d", len(got.PhotoUrls), len(pet.PhotoUrls))
	}
	if len(got.Tags) != len(pet.Tags) {
		t.Errorf("Tags length: got %d, want %d", len(got.Tags), len(pet.Tags))
	}

	if err := petRepo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = petRepo.GetByID(ctx, created.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPetUpdate(t *testing.T) {
	ctx := context.Background()

	original := models.Pet{
		Name:      "UpdateDog_Original",
		PhotoUrls: []string{"https://example.com/orig.jpg"},
		Status:    "available",
	}

	created, err := petRepo.Create(ctx, original)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { petRepo.Delete(ctx, created.ID) }) //nolint:errcheck

	updated := models.Pet{
		ID:   created.ID,
		Name: "UpdateDog_Changed",
		Category: &models.Category{
			Name: "Cats",
		},
		PhotoUrls: []string{"https://example.com/new.jpg"},
		Status:    "pending",
	}

	result, err := petRepo.Update(ctx, updated)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.Name != updated.Name {
		t.Errorf("Name: got %q, want %q", result.Name, updated.Name)
	}
	if result.Status != updated.Status {
		t.Errorf("Status: got %q, want %q", result.Status, updated.Status)
	}
	if result.Category == nil || result.Category.Name != "Cats" {
		t.Errorf("Category: got %v, want Cats", result.Category)
	}
}

func TestPetFindByStatus(t *testing.T) {
	ctx := context.Background()

	available := models.Pet{Name: "FindStatus_Available", PhotoUrls: []string{"u"}, Status: "available"}
	sold := models.Pet{Name: "FindStatus_Sold", PhotoUrls: []string{"u"}, Status: "sold"}

	a, err := petRepo.Create(ctx, available)
	if err != nil {
		t.Fatalf("Create available: %v", err)
	}
	t.Cleanup(func() { petRepo.Delete(ctx, a.ID) }) //nolint:errcheck

	s, err := petRepo.Create(ctx, sold)
	if err != nil {
		t.Fatalf("Create sold: %v", err)
	}
	t.Cleanup(func() { petRepo.Delete(ctx, s.ID) }) //nolint:errcheck

	availablePets, err := petRepo.FindByStatus(ctx, "available")
	if err != nil {
		t.Fatalf("FindByStatus available: %v", err)
	}
	if !containsPetID(availablePets, a.ID) {
		t.Errorf("expected available pet %d in results", a.ID)
	}
	if containsPetID(availablePets, s.ID) {
		t.Errorf("did not expect sold pet %d in available results", s.ID)
	}

	soldPets, err := petRepo.FindByStatus(ctx, "sold")
	if err != nil {
		t.Fatalf("FindByStatus sold: %v", err)
	}
	if !containsPetID(soldPets, s.ID) {
		t.Errorf("expected sold pet %d in results", s.ID)
	}
}

func TestPetFindByTags(t *testing.T) {
	ctx := context.Background()

	tagged := models.Pet{
		Name:      "FindTags_Dog",
		PhotoUrls: []string{"u"},
		Status:    "available",
		Tags:      []models.Tag{{Name: "unique-tag-xyz"}},
	}

	created, err := petRepo.Create(ctx, tagged)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { petRepo.Delete(ctx, created.ID) }) //nolint:errcheck

	results, err := petRepo.FindByTags(ctx, []string{"unique-tag-xyz"})
	if err != nil {
		t.Fatalf("FindByTags: %v", err)
	}
	if !containsPetID(results, created.ID) {
		t.Errorf("expected pet %d in tag results", created.ID)
	}

	noResults, err := petRepo.FindByTags(ctx, []string{"nonexistent-tag-abc"})
	if err != nil {
		t.Fatalf("FindByTags nonexistent: %v", err)
	}
	if len(noResults) != 0 {
		t.Errorf("expected empty results for nonexistent tag, got %d", len(noResults))
	}
}

func TestPetAddPhotoURL(t *testing.T) {
	ctx := context.Background()

	pet := models.Pet{Name: "AddPhoto_Dog", PhotoUrls: []string{"https://example.com/first.jpg"}, Status: "available"}
	created, err := petRepo.Create(ctx, pet)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { petRepo.Delete(ctx, created.ID) }) //nolint:errcheck

	_, err = petRepo.AddPhotoURL(ctx, created.ID, "https://example.com/second.jpg")
	if err != nil {
		t.Fatalf("AddPhotoURL: %v", err)
	}

	got, err := petRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(got.PhotoUrls) != 2 {
		t.Errorf("PhotoUrls length: got %d, want 2", len(got.PhotoUrls))
	}
}

func TestPetGetByID_NotFound(t *testing.T) {
	_, err := petRepo.GetByID(context.Background(), 999999999)
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPetDelete_NotFound(t *testing.T) {
	err := petRepo.Delete(context.Background(), 999999999)
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func containsPetID(pets []models.Pet, id int64) bool {
	for _, p := range pets {
		if p.ID == id {
			return true
		}
	}
	return false
}
