package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all models
// Models should embed this struct to get ID, timestamps, and soft delete support
type BaseModel struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// BeforeCreate hook generates a UUID if not set
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}

// Polymorphic provides fields for polymorphic relationships
// See: https://gorm.io/docs/polymorphism.html
//
// Usage example:
//
//	type ChildModel struct {
//	    model.Polymorphic `gorm:"embedded"`
//	    Name string
//	}
//
//	type ParentModel struct {
//	    Field ChildModel `gorm:"polymorphic:Owner;polymorphicValue:child_model;"`
//	}
type Polymorphic struct {
	OwnerID   uuid.UUID `gorm:"type:uuid;index"`
	OwnerType string    `gorm:"index"`
}
