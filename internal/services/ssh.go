package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"

	"termunator/internal/models"
)

type HostKeyInfo struct {
	Hostname    string `json:"hostname"`
	Algorithm   string `json:"algorithm"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"publicKey"`
	IsNewHost   bool   `json:"isNewHost"`
}

type HostKeyVerificationNeeded struct {
	HostKeyInfo HostKeyInfo
}

func (e HostKeyVerificationNeeded) Error() string {
	return fmt.Sprintf("host key verification needed for %s", e.HostKeyInfo.Hostname)
}

type SSHService struct {
	sessions map[string]*SSHSession
	mutex    sync.RWMutex
	ctx      context.Context
}

type SSHSession struct {
	ID             string
	Host           *models.Host
	Client         *ssh.Client
	Session        *ssh.Session
	IsActive       bool
	stdin          io.WriteCloser
	stdout         io.Reader
	stderr         io.Reader
	ctx            context.Context
	Ping           int64     // last measured ping in ms
	lastPingSentAt time.Time // last time a ping or command was sent
}

func NewSSHService() *SSHService {
	return &SSHService{
		sessions: make(map[string]*SSHSession),
	}
}

func (s *SSHService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// SSH connection to a host with default terminal size
func (s *SSHService) Connect(host *models.Host) (*SSHSession, error) {
	return s.ConnectWithDimensions(host, 80, 24)
}

// SSH connection to a host with specified terminal dimensions
func (s *SSHService) ConnectWithDimensions(host *models.Host, cols, rows int) (*SSHSession, error) {
	log.Printf("SSH SERVICE - ConnectWithDimensions called: host=%s@%s:%d, dims=%dx%d", host.Username, host.Hostname, host.Port, cols, rows)

	// Build SSH client config
	config, err := s.BuildSSHConfig(host)
	if err != nil {
		log.Printf("SSH SERVICE - Failed to build SSH config: %v", err)
		return nil, fmt.Errorf("failed to build SSH config: %w", err)
	}
	log.Printf("SSH SERVICE - SSH config built successfully")

	// Connect to the host
	address := fmt.Sprintf("%s:%d", host.Hostname, host.Port)
	log.Printf("SSH SERVICE - Attempting to dial %s", address)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Printf("SSH SERVICE - Failed to dial %s: %v", address, err)
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	log.Printf("SSH SERVICE - Successfully connected to %s", address)

	// Create a new session
	log.Printf("SSH SERVICE - Creating new SSH session")
	session, err := client.NewSession()
	if err != nil {
		log.Printf("SSH SERVICE - Failed to create session: %v", err)
		client.Close()
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	log.Printf("SSH SERVICE - SSH session created successfully")

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// Request a pseudo terminal with specified dimensions
	log.Printf("SSH SERVICE - Requesting PTY with dimensions %dx%d", cols, rows)
	if err := session.RequestPty("xterm-256color", cols, rows, modes); err != nil {
		log.Printf("SSH SERVICE - Failed to request PTY: %v", err)
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to request pty: %w", err)
	}
	log.Printf("SSH SERVICE - PTY requested successfully")

	// Set up pipes
	log.Printf("SSH SERVICE - Setting up stdin pipe")
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Printf("SSH SERVICE - Failed to create stdin pipe: %v", err)
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	log.Printf("SSH SERVICE - Setting up stdout pipe")
	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Printf("SSH SERVICE - Failed to create stdout pipe: %v", err)
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	log.Printf("SSH SERVICE - Setting up stderr pipe")
	stderr, err := session.StderrPipe()
	if err != nil {
		log.Printf("SSH SERVICE - Failed to create stderr pipe: %v", err)
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start shell
	log.Printf("SSH SERVICE - Starting shell")
	if err := session.Shell(); err != nil {
		log.Printf("SSH SERVICE - Failed to start shell: %v", err)
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}
	log.Printf("SSH SERVICE - Shell started successfully")

	sshSession := &SSHSession{
		ID:       fmt.Sprintf("session_%d", time.Now().UnixNano()),
		Host:     host,
		Client:   client,
		Session:  session,
		IsActive: true,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		ctx:      s.ctx,
	}

	log.Printf("SSH SERVICE - Created session with ID: %s", sshSession.ID)

	s.mutex.Lock()
	s.sessions[sshSession.ID] = sshSession
	s.mutex.Unlock()

	go s.streamOutput(sshSession)

	log.Printf("SSH SERVICE - Connection successful, returning session %s", sshSession.ID)
	return sshSession, nil
}

func (s *SSHService) BuildSSHConfig(host *models.Host) (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            host.Username,
		Timeout:         30 * time.Second,
		HostKeyCallback: s.createHostKeyCallback(),
	}

	log.Printf("SSH SERVICE - Building SSH config for %s@%s with auth method: %s", host.Username, host.Hostname, host.AuthMethod)

	switch host.AuthMethod {
	case models.AuthPassword:
		if host.Password == "" {
			return nil, fmt.Errorf("password is required for password authentication")
		}
		log.Printf("SSH SERVICE - Using password authentication")
		config.Auth = []ssh.AuthMethod{
			ssh.Password(host.Password),
		}

	case models.AuthPrivateKey:
		if host.PrivateKey == "" {
			return nil, fmt.Errorf("private key is required for key authentication")
		}

		log.Printf("SSH SERVICE - Using private key authentication")
		// Try to parse the private key
		signer, err := ssh.ParsePrivateKey([]byte(host.PrivateKey))
		if err != nil {
			// Check if it's an encrypted key that needs a passphrase
			if _, ok := err.(*ssh.PassphraseMissingError); ok {
				return nil, fmt.Errorf("private key is encrypted and requires a passphrase (not yet supported)")
			}
			return nil, fmt.Errorf("failed to parse private key (key may be invalid or corrupted): %w", err)
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}

	case models.AuthAgent:
		log.Printf("SSH SERVICE - Using SSH agent authentication")
		authMethods, err := s.getSSHAgentAuth()
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH agent auth: %w", err)
		}
		config.Auth = authMethods

	default:
		return nil, fmt.Errorf("unsupported authentication method: %s", host.AuthMethod)
	}

	log.Printf("SSH SERVICE - SSH config built with %d auth methods", len(config.Auth))
	return config, nil
}

func (s *SSHService) createHostKeyCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Get fingerprint, algorithm and marshaled public key
		fingerprint := ssh.FingerprintSHA256(key)
		algorithm := key.Type()
		publicKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))

		log.Printf("SSH SERVICE - Host key received: %s %s for %s", algorithm, fingerprint, hostname)

		// Try to load known_hosts file to check if host is known
		homeDir, err := os.UserHomeDir()
		var isKnownHost bool = false

		if err == nil {
			knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")
			log.Printf("SSH SERVICE - Checking known_hosts file: %s", knownHostsPath)

			if callback, err := knownhosts.New(knownHostsPath); err == nil {
				// Check if the host key is already known
				err := callback(hostname, remote, key)
				if err == nil {
					// Host key is known and valid
					log.Printf("SSH SERVICE - Host key is known and valid for %s", hostname)
					return nil
				}

				// Check if it's a key mismatch or unknown host
				if keyErr, ok := err.(*knownhosts.KeyError); ok {
					log.Printf("SSH SERVICE - Host key mismatch for %s: %v", hostname, keyErr)
					isKnownHost = true
				} else {
					log.Printf("SSH SERVICE - Unknown host or other error for %s: %v", hostname, err)
				}
				// If not a KeyError, it's probably an unknown host
			} else {
				log.Printf("SSH SERVICE - Failed to load known_hosts: %v", err)
			}
		} else {
			log.Printf("SSH SERVICE - Failed to get home directory: %v", err)
		}

		hostKeyInfo := HostKeyInfo{
			Hostname:    hostname,
			Algorithm:   algorithm,
			Fingerprint: fingerprint,
			PublicKey:   publicKey,
			IsNewHost:   !isKnownHost,
		}

		log.Printf("SSH SERVICE - Host key verification needed for %s (isNewHost: %v)", hostname, !isKnownHost)
		return HostKeyVerificationNeeded{HostKeyInfo: hostKeyInfo}
	}
}

