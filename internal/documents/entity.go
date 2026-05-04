package documents

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID        uuid.UUID
	Name      string
	Type      string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
