package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Setting represents a key-value configuration setting in the database
type Setting struct {
	bun.BaseModel `bun:"table:settings,alias:st"`

	Key       string    `bun:"key,pk" json:"key"`
	Value     string    `bun:"value,notnull" json:"value"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeInsert hook sets timestamp before inserting
func (s *Setting) BeforeInsert(ctx context.Context) error {
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook updates the timestamp before updating
func (s *Setting) BeforeUpdate(ctx context.Context) error {
	s.UpdatedAt = time.Now()
	return nil
}
