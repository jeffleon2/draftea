package posgrest

import (
	"context"

	"gorm.io/gorm"
)

// repository is a generic GORM-based repository implementation.
// It provides standard CRUD operations for any entity type T.
type repository[T interface{}] struct {
	db *gorm.DB
}

// New creates a new generic repository instance for type T.
// The repository uses the provided GORM database connection for all operations.
func New[T interface{}](db *gorm.DB) *repository[T] {
	return &repository[T]{
		db,
	}
}

// Create inserts a new entity into the database.
func (r *repository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(&entity).Error
}

// GetAll retrieves all entities of type T from the database.
func (r *repository[T]) GetAll(ctx context.Context) (*[]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return &entities, nil
}

// GetByID retrieves a single entity by its ID.
func (r *repository[T]) GetByID(ctx context.Context, id string) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetBy retrieves entities matching a specific field value.
// The key parameter is the field name, and value is the value to match.
func (r *repository[T]) GetBy(ctx context.Context, key string, value interface{}) (*[]T, error) {
	var entity []T
	if err := r.db.WithContext(ctx).Where(key, value).Find(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Update updates an existing entity identified by ID.
func (r *repository[T]) Update(ctx context.Context, entity *T, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Updates(entity).Error
}

// Delete removes an entity by its ID.
func (r *repository[T]) Delete(ctx context.Context, id string) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}
