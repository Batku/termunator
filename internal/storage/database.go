package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/ssh"

	"termunator/internal/models"
)

type Database struct {
	db         *sql.DB
	encryption *EncryptionService
}

func NewDatabase(dbPath string, encryption *EncryptionService) (*Database, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{
		db:         db,
		encryption: encryption,
	}

	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return database, nil
}

func (d *Database) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS hosts (
			id TEXT PRIMARY KEY,
			label TEXT NOT NULL,
			hostname TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 22,
			username TEXT NOT NULL,
			auth_method TEXT NOT NULL,
			password TEXT,
			private_key TEXT,
			private_key_id TEXT,
			tags TEXT,
			last_used DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS private_keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			key_type TEXT NOT NULL,
			key_data TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS macros (
			id TEXT PRIMARY KEY,
			label TEXT NOT NULL,
			commands TEXT NOT NULL,
			host_ids TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS history (
			id TEXT PRIMARY KEY,
			host_id TEXT NOT NULL,
			session_id TEXT NOT NULL,
			command TEXT NOT NULL,
			output TEXT,
			exit_code INTEGER,
			timestamp DATETIME NOT NULL,
			FOREIGN KEY (host_id) REFERENCES hosts (id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_hosts_last_used ON hosts(last_used DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_history_host_id ON history(host_id)`,
		`CREATE INDEX IF NOT EXISTS idx_history_timestamp ON history(timestamp DESC)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration query: %w", err)
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// Host operations

func (d *Database) CreateHost(req models.HostCreateRequest) (*models.Host, error) {
	host := &models.Host{
		ID:         uuid.New().String(),
		Label:      req.Label,
		Hostname:   req.Hostname,
		Port:       req.Port,
		Username:   req.Username,
		AuthMethod: req.AuthMethod,
		Tags:       req.Tags,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Encrypt sensitive data
	if req.Password != "" {
		encrypted, err := d.encryption.Encrypt(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		host.Password = encrypted
	}

	if req.PrivateKey != "" {
		encrypted, err := d.encryption.Encrypt(req.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt private key: %w", err)
		}
		host.PrivateKey = encrypted
	}

	tagsJSON, _ := json.Marshal(host.Tags)

	query := `INSERT INTO hosts (id, label, hostname, port, username, auth_method, password, private_key, tags, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query, host.ID, host.Label, host.Hostname, host.Port, host.Username,
		host.AuthMethod, host.Password, host.PrivateKey, string(tagsJSON), host.CreatedAt, host.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Don't return encrypted data in response
	host.Password = ""
	host.PrivateKey = ""

	return host, nil
}

func (d *Database) GetHosts() ([]*models.Host, error) {
	query := `SELECT id, label, hostname, port, username, auth_method, tags, last_used, created_at, updated_at
			  FROM hosts ORDER BY last_used DESC, created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query hosts: %w", err)
	}
	defer rows.Close()

	var hosts []*models.Host
	for rows.Next() {
		host := &models.Host{}
		var tagsJSON sql.NullString
		var lastUsed sql.NullTime

		err := rows.Scan(&host.ID, &host.Label, &host.Hostname, &host.Port, &host.Username,
			&host.AuthMethod, &tagsJSON, &lastUsed, &host.CreatedAt, &host.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan host: %w", err)
		}

		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &host.Tags)
		}

		if lastUsed.Valid {
			host.LastUsed = &lastUsed.Time
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}

// GetHost retrieves a host by ID with decrypted credentials
func (d *Database) GetHost(id string) (*models.Host, error) {
	query := `SELECT id, label, hostname, port, username, auth_method, password, private_key, tags, last_used, created_at, updated_at
			  FROM hosts WHERE id = ?`

	host := &models.Host{}
	var tagsJSON sql.NullString
	var lastUsed sql.NullTime
	var password, privateKey sql.NullString

	err := d.db.QueryRow(query, id).Scan(&host.ID, &host.Label, &host.Hostname, &host.Port, &host.Username,
		&host.AuthMethod, &password, &privateKey, &tagsJSON, &lastUsed, &host.CreatedAt, &host.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	if tagsJSON.Valid {
		json.Unmarshal([]byte(tagsJSON.String), &host.Tags)
	}

	if lastUsed.Valid {
		host.LastUsed = &lastUsed.Time
	}

	// Decrypt sensitive data
	if password.Valid && password.String != "" {
		decrypted, err := d.encryption.Decrypt(password.String)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		host.Password = decrypted
	}

	if privateKey.Valid && privateKey.String != "" {
		decrypted, err := d.encryption.Decrypt(privateKey.String)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt private key: %w", err)
		}
		host.PrivateKey = decrypted
	}

	return host, nil
}

func (d *Database) UpdateHostLastUsed(id string) error {
	query := `UPDATE hosts SET last_used = ?, updated_at = ? WHERE id = ?`
	now := time.Now()
	_, err := d.db.Exec(query, now, now, id)
	return err
}

func (d *Database) DeleteHost(id string) error {
	query := `DELETE FROM hosts WHERE id = ?`
	_, err := d.db.Exec(query, id)
	return err
}

func (d *Database) UpdateHost(id string, req models.HostCreateRequest) (*models.Host, error) {
	existingHost, err := d.GetHost(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing host: %w", err)
	}
	if existingHost == nil {
		return nil, fmt.Errorf("host not found")
	}

	host := &models.Host{
		ID:         id,
		Label:      req.Label,
		Hostname:   req.Hostname,
		Port:       req.Port,
		Username:   req.Username,
		AuthMethod: req.AuthMethod,
		Tags:       req.Tags,
		CreatedAt:  existingHost.CreatedAt, // Keep original creation time
		UpdatedAt:  time.Now(),
		LastUsed:   existingHost.LastUsed, // Keep last used time
	}

	// Encrypt sensitive data
	if req.Password != "" {
		encrypted, err := d.encryption.Encrypt(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		host.Password = encrypted
	}

	if req.PrivateKey != "" {
		encrypted, err := d.encryption.Encrypt(req.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt private key: %w", err)
		}
		host.PrivateKey = encrypted
	}

	tagsJSON, _ := json.Marshal(host.Tags)

	query := `UPDATE hosts SET label = ?, hostname = ?, port = ?, username = ?, auth_method = ?, 
			  password = ?, private_key = ?, tags = ?, updated_at = ?
			  WHERE id = ?`

	_, err = d.db.Exec(query, host.Label, host.Hostname, host.Port, host.Username,
		host.AuthMethod, host.Password, host.PrivateKey, string(tagsJSON),
		host.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update host: %w", err)
	}

	host.Password = ""
	host.PrivateKey = ""

	return host, nil
}

// Macro operations

func (d *Database) CreateMacro(req models.MacroCreateRequest) (*models.Macro, error) {
	macro := &models.Macro{
		ID:        uuid.New().String(),
		Label:     req.Label,
		Commands:  req.Commands,
		HostIDs:   req.HostIDs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	commandsJSON, _ := json.Marshal(macro.Commands)
	hostIDsJSON, _ := json.Marshal(macro.HostIDs)

	query := `INSERT INTO macros (id, label, commands, host_ids, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query, macro.ID, macro.Label, string(commandsJSON), string(hostIDsJSON), macro.CreatedAt, macro.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create macro: %w", err)
	}

	return macro, nil
}

func (d *Database) GetMacros() ([]*models.Macro, error) {
	query := `SELECT id, label, commands, host_ids, created_at, updated_at FROM macros ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query macros: %w", err)
	}
	defer rows.Close()

	var macros []*models.Macro
	for rows.Next() {
		macro := &models.Macro{}
		var commandsJSON, hostIDsJSON string

		err := rows.Scan(&macro.ID, &macro.Label, &commandsJSON, &hostIDsJSON, &macro.CreatedAt, &macro.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan macro: %w", err)
		}

		json.Unmarshal([]byte(commandsJSON), &macro.Commands)
		json.Unmarshal([]byte(hostIDsJSON), &macro.HostIDs)

		macros = append(macros, macro)
	}

	return macros, nil
}

// History operations

func (d *Database) AddHistoryEntry(entry models.HistoryEntry) error {
	entry.ID = uuid.New().String()
	entry.Timestamp = time.Now()

	query := `INSERT INTO history (id, host_id, session_id, command, output, exit_code, timestamp)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query, entry.ID, entry.HostID, entry.SessionID, entry.Command, entry.Output, entry.ExitCode, entry.Timestamp)
	return err
}

func (d *Database) GetHistory(hostID string, limit int) ([]*models.HistoryEntry, error) {
	query := `SELECT id, host_id, session_id, command, output, exit_code, timestamp
			  FROM history WHERE host_id = ? ORDER BY timestamp DESC LIMIT ?`

	rows, err := d.db.Query(query, hostID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var history []*models.HistoryEntry
	for rows.Next() {
		entry := &models.HistoryEntry{}
		err := rows.Scan(&entry.ID, &entry.HostID, &entry.SessionID, &entry.Command, &entry.Output, &entry.ExitCode, &entry.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history entry: %w", err)
		}
		history = append(history, entry)
	}

	return history, nil
}

// Private Key Management Methods

func (d *Database) CreatePrivateKey(req models.PrivateKeyCreateRequest) (*models.PrivateKey, error) {
	// Parse the key to get fingerprint and type
	signer, err := ssh.ParsePrivateKey([]byte(req.KeyData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	fingerprint := ssh.FingerprintSHA256(signer.PublicKey())
	keyType := signer.PublicKey().Type()

	privateKey := &models.PrivateKey{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Fingerprint: fingerprint,
		KeyType:     keyType,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	encrypted, err := d.encryption.Encrypt(req.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}
	privateKey.KeyData = encrypted

	query := `INSERT INTO private_keys (id, name, fingerprint, key_type, key_data, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = d.db.Exec(query, privateKey.ID, privateKey.Name, privateKey.Fingerprint,
		privateKey.KeyType, privateKey.KeyData, privateKey.CreatedAt, privateKey.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	privateKey.KeyData = ""
	return privateKey, nil
}

func (d *Database) GetPrivateKeys() ([]*models.PrivateKeyInfo, error) {
	query := `SELECT id, name, fingerprint, key_type, created_at
			  FROM private_keys ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query private keys: %w", err)
	}
	defer rows.Close()

	var keys []*models.PrivateKeyInfo
	for rows.Next() {
		key := &models.PrivateKeyInfo{}
		err := rows.Scan(&key.ID, &key.Name, &key.Fingerprint, &key.KeyType, &key.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan private key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func (d *Database) GetPrivateKey(id string) (*models.PrivateKey, error) {
	query := `SELECT id, name, fingerprint, key_type, key_data, created_at, updated_at
			  FROM private_keys WHERE id = ?`

	key := &models.PrivateKey{}
	var keyData string

	err := d.db.QueryRow(query, id).Scan(&key.ID, &key.Name, &key.Fingerprint,
		&key.KeyType, &keyData, &key.CreatedAt, &key.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	decrypted, err := d.encryption.Decrypt(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}
	key.KeyData = decrypted

	return key, nil
}

func (d *Database) DeletePrivateKey(id string) error {
	query := `DELETE FROM private_keys WHERE id = ?`
	_, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete private key: %w", err)
	}
	return nil
}
