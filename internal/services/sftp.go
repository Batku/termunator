package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"termunator/internal/models"
)

type SFTPService struct {
	clients map[string]*SFTPClient
	mutex   sync.RWMutex
}

func (s *SFTPService) UploadFileFromBytes(hostID, remotePath string, data []byte) error {
	s.mutex.RLock()
	client, exists := s.clients[hostID]
	s.mutex.RUnlock()
	if !exists || !client.IsActive {
		return fmt.Errorf("SFTP client not found or inactive for host %s", hostID)
	}
	remoteFile, err := client.Client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()
	_, err = remoteFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to remote file: %w", err)
	}
	return nil
}

type SFTPClient struct {
	HostID    string
	SSHClient *ssh.Client
	Client    *sftp.Client
	IsActive  bool
}

func NewSFTPService() *SFTPService {
	return &SFTPService{
		clients: make(map[string]*SFTPClient),
	}
}

func (s *SFTPService) Connect(host *models.Host, sshService *SSHService) (*SFTPClient, error) {
	// Build SSH client config
	config, err := sshService.BuildSSHConfig(host)
	if err != nil {
		return nil, fmt.Errorf("failed to build SSH config: %w", err)
	}

	// Connect to the host
	address := fmt.Sprintf("%s:%d", host.Hostname, host.Port)
	sshClient, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	client := &SFTPClient{
		HostID:    host.ID,
		SSHClient: sshClient,
		Client:    sftpClient,
		IsActive:  true,
	}

	s.mutex.Lock()
	s.clients[host.ID] = client
	s.mutex.Unlock()

	return client, nil
}

func (s *SFTPService) ListDirectory(hostID, path string) ([]*models.SFTPFileInfo, error) {
	s.mutex.RLock()
	client, exists := s.clients[hostID]
	s.mutex.RUnlock()

	if !exists || !client.IsActive {
		return nil, fmt.Errorf("SFTP client not found or inactive for host %s", hostID)
	}

	files, err := client.Client.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory %s: %w", path, err)
	}

	var result []*models.SFTPFileInfo
	for _, file := range files {
		fileInfo := &models.SFTPFileInfo{
			Name:        file.Name(),
			Path:        filepath.Join(path, file.Name()),
			Size:        file.Size(),
			Mode:        file.Mode().String(),
			ModTime:     file.ModTime(),
			IsDir:       file.IsDir(),
			Permissions: file.Mode().Perm().String(),
		}
		result = append(result, fileInfo)
	}

	return result, nil
}

func (s *SFTPService) DownloadFile(hostID, remotePath, localPath string) error {
	s.mutex.RLock()
	client, exists := s.clients[hostID]
	s.mutex.RUnlock()

	if !exists || !client.IsActive {
		return fmt.Errorf("SFTP client not found or inactive for host %s", hostID)
	}

	// Open remote file
	remoteFile, err := client.Client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	// Copy data
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	return nil
}

func (s *SFTPService) UploadFile(hostID, localPath, remotePath string) error {
	s.mutex.RLock()
	client, exists := s.clients[hostID]
	s.mutex.RUnlock()

	if !exists || !client.IsActive {
		return fmt.Errorf("SFTP client not found or inactive for host %s", hostID)
	}

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	// Create remote file
	remoteFile, err := client.Client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Copy data
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	return nil
}

// Creates a directory on the remote server
func (s *SFTPService) CreateDirectory(hostID, path string) error {
	s.mutex.RLock()
	client, exists := s.clients[hostID]
	s.mutex.RUnlock()

	if !exists || !client.IsActive {
		return fmt.Errorf("SFTP client not found or inactive for host %s", hostID)
	}

	return client.Client.Mkdir(path)
}

// Deletes a file on the remote server
func (s *SFTPService) DeleteFile(hostID, path string) error {
	s.mutex.RLock()
	client, exists := s.clients[hostID]
	s.mutex.RUnlock()

	if !exists || !client.IsActive {
		return fmt.Errorf("SFTP client not found or inactive for host %s", hostID)
	}

	return client.Client.Remove(path)
}

// Deletes a directory on the remote server
func (s *SFTPService) DeleteDirectory(hostID, path string) error {
	s.mutex.RLock()
	client, exists := s.clients[hostID]
	s.mutex.RUnlock()

	if !exists || !client.IsActive {
		return fmt.Errorf("SFTP client not found or inactive for host %s", hostID)
	}

	return client.Client.RemoveDirectory(path)
}

func (s *SFTPService) GetWorkingDirectory(hostID string) (string, error) {
	s.mutex.RLock()
	client, exists := s.clients[hostID]
	s.mutex.RUnlock()

	if !exists || !client.IsActive {
		return "", fmt.Errorf("SFTP client not found or inactive for host %s", hostID)
	}

	return client.Client.Getwd()
}

// this only validates the path, the rest is done on the frontend
func (s *SFTPService) ChangeDirectory(hostID, path string) error {
	s.mutex.RLock()
	client, exists := s.clients[hostID]
	s.mutex.RUnlock()

	if !exists || !client.IsActive {
		return fmt.Errorf("SFTP client not found or inactive for host %s", hostID)
	}

	// Check if the path exists and is a directory
	stat, err := client.Client.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	return nil
}

func (s *SFTPService) CloseConnection(hostID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, exists := s.clients[hostID]
	if !exists {
		return fmt.Errorf("SFTP client not found for host %s", hostID)
	}

	client.IsActive = false
	if client.Client != nil {
		client.Client.Close()
	}
	if client.SSHClient != nil {
		client.SSHClient.Close()
	}

	delete(s.clients, hostID)
	return nil
}

func (s *SFTPService) GetActiveClients() map[string]*SFTPClient {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[string]*SFTPClient)
	for id, client := range s.clients {
		if client.IsActive {
			result[id] = client
		}
	}
	return result
}
