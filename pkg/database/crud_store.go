package database

import (
	"gorm.io/gorm"

	"github.com/cstone-io/twine/pkg/errors"
)

// CRUDStoreInterface defines the interface for CRUD operations
type CRUDStoreInterface[T any] interface {
	List(preloads ...string) ([]T, error)
	Get(id string, preloads ...string) (*T, error)
	Create(item T) error
	Update(item T) error
	Delete(id string) error
}

// CRUDStore provides generic CRUD operations for any model type
type CRUDStore[T any] struct {
	client *gorm.DB
}

// NewCRUDStore creates a new CRUD store for type T
func NewCRUDStore[T any](client *gorm.DB) *CRUDStore[T] {
	return &CRUDStore[T]{client: client}
}

// List retrieves all records with optional preloads
func (s *CRUDStore[T]) List(preloads ...string) ([]T, error) {
	query := s.client
	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	var items []T
	result := query.Find(&items)
	if result.Error != nil {
		return items, errors.ErrDatabaseRead.Wrap(result.Error).WithValue(items)
	}

	return items, nil
}

// Get retrieves a single record by ID with optional preloads
func (s *CRUDStore[T]) Get(id string, preloads ...string) (*T, error) {
	query := s.client
	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	var item T
	result := query.First(&item, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.ErrDatabaseObjectNotFound.Wrap(result.Error)
		}
		return nil, errors.ErrDatabaseRead.Wrap(result.Error).WithValue(item)
	}
	return &item, nil
}

// Create inserts a new record
func (s *CRUDStore[T]) Create(item T) error {
	result := s.client.Create(&item)
	if result.Error != nil {
		return errors.ErrDatabaseWrite.Wrap(result.Error)
	}
	return nil
}

// Update saves changes to an existing record
func (s *CRUDStore[T]) Update(item T) error {
	result := s.client.Save(&item)
	if result.Error != nil {
		return errors.ErrDatabaseUpdate.Wrap(result.Error)
	}
	return nil
}

// Delete soft-deletes a record by ID
func (s *CRUDStore[T]) Delete(id string) error {
	result := s.client.Delete(new(T), "id = ?", id)
	if result.Error != nil {
		return errors.ErrDatabaseDelete.Wrap(result.Error)
	}
	return nil
}
