package models

import "time"

// unimplemented

type HistoryEntry struct {
	ID        string    `json:"id" db:"id"`
	HostID    string    `json:"host_id" db:"host_id"`
	SessionID string    `json:"session_id" db:"session_id"`
	Command   string    `json:"command" db:"command"`
	Output    string    `json:"output" db:"output"`
	ExitCode  int       `json:"exit_code" db:"exit_code"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

type Session struct {
	ID           string     `json:"id"`
	HostID       string     `json:"host_id"`
	Host         *Host      `json:"host,omitempty"`
	IsActive     bool       `json:"is_active"`
	StartedAt    time.Time  `json:"started_at"`
	LastActivity *time.Time `json:"last_activity,omitempty"`
}

type SFTPFileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Mode        string    `json:"mode"`
	ModTime     time.Time `json:"mod_time"`
	IsDir       bool      `json:"is_dir"`
	Permissions string    `json:"permissions"`
}
