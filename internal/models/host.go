package models

import "time"

type AuthMethod string

const (
	AuthPassword   AuthMethod = "password"
	AuthPrivateKey AuthMethod = "private_key"
	AuthAgent      AuthMethod = "ssh_agent"
)

type Host struct {
	ID         string     `json:"id" db:"id"`
	Label      string     `json:"label" db:"label"`
	Hostname   string     `json:"hostname" db:"hostname"`
	Port       int        `json:"port" db:"port"`
	Username   string     `json:"username" db:"username"`
	AuthMethod AuthMethod `json:"auth_method" db:"auth_method"`
	Password   string     `json:"password,omitempty" db:"password"`       // Encrypted
	PrivateKey string     `json:"private_key,omitempty" db:"private_key"` // Encrypted
	Tags       []string   `json:"tags" db:"tags"`
	LastUsed   *time.Time `json:"last_used" db:"last_used"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

type HostCreateRequest struct {
	Label      string     `json:"label"`
	Hostname   string     `json:"hostname"`
	Port       int        `json:"port"`
	Username   string     `json:"username"`
	AuthMethod AuthMethod `json:"auth_method"`
	Password   string     `json:"password,omitempty"`
	PrivateKey string     `json:"private_key,omitempty"`
	Tags       []string   `json:"tags"`
}

type HostUpdateRequest struct {
	Label      *string     `json:"label,omitempty"`
	Hostname   *string     `json:"hostname,omitempty"`
	Port       *int        `json:"port,omitempty"`
	Username   *string     `json:"username,omitempty"`
	AuthMethod *AuthMethod `json:"auth_method,omitempty"`
	Password   *string     `json:"password,omitempty"`
	PrivateKey *string     `json:"private_key,omitempty"`
	Tags       []string    `json:"tags,omitempty"`
}
