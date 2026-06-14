package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Sample represents a sample record in the database
type Sample struct {
	bun.BaseModel `bun:"table:samples,alias:s"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	Name        string    `bun:"name,notnull" json:"name"`
	Description string    `bun:"description" json:"description"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeInsert hook sets timestamps before inserting
func (s *Sample) BeforeInsert(ctx context.Context) error {
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate hook updates the timestamp before updating
func (s *Sample) BeforeUpdate(ctx context.Context) error {
	s.UpdatedAt = time.Now()
	return nil
}
