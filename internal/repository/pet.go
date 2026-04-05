package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jtlwheeler/petstore/internal/models"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

const petSelectQuery = `
SELECT p.id, p.name, p.status,
    c.id, c.name,
    COALESCE(array_agg(DISTINCT pu.url) FILTER (WHERE pu.url IS NOT NULL), '{}') as photo_urls,
    COALESCE(json_agg(DISTINCT jsonb_build_object('id', t.id, 'name', t.name)) FILTER (WHERE t.id IS NOT NULL), '[]') as tags
FROM pets p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN pet_photo_urls pu ON pu.pet_id = p.id
LEFT JOIN pet_tags pt ON pt.pet_id = p.id
LEFT JOIN tags t ON t.id = pt.tag_id
WHERE p.id = $1
GROUP BY p.id, p.name, p.status, c.id, c.name`

// PetRepository provides access to pet storage.
type PetRepository struct {
	pool *pgxpool.Pool
}

// NewPetRepository creates a new PetRepository.
func NewPetRepository(pool *pgxpool.Pool) *PetRepository {
	return &PetRepository{pool: pool}
}

// scanPet scans a single row from the aggregated pet query into a Pet struct.
func scanPet(row pgx.Row) (models.Pet, error) {
	var pet models.Pet
	var catID *int64
	var catName *string
	var photoURLs []string
	var tagsJSON []byte

	err := row.Scan(
		&pet.ID, &pet.Name, &pet.Status,
		&catID, &catName,
		&photoURLs,
		&tagsJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Pet{}, ErrNotFound
		}
		return models.Pet{}, err
	}

	if catID != nil {
		pet.Category = &models.Category{ID: *catID}
		if catName != nil {
			pet.Category.Name = *catName
		}
	}

	pet.PhotoUrls = photoURLs
	if pet.PhotoUrls == nil {
		pet.PhotoUrls = []string{}
	}

	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &pet.Tags); err != nil {
			return models.Pet{}, fmt.Errorf("unmarshal tags: %w", err)
		}
	}
	if pet.Tags == nil {
		pet.Tags = []models.Tag{}
	}

	return pet, nil
}

// upsertCategory inserts or finds an existing category by name and returns its ID.
func (r *PetRepository) upsertCategory(ctx context.Context, tx pgx.Tx, cat *models.Category) (int64, error) {
	if cat == nil || cat.Name == "" {
		return 0, nil
	}
	var id int64
	err := tx.QueryRow(ctx,
		`INSERT INTO categories (name) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id`,
		cat.Name,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Row already existed, fetch it
			err = tx.QueryRow(ctx, `SELECT id FROM categories WHERE name = $1`, cat.Name).Scan(&id)
			if err != nil {
				return 0, err
			}
			return id, nil
		}
		return 0, err
	}
	return id, nil
}

// upsertTag inserts or finds a tag by name and returns its ID.
func (r *PetRepository) upsertTag(ctx context.Context, tx pgx.Tx, tag models.Tag) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx,
		`INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`,
		tag.Name,
	).Scan(&id)
	return id, err
}

