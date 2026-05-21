// Package domain contains the core business entities of the ATS system.
// These types are framework-agnostic and must not import any adapter or infrastructure packages.
package domain

import "github.com/google/uuid"

// CV represents a candidate's curriculum vitae as stored in the system.
// RawText holds the original plain-text content extracted from the uploaded PDF.
// Skills is populated after AI parsing and persisted via the skills/cv_skills join tables.
type CV struct {
	UUID           uuid.UUID
	FirstName      string
	LastName       string
	CVTitle        string
	Specialization string
	Skills         []Skill
	WorkExperience int
	RawText        string
}