func (s *SSHService) getSSHAgentAuth() ([]ssh.AuthMethod, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK environment variable not set")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	agentClient := agent.NewClient(conn)
	return []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}, nil
}

func (s *SSHService) SendInput(sessionID, input string) error {
	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if !session.IsActive {
		return fmt.Errorf("session %s is not active", sessionID)
	}

	// Track ping send time for interactive input
	session.lastPingSentAt = time.Now()

	_, err := session.stdin.Write([]byte(input))
	return err
}

func (s *SSHService) ResizeTerminal(sessionID string, width, height int) error {
	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if !session.IsActive {
		return fmt.Errorf("session %s is not active", sessionID)
	}

	return session.Session.WindowChange(height, width)
}

func (s *SSHService) ReadOutput(sessionID string) (string, error) {

	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists {
		log.Printf("Session not found: %s", sessionID)
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if !session.IsActive {
		log.Printf("Session %s is not active", sessionID)
		return "", fmt.Errorf("session %s is not active", sessionID)
	}

	output := s.getAndClearSessionBuffer(sessionID)

	return output, nil
}

func (s *SSHService) CheckSessionHealth(sessionID string) error {
	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if !session.IsActive {
		return fmt.Errorf("session %s is not active", sessionID)
	}

	// Track ping send time for keepalive
	session.lastPingSentAt = time.Now()

	_, _, err := session.Client.SendRequest("ping", true, nil)
	if err != nil {
		return fmt.Errorf("ping failed for session %s: %w", sessionID, err)
	}
	return nil
}

func (s *SSHService) CloseSession(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.IsActive = false
	if session.Session != nil {
		session.Session.Close()
	}
	if session.Client != nil {
		session.Client.Close()
	}

	delete(s.sessions, sessionID)
	return nil
}

func (s *SSHService) GetActiveSessions() map[string]*SSHSession {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[string]*SSHSession)
	for id, session := range s.sessions {
		if session.IsActive {
			result[id] = session
		}
	}
	return result
}

