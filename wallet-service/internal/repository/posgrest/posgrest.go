package posgrest

import (
	"context"

	"gorm.io/gorm"
)

type repository[T interface{}] struct {
	db *gorm.DB
}

func New[T interface{}](db *gorm.DB) *repository[T] {
	return &repository[T]{
		db,
	}
}

func (r *repository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(&entity).Error
}

func (r *repository[T]) GetAll(ctx context.Context) (*[]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return &entities, nil
}

func (r *repository[T]) GetByID(ctx context.Context, id string) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *repository[T]) GetBy(ctx context.Context, key string, value interface{}) (*[]T, error) {
	var entity []T
	if err := r.db.WithContext(ctx).Where(key, value).Find(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *repository[T]) Update(ctx context.Context, entity *T, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Updates(entity).Error
}

func (r *repository[T]) Delete(ctx context.Context, id string) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}
