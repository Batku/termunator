package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"termunator/internal/models"
	"termunator/internal/services"
	"termunator/internal/storage"
)

type App struct {
	ctx         context.Context
	db          *storage.Database
	sshService  *services.SSHService
	sftpService *services.SFTPService
	encryption  *storage.EncryptionService
}

func NewApp() *App {
	return &App{
		sshService:  services.NewSSHService(),
		sftpService: services.NewSFTPService(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.sshService.SetContext(ctx)

	// Initialize db
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".termunator")
	dbPath := filepath.Join(dataDir, "data.db")
	saltPath := filepath.Join(dataDir, "salt")

	// Ensure data directory exists
	os.MkdirAll(dataDir, 0755)

	// TODO
	// For now, use a default password - in the future this will be prompted from user
	masterPassword := "default_key"

	// Load or generate salt
	var salt []byte
	if saltData, err := os.ReadFile(saltPath); err == nil {
		salt = saltData
	} else {
		salt = storage.GenerateSalt()
		os.WriteFile(saltPath, salt, 0600)
	}

	a.encryption = storage.NewEncryptionService(masterPassword, salt)

	var err error
	a.db, err = storage.NewDatabase(dbPath, a.encryption)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
	}
}

// Host Management Methods

func (a *App) CreateHost(req models.HostCreateRequest) (*models.Host, error) {
	return a.db.CreateHost(req)
}

func (a *App) GetHosts() ([]*models.Host, error) {
	return a.db.GetHosts()
}

func (a *App) DeleteHost(id string) error {
	return a.db.DeleteHost(id)
}

func (a *App) UpdateHost(id string, req models.HostCreateRequest) (*models.Host, error) {
	return a.db.UpdateHost(id, req)
}

// SSH Session Methods

// establishes an SSH connection to a host with specified terminal dimensions
func (a *App) ConnectSSHWithDimensions(hostID string, cols, rows int) (string, error) {
	//log.Printf("APP API - ConnectSSHWithDimensions called: hostID=%s, cols=%d, rows=%d", hostID, cols, rows)

	host, err := a.db.GetHost(hostID)
	if err != nil {
		log.Printf("APP API - Failed to get host %s: %v", hostID, err)
		return "", fmt.Errorf("failed to get host: %w", err)
	}
	if host == nil {
		log.Printf("APP API - Host %s not found", hostID)
		return "", fmt.Errorf("host not found")
	}

	log.Printf("APP API - Connecting to host: %s@%s:%d", host.Username, host.Hostname, host.Port)
	session, err := a.sshService.ConnectWithDimensions(host, cols, rows)
	if err != nil {
		log.Printf("APP API - SSH connection failed for host %s: %v", hostID, err)
		return "", err
	}

	log.Printf("APP API - SSH connection successful, session ID: %s", session.ID)

	a.db.UpdateHostLastUsed(hostID)

	return session.ID, nil
}

func (a *App) GetSessionOutput(sessionID string) (string, error) {
	//log.Printf("APP API - GetSessionOutput called for session: %s", sessionID)
	output, err := a.sshService.ReadOutput(sessionID)
	if err != nil {
		log.Printf("APP API - GetSessionOutput ERROR for session %s: %v", sessionID, err)
	} else {
		//log.Printf("APP API - GetSessionOutput SUCCESS for session %s: length=%d, content=%q", sessionID, len(output), output)
	}
	return output, err
}

func (a *App) SendInput(sessionID, input string) error {
	return a.sshService.SendInput(sessionID, input)
}

func (a *App) ResizeTerminal(sessionID string, width, height int) error {
	return a.sshService.ResizeTerminal(sessionID, width, height)
}

// CheckSessionHealth checks if an SSH session is responsive
func (a *App) CheckSessionHealth(sessionID string) error {
	return a.sshService.CheckSessionHealth(sessionID)
}

func (a *App) CloseSSHSession(sessionID string) error {
	return a.sshService.CloseSession(sessionID)
}

func (a *App) GetActiveSessions() map[string]*services.SSHSession {
	return a.sshService.GetActiveSessions()
}

func (a *App) GetSessionPing(sessionID string) (int64, error) {
	return a.sshService.GetSessionPing(sessionID)
}

// SFTP Methods

func (a *App) ConnectSFTP(hostID string) error {
	host, err := a.db.GetHost(hostID)
	if err != nil {
		return fmt.Errorf("failed to get host: %w", err)
	}
	if host == nil {
		return fmt.Errorf("host not found")
	}

	_, err = a.sftpService.Connect(host, a.sshService)
	return err
}

func (a *App) GetClientHome() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return homeDir, nil
}

func (a *App) ListDirectory(hostID, path string) ([]*models.SFTPFileInfo, error) {
	return a.sftpService.ListDirectory(hostID, path)
}

func (a *App) MakeDirectory(hostID, path string) error {
	return a.sftpService.CreateDirectory(hostID, path)
}

func (a *App) UploadFileFromBytes(hostID, remotePath, base64Data string) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}
	return a.sftpService.UploadFileFromBytes(hostID, remotePath, data)
}

func (a *App) ListLocalDirectory(path string) ([]*models.SFTPFileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	var result []*models.SFTPFileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		fileInfo := &models.SFTPFileInfo{
			Name:        entry.Name(),
			Path:        filepath.Join(path, entry.Name()),
			Size:        info.Size(),
			Mode:        info.Mode().String(),
			ModTime:     info.ModTime(),
			IsDir:       entry.IsDir(),
			Permissions: info.Mode().Perm().String(),
		}
		result = append(result, fileInfo)
	}

	return result, nil
}

