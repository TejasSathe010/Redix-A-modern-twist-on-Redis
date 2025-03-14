package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"crypto/subtle"

	"golang.org/x/crypto/argon2"
)

type Authenticator struct {
	users     map[string]*User
	sessions  map[string]*Session
	salt      []byte
	mutex     sync.RWMutex
	sessionMu sync.RWMutex
}

type User struct {
	Username string
	Password []byte // Hashed password
	Role     string
}

type Session struct {
	ID        string
	Username  string
	ExpiresAt time.Time
}

func NewAuthenticator() *Authenticator {
	return &Authenticator{
		users:    make(map[string]*User),
		sessions: make(map[string]*Session),
		salt:     make([]byte, 16),
	}
}

func (a *Authenticator) initSalt() error {
	// Generate a random salt if not already set
	if len(a.salt) == 0 || len(a.salt) != 16 {
		a.salt = make([]byte, 16)
		_, err := rand.Read(a.salt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Authenticator) CreateUser(username, password string, role string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Check if user already exists
	if _, exists := a.users[username]; exists {
		return fmt.Errorf("user already exists")
	}

	// Hash the password
	hashedPassword, err := a.hashPassword(password)
	if err != nil {
		return err
	}

	// Create user
	a.users[username] = &User{
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}

	return nil
}

func (a *Authenticator) Authenticate(username, password string) (*Session, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	user, exists := a.users[username]
	if !exists {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Verify password
	if !a.verifyPassword(password, string(user.Password)) {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Create session
	session, err := a.createSession(user.Username)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (a *Authenticator) hashPassword(password string) ([]byte, error) {
	if err := a.initSalt(); err != nil {
		return nil, err
	}

	// Argon2 parameters (adjust based on security requirements)
	timeCost := 1
	memoryCost := 64 * 1024
	threads := 4
	hashLen := 32

	// Hash the password
	hash := argon2.IDKey([]byte(password), a.salt, timeCost, memoryCost, threads, hashLen)

	// Encode salt and hash using base64 for storage
	var buf [128]byte
	binary.BigEndian.PutUint32(buf[0:4], uint32(timeCost))
	binary.BigEndian.PutUint32(buf[4:8], uint32(memoryCost))
	binary.BigEndian.PutUint32(buf[8:12], uint32(threads))
	copy(buf[12:28], a.salt)
	copy(buf[28:60], hash)

	// Convert string to []byte
	encoded := base64.RawStdEncoding.EncodeToString(buf[:60])
	return []byte(encoded), nil
}

func (a *Authenticator) verifyPassword(password, hash string) bool {
	// Decode the hash
	data, err := base64.RawStdEncoding.DecodeString(hash)
	if err != nil || len(data) != 60 {
		return false
	}

	// Extract parameters and salt
	var params [12]byte
	copy(params[:], data[0:12])
	timeCost := int(binary.BigEndian.Uint32(params[0:4]))
	memoryCost := int(binary.BigEndian.Uint32(params[4:8]))
	threads := int(binary.BigEndian.Uint32(params[8:12]))
	salt := data[12:28]

	// Hash the provided password with the same parameters and salt
	comparisonHash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, 32)

	// Compare the hashes
	return subtle.ConstantTimeCompare(data[28:60], comparisonHash) == 1
}

func (a *Authenticator) createSession(username string) (*Session, error) {
	// Generate a random session ID
	sessionID := make([]byte, 32)
	_, err := rand.Read(sessionID)
	if err != nil {
		return nil, err
	}

	a.sessionMu.Lock()
	defer a.sessionMu.Unlock()

	session := &Session{
		ID:        base64.RawStdEncoding.EncodeToString(sessionID),
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Sessions expire after 24 hours
	}

	a.sessions[session.ID] = session
	return session, nil
}

func (a *Authenticator) ValidateSession(sessionID string) (*User, error) {
	a.sessionMu.RLock()
	defer a.sessionMu.RUnlock()

	session, exists := a.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	if time.Now().After(session.ExpiresAt) {
		delete(a.sessions, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	user, exists := a.users[session.Username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (a *Authenticator) InvalidateSession(sessionID string) {
	a.sessionMu.Lock()
	defer a.sessionMu.Unlock()

	delete(a.sessions, sessionID)
}
