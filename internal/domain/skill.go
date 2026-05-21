package domain

import "github.com/google/uuid"

// Skill represents a professional skill extracted from a CV.
// Skills are stored in a normalised table and linked to CVs via a many-to-many join.
// UUID is assigned by the postgres adapter after upsert; Name must be unique across the skills table.
type Skill struct {
	UUID uuid.UUID
	Name string
}