func (a *App) DownloadFile(hostID, remotePath, localPath string) error {
	return a.sftpService.DownloadFile(hostID, remotePath, localPath)
}

func (a *App) UploadFile(hostID, localPath, remotePath string) error {
	return a.sftpService.UploadFile(hostID, localPath, remotePath)
}

func (a *App) ReadLocalFileAsBytes(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read local file %s: %w", path, err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

// Macro Methods

//TODO: Implement macros

func (a *App) CreateMacro(req models.MacroCreateRequest) (*models.Macro, error) {
	return a.db.CreateMacro(req)
}

func (a *App) GetMacros() ([]*models.Macro, error) {
	return a.db.GetMacros()
}

func (a *App) ExecuteMacro(macroID, hostID string) error {
	// This would execute each command in the macro sequentially
	return fmt.Errorf("macro execution not yet implemented")
}

// History Methods

//TODO: Implement command history logging

func (a *App) GetHistory(hostID string, limit int) ([]*models.HistoryEntry, error) {
	return a.db.GetHistory(hostID, limit)
}

// Host Key Management Methods

func (a *App) AcceptHostKey(hostname, publicKey string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Ensure .ssh directory exists
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// publicKey should already be in the authorized/known_hosts key form (e.g. "ssh-ed25519 AAAA..."),
	// write a proper known_hosts line: "hostname <keyType> <base64key>"
	entry := fmt.Sprintf("%s %s\n", hostname, strings.TrimSpace(publicKey))

	file, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open known_hosts file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write to known_hosts file: %w", err)
	}

	log.Printf("Added host key for %s to known_hosts", hostname)
	return nil
}

func (a *App) ConnectSSHWithHostKeyVerification(hostID string, cols, rows int) (map[string]interface{}, error) {
	host, err := a.db.GetHost(hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}
	if host == nil {
		return nil, fmt.Errorf("host not found")
	}

	session, err := a.sshService.ConnectWithDimensions(host, cols, rows)
	if err != nil {
		log.Printf("Connection error for host %s: %v (type: %T)", host.Hostname, err, err)

		if hostKeyErr, ok := err.(services.HostKeyVerificationNeeded); ok {
			log.Printf("Host key verification needed for %s", host.Hostname)
			// Return host key info for frontend to handle
			return map[string]interface{}{
				"needsHostKeyVerification": true,
				"hostKeyInfo": map[string]interface{}{
					"hostname":    hostKeyErr.HostKeyInfo.Hostname,
					"algorithm":   hostKeyErr.HostKeyInfo.Algorithm,
					"fingerprint": hostKeyErr.HostKeyInfo.Fingerprint,
					"publicKey":   hostKeyErr.HostKeyInfo.PublicKey,
					"isNewHost":   hostKeyErr.HostKeyInfo.IsNewHost,
				},
			}, nil
		}

		// Check if it's wrapped in another error
		var hostKeyErr services.HostKeyVerificationNeeded
		if errors.As(err, &hostKeyErr) {
			log.Printf("Host key verification needed (wrapped) for %s", host.Hostname)
			// Return host key info for frontend to handle
			return map[string]interface{}{
				"needsHostKeyVerification": true,
				"hostKeyInfo": map[string]interface{}{
					"hostname":    hostKeyErr.HostKeyInfo.Hostname,
					"algorithm":   hostKeyErr.HostKeyInfo.Algorithm,
					"fingerprint": hostKeyErr.HostKeyInfo.Fingerprint,
					"publicKey":   hostKeyErr.HostKeyInfo.PublicKey,
					"isNewHost":   hostKeyErr.HostKeyInfo.IsNewHost,
				},
			}, nil
		}

		log.Printf("Other connection error for %s: %v", host.Hostname, err)
		return nil, err
	}

	// Connection successful, update last used and return session ID
	a.db.UpdateHostLastUsed(hostID)
	return map[string]interface{}{
		"needsHostKeyVerification": false,
		"sessionId":                session.ID,
	}, nil
}

func (a *App) GetKnownHosts() ([]map[string]interface{}, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")

	// Check if file exists
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		return []map[string]interface{}{}, nil
	}

	file, err := os.Open(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open known_hosts file: %w", err)
	}
	defer file.Close()

	var hosts []map[string]interface{}

	// this is a placeholder
	hosts = append(hosts, map[string]interface{}{
		"hostname":    "example.com",
		"algorithm":   "ECDSA",
		"fingerprint": "SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8",
		"addedDate":   "2025-09-15",
	})

	return hosts, nil
}

func (a *App) RemoveKnownHost(hostname string) error {
	// TODO: Implement actual known_hosts file parsing and removal
	log.Printf("Would remove known host: %s", hostname)
	return nil
}

func (a *App) ClearKnownHosts() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")

	// Create an empty file (effectively clearing it)
	file, err := os.Create(knownHostsPath)
	if err != nil {
		return fmt.Errorf("failed to clear known_hosts file: %w", err)
	}
	defer file.Close()

	log.Printf("Cleared all known hosts")
	return nil
}

// Private Key Management Methods

func (a *App) CreatePrivateKey(req models.PrivateKeyCreateRequest) (*models.PrivateKey, error) {
	return a.db.CreatePrivateKey(req)
}

func (a *App) GetPrivateKeys() ([]*models.PrivateKeyInfo, error) {
	return a.db.GetPrivateKeys()
}

func (a *App) GetPrivateKey(id string) (*models.PrivateKey, error) {
	return a.db.GetPrivateKey(id)
}

func (a *App) DeletePrivateKey(id string) error {
	return a.db.DeletePrivateKey(id)
}

func (a *App) Cleanup() {
	if a.db != nil {
		a.db.Close()
	}
}