// Create inserts a new pet along with category, photo URLs, and tags.
func (r *PetRepository) Create(ctx context.Context, pet models.Pet) (models.Pet, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Pet{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var catID *int64
	if pet.Category != nil {
		id, err := r.upsertCategory(ctx, tx, pet.Category)
		if err != nil {
			return models.Pet{}, err
		}
		catID = &id
	}

	var petID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO pets (name, category_id, status) VALUES ($1, $2, $3) RETURNING id`,
		pet.Name, catID, pet.Status,
	).Scan(&petID)
	if err != nil {
		return models.Pet{}, err
	}

	for _, url := range pet.PhotoUrls {
		_, err = tx.Exec(ctx,
			`INSERT INTO pet_photo_urls (pet_id, url) VALUES ($1, $2)`,
			petID, url,
		)
		if err != nil {
			return models.Pet{}, err
		}
	}

	for _, tag := range pet.Tags {
		tagID, err := r.upsertTag(ctx, tx, tag)
		if err != nil {
			return models.Pet{}, err
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO pet_tags (pet_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			petID, tagID,
		)
		if err != nil {
			return models.Pet{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Pet{}, err
	}

	return r.GetByID(ctx, petID)
}

// Update updates an existing pet record and replaces photo URLs and tags.
func (r *PetRepository) Update(ctx context.Context, pet models.Pet) (models.Pet, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Pet{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var catID *int64
	if pet.Category != nil {
		id, err := r.upsertCategory(ctx, tx, pet.Category)
		if err != nil {
			return models.Pet{}, err
		}
		catID = &id
	}

	result, err := tx.Exec(ctx,
		`UPDATE pets SET name = $1, category_id = $2, status = $3 WHERE id = $4`,
		pet.Name, catID, pet.Status, pet.ID,
	)
	if err != nil {
		return models.Pet{}, err
	}
	if result.RowsAffected() == 0 {
		return models.Pet{}, ErrNotFound
	}

	// Replace photo URLs
	_, err = tx.Exec(ctx, `DELETE FROM pet_photo_urls WHERE pet_id = $1`, pet.ID)
	if err != nil {
		return models.Pet{}, err
	}
	for _, url := range pet.PhotoUrls {
		_, err = tx.Exec(ctx,
			`INSERT INTO pet_photo_urls (pet_id, url) VALUES ($1, $2)`,
			pet.ID, url,
		)
		if err != nil {
			return models.Pet{}, err
		}
	}

	// Replace tags
	_, err = tx.Exec(ctx, `DELETE FROM pet_tags WHERE pet_id = $1`, pet.ID)
	if err != nil {
		return models.Pet{}, err
	}
	for _, tag := range pet.Tags {
		tagID, err := r.upsertTag(ctx, tx, tag)
		if err != nil {
			return models.Pet{}, err
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO pet_tags (pet_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			pet.ID, tagID,
		)
		if err != nil {
			return models.Pet{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Pet{}, err
	}

	return r.GetByID(ctx, pet.ID)
}

// GetByID retrieves a pet by its ID.
func (r *PetRepository) GetByID(ctx context.Context, id int64) (models.Pet, error) {
	row := r.pool.QueryRow(ctx, petSelectQuery, id)
	return scanPet(row)
}

// FindByStatus returns all pets with the given status.
func (r *PetRepository) FindByStatus(ctx context.Context, status string) ([]models.Pet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name, p.status,
		    c.id, c.name,
		    COALESCE(array_agg(DISTINCT pu.url) FILTER (WHERE pu.url IS NOT NULL), '{}') as photo_urls,
		    COALESCE(json_agg(DISTINCT jsonb_build_object('id', t.id, 'name', t.name)) FILTER (WHERE t.id IS NOT NULL), '[]') as tags
		FROM pets p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN pet_photo_urls pu ON pu.pet_id = p.id
		LEFT JOIN pet_tags pt ON pt.pet_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.status = $1
		GROUP BY p.id, p.name, p.status, c.id, c.name`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectPets(rows)
}

// FindByTags returns all pets that have any of the given tags.
func (r *PetRepository) FindByTags(ctx context.Context, tags []string) ([]models.Pet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name, p.status,
		    c.id, c.name,
		    COALESCE(array_agg(DISTINCT pu.url) FILTER (WHERE pu.url IS NOT NULL), '{}') as photo_urls,
		    COALESCE(json_agg(DISTINCT jsonb_build_object('id', t.id, 'name', t.name)) FILTER (WHERE t.id IS NOT NULL), '[]') as tags
		FROM pets p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN pet_photo_urls pu ON pu.pet_id = p.id
		LEFT JOIN pet_tags pt ON pt.pet_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.id IN (
		    SELECT DISTINCT pt2.pet_id FROM pet_tags pt2
		    JOIN tags t2 ON t2.id = pt2.tag_id
		    WHERE t2.name = ANY($1)
		)
		GROUP BY p.id, p.name, p.status, c.id, c.name`, tags)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectPets(rows)
}

// collectPets reads all rows from a pgx.Rows and returns a slice of pets.
func collectPets(rows pgx.Rows) ([]models.Pet, error) {
	var pets []models.Pet
	for rows.Next() {
		var pet models.Pet
		var catID *int64
		var catName *string
		var photoURLs []string
		var tagsJSON []byte

		err := rows.Scan(
			&pet.ID, &pet.Name, &pet.Status,
			&catID, &catName,
			&photoURLs,
			&tagsJSON,
		)
		if err != nil {
			return nil, err
		}

		if catID != nil {
			pet.Category = &models.Category{ID: *catID}
			if catName != nil {
				pet.Category.Name = *catName
			}
		}

		pet.PhotoUrls = photoURLs
		if pet.PhotoUrls == nil {
			pet.PhotoUrls = []string{}
		}

		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &pet.Tags); err != nil {
				return nil, fmt.Errorf("unmarshal tags: %w", err)
			}
		}
		if pet.Tags == nil {
			pet.Tags = []models.Tag{}
		}

		pets = append(pets, pet)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if pets == nil {
		pets = []models.Pet{}
	}
	return pets, nil
}

// Delete removes a pet by ID.
func (r *PetRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM pets WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// AddPhotoURL saves a photo URL for a pet and returns an ApiResponse.
func (r *PetRepository) AddPhotoURL(ctx context.Context, petID int64, url string) (models.ApiResponse, error) {
	_, err := r.GetByID(ctx, petID)
	if err != nil {
		return models.ApiResponse{}, err
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO pet_photo_urls (pet_id, url) VALUES ($1, $2)`,
		petID, url,
	)
	if err != nil {
		return models.ApiResponse{}, err
	}

	return models.ApiResponse{Code: 200, Message: "File uploaded"}, nil
}
