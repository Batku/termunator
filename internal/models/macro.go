package models

import "time"

// these are unimplemented
type Macro struct {
	ID        string    `json:"id" db:"id"`
	Label     string    `json:"label" db:"label"`
	Commands  []string  `json:"commands" db:"commands"`
	HostIDs   []string  `json:"host_ids" db:"host_ids"` // Empty means applies to all hosts
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MacroCreateRequest struct {
	Label    string   `json:"label"`
	Commands []string `json:"commands"`
	HostIDs  []string `json:"host_ids"`
}

type MacroUpdateRequest struct {
	Label    *string  `json:"label,omitempty"`
	Commands []string `json:"commands,omitempty"`
	HostIDs  []string `json:"host_ids,omitempty"`
}