func (s *SSHService) ExecuteCommand(host *models.Host, command string) (string, error) {
	config, err := s.BuildSSHConfig(host)
	if err != nil {
		return "", fmt.Errorf("failed to build SSH config: %w", err)
	}

	address := fmt.Sprintf("%s:%d", host.Hostname, host.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (s *SSHService) streamOutput(session *SSHSession) {
	log.Printf("STREAM - Starting output streaming for session %s", session.ID)

	buffer := make([]byte, 4096)

	for session.IsActive {
		// Set a read timeout to make this non-blocking
		if conn, ok := session.stdout.(interface{ SetReadDeadline(time.Time) error }); ok {
			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		}

		n, err := session.stdout.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if err == io.EOF {
				log.Printf("STREAM - EOF reached for session %s", session.ID)
				break
			}
			log.Printf("STREAM - Error reading from session %s: %v", session.ID, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if n > 0 {
			output := string(buffer[:n])

			// If a ping was sent, calculate round-trip time
			if !session.lastPingSentAt.IsZero() {
				session.Ping = time.Since(session.lastPingSentAt).Milliseconds()
				session.lastPingSentAt = time.Time{}
			}

			// Store in session buffer for immediate retrieval
			s.addToSessionBuffer(session.ID, output)
		}

		// Small delay to prevent excessive CPU usage
		time.Sleep(10 * time.Millisecond)
	}

	log.Printf("STREAM - Stopped streaming for session %s", session.ID)
}

func (s *SSHService) GetSessionPing(sessionID string) (int64, error) {
	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()
	if !exists {
		return 0, fmt.Errorf("session %s not found", sessionID)
	}
	return session.Ping, nil
}

var sessionBuffers = make(map[string]string)
var bufferMutex sync.RWMutex

func (s *SSHService) addToSessionBuffer(sessionID, output string) {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
	sessionBuffers[sessionID] += output
}

func (s *SSHService) getAndClearSessionBuffer(sessionID string) string {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
	output := sessionBuffers[sessionID]
	sessionBuffers[sessionID] = ""
	return output
}
