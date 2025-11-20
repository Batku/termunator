package models

import "time"

type PrivateKey struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Fingerprint string    `json:"fingerprint" db:"fingerprint"`
	KeyType     string    `json:"key_type" db:"key_type"`           // RSA, ECDSA, ED25519, etc.
	KeyData     string    `json:"key_data,omitempty" db:"key_data"` // Encrypted
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type PrivateKeyCreateRequest struct {
	Name    string `json:"name"`
	KeyData string `json:"key_data"`
}

type PrivateKeyInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Fingerprint string    `json:"fingerprint"`
	KeyType     string    `json:"key_type"`
	CreatedAt   time.Time `json:"created_at"`
}
