package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthCode struct {
	CodeHash    string
	UserID      int
	ClientID    string
	RedirectURI string
	State       string
	ExpiresAt   time.Time
}

type PwdResetToken struct {
	TokenHash string
	UserID    int
	ExpiresAt time.Time
}

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

type PasswordHasher struct {
	cost int
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{cost: 10}
}

func (h *PasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

func (h *PasswordHasher) Compare(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type TokenGenerator interface {
	Generate() (string, error)
}

type SecureTokenGenerator struct {
	length int
}

func NewSecureTokenGenerator(length int) *SecureTokenGenerator {
	return &SecureTokenGenerator{length: length}
}

func (g *SecureTokenGenerator) Generate() (string, error) {
	bytes := make([]byte, g.length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

type EmailValidator interface {
	Validate(email string) bool
}

type RegexEmailValidator struct {
	pattern *regexp.Regexp
}

func NewEmailValidator() *RegexEmailValidator {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return &RegexEmailValidator{pattern: pattern}
}

func (v *RegexEmailValidator) Validate(email string) bool {
	return v.pattern.MatchString(email)
}

type AuthCodeManager struct {
	hasher    Hasher
	generator TokenGenerator
	ttl       time.Duration
}

func NewAuthCodeManager(hasher Hasher, generator TokenGenerator, ttl time.Duration) *AuthCodeManager {
	return &AuthCodeManager{
		hasher:    hasher,
		generator: generator,
		ttl:       ttl,
	}
}

func (m *AuthCodeManager) Generate(code string) (string, error) {
	hash, err := m.hasher.Hash(code)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func (m *AuthCodeManager) Verify(codeHash, code string) bool {
	return m.hasher.Compare(codeHash, code)
}

func (m *AuthCodeManager) CreateAuthCode(userID int, clientID, redirectURI, state string) (*AuthCode, string, error) {
	code, err := m.generator.Generate()
	if err != nil {
		return nil, "", err
	}

	codeHash, err := m.Generate(code)
	if err != nil {
		return nil, "", err
	}

	authCode := &AuthCode{
		CodeHash:    codeHash,
		UserID:      userID,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		State:       state,
		ExpiresAt:   time.Now().Add(m.ttl),
	}

	return authCode, code, nil
}

func (m *AuthCodeManager) CreatePwdResetToken(userID int) (*PwdResetToken, string, error) {
	token, err := m.generator.Generate()
	if err != nil {
		return nil, "", err
	}

	tokenHash, err := m.Generate(token)
	if err != nil {
		return nil, "", err
	}

	resetToken := &PwdResetToken{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(m.ttl),
	}

	return resetToken, token, nil
}
