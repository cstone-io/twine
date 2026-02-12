package database

import (
	"reflect"

	"gorm.io/gorm"

	"github.com/cstone-io/twine/pkg/errors"
)

// Seeder provides a framework for seeding test data
type Seeder struct {
	db        *gorm.DB
	batchSize int
}

// NewSeeder creates a new Seeder instance
func NewSeeder(db *gorm.DB, batchSize int) *Seeder {
	if batchSize <= 0 {
		batchSize = 100
	}
	return &Seeder{
		db:        db,
		batchSize: batchSize,
	}
}

// Seed inserts a slice of records into the database in batches
func (s *Seeder) Seed(records any) error {
	value := reflect.ValueOf(records)
	if value.Kind() != reflect.Slice {
		return errors.ErrSeedObject.WithValue("records must be a slice")
	}

	if value.Len() == 0 {
		return nil
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < value.Len(); i += s.batchSize {
			end := i + s.batchSize
			if end > value.Len() {
				end = value.Len()
			}

			batch := value.Slice(i, end).Interface()
			if err := tx.CreateInBatches(batch, s.batchSize).Error; err != nil {
				return errors.ErrSeedObject.Wrap(err)
			}
		}
		return nil
	})

	return err
}

// SeedOne inserts a single record into the database
func (s *Seeder) SeedOne(record any) error {
	if err := s.db.Create(record).Error; err != nil {
		return errors.ErrSeedObject.Wrap(err)
	}
	return nil
}

// Clear truncates the table for the given model
func (s *Seeder) Clear(model any) error {
	if err := s.db.Unscoped().Delete(model).Error; err != nil {
		return errors.ErrSeedObject.Wrap(err)
	}
	return nil
}
