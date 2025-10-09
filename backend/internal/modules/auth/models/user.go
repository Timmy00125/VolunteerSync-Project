package models

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

// AccountStatus represents the status of a user account
type AccountStatus string

const (
	// AccountStatusActive represents an active user account
	AccountStatusActive AccountStatus = "active"
	// AccountStatusInactive represents an inactive user account (e.g., after 12 months of no login)
	AccountStatusInactive AccountStatus = "inactive"
	// AccountStatusSuspended represents a suspended user account
	AccountStatusSuspended AccountStatus = "suspended"
)

// Argon2 parameters for password hashing
// These values provide a good balance between security and performance
const (
	argon2Time      = 1         // Number of iterations
	argon2Memory    = 64 * 1024 // Memory in KiB (64 MB)
	argon2Threads   = 4         // Number of threads
	argon2KeyLength = 32        // Length of the derived key
	saltLength      = 16        // Length of the random salt
)

var (
	// ErrInvalidPassword is returned when password verification fails
	ErrInvalidPassword = errors.New("invalid password")
	// ErrInvalidSecurityAnswer is returned when security answer verification fails
	ErrInvalidSecurityAnswer = errors.New("invalid security answer")
	// ErrPasswordTooWeak is returned when password doesn't meet complexity requirements
	ErrPasswordTooWeak = errors.New("password must be at least 8 characters and contain letters and numbers")
)

// User represents a platform user with authentication credentials
// This is the core entity for all user types (volunteers, organization admins, etc.)
type User struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email               string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash        string         `gorm:"type:varchar(255);not null" json:"-"` // Never expose in JSON
	FirstName           string         `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName            string         `gorm:"type:varchar(100);not null" json:"last_name"`
	Phone               *string        `gorm:"type:varchar(20)" json:"phone,omitempty"`
	AccountStatus       AccountStatus  `gorm:"type:varchar(20);not null;default:'active'" json:"account_status"`
	LastLoginAt         *time.Time     `gorm:"type:timestamp" json:"last_login_at,omitempty"`
	EmailVerified       bool           `gorm:"not null;default:false" json:"email_verified"`
	SecurityQuestion1   string         `gorm:"column:security_question_1;type:varchar(255);not null" json:"-"`    // Never expose in JSON
	SecurityAnswer1Hash string         `gorm:"column:security_answer_1_hash;type:varchar(255);not null" json:"-"` // Never expose in JSON
	SecurityQuestion2   string         `gorm:"column:security_question_2;type:varchar(255);not null" json:"-"`    // Never expose in JSON
	SecurityAnswer2Hash string         `gorm:"column:security_answer_2_hash;type:varchar(255);not null" json:"-"` // Never expose in JSON
	SecurityQuestion3   string         `gorm:"column:security_question_3;type:varchar(255);not null" json:"-"`    // Never expose in JSON
	SecurityAnswer3Hash string         `gorm:"column:security_answer_3_hash;type:varchar(255);not null" json:"-"` // Never expose in JSON
	CreatedAt           time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to generate UUID if not provided
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// hashWithArgon2 performs Argon2id hashing with a random salt
// Returns the hash in the format: base64(salt)$base64(hash)
func hashWithArgon2(plaintext string) (string, error) {
	// Generate random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate hash using Argon2id
	hash := argon2.IDKey(
		[]byte(plaintext),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLength,
	)

	// Encode salt and hash as base64 and combine
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("%s$%s", encodedSalt, encodedHash), nil
}

// verifyArgon2Hash verifies a plaintext string against an Argon2 hash
func verifyArgon2Hash(plaintext, hashString string) (bool, error) {
	// Split the hash string into salt and hash
	var encodedSalt, encodedHash string
	if _, err := fmt.Sscanf(hashString, "%s$%s", &encodedSalt, &encodedHash); err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}

	// Decode salt and hash from base64
	salt, err := base64.RawStdEncoding.DecodeString(encodedSalt)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(encodedHash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Generate hash from plaintext using the same salt
	computedHash := argon2.IDKey(
		[]byte(plaintext),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLength,
	)

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(expectedHash, computedHash) == 1, nil
}

// SetPassword hashes and sets the user's password using Argon2id
func (u *User) SetPassword(password string) error {
	// Validate password strength
	if err := ValidatePasswordStrength(password); err != nil {
		return err
	}

	hash, err := hashWithArgon2(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.PasswordHash = hash
	return nil
}

// VerifyPassword checks if the provided password matches the stored hash
func (u *User) VerifyPassword(password string) error {
	valid, err := verifyArgon2Hash(password, u.PasswordHash)
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}

	if !valid {
		return ErrInvalidPassword
	}

	return nil
}

// SetSecurityAnswers hashes and sets all three security question answers
func (u *User) SetSecurityAnswers(q1, a1, q2, a2, q3, a3 string) error {
	// Hash all three security answers
	hash1, err := hashWithArgon2(a1)
	if err != nil {
		return fmt.Errorf("failed to hash security answer 1: %w", err)
	}

	hash2, err := hashWithArgon2(a2)
	if err != nil {
		return fmt.Errorf("failed to hash security answer 2: %w", err)
	}

	hash3, err := hashWithArgon2(a3)
	if err != nil {
		return fmt.Errorf("failed to hash security answer 3: %w", err)
	}

	// Set all questions and hashed answers
	u.SecurityQuestion1 = q1
	u.SecurityAnswer1Hash = hash1
	u.SecurityQuestion2 = q2
	u.SecurityAnswer2Hash = hash2
	u.SecurityQuestion3 = q3
	u.SecurityAnswer3Hash = hash3

	return nil
}

// VerifySecurityAnswer verifies a single security answer by question number (1, 2, or 3)
func (u *User) VerifySecurityAnswer(questionNumber int, answer string) error {
	var hashToVerify string

	switch questionNumber {
	case 1:
		hashToVerify = u.SecurityAnswer1Hash
	case 2:
		hashToVerify = u.SecurityAnswer2Hash
	case 3:
		hashToVerify = u.SecurityAnswer3Hash
	default:
		return fmt.Errorf("invalid question number: must be 1, 2, or 3")
	}

	valid, err := verifyArgon2Hash(answer, hashToVerify)
	if err != nil {
		return fmt.Errorf("security answer verification failed: %w", err)
	}

	if !valid {
		return ErrInvalidSecurityAnswer
	}

	return nil
}

// VerifySecurityAnswers verifies multiple security answers
// Returns the number of correct answers and any error
// Per FR-003a, minimum 2 of 3 correct answers required for password reset
func (u *User) VerifySecurityAnswers(answers map[int]string) (int, error) {
	correctCount := 0

	for questionNumber, answer := range answers {
		err := u.VerifySecurityAnswer(questionNumber, answer)
		if err == nil {
			correctCount++
		} else if !errors.Is(err, ErrInvalidSecurityAnswer) {
			// Return only if it's an unexpected error, not just wrong answer
			return correctCount, err
		}
	}

	return correctCount, nil
}

// GetSecurityQuestions returns the three security questions (not the answers)
// Used during password reset flow
func (u *User) GetSecurityQuestions() []string {
	return []string{
		u.SecurityQuestion1,
		u.SecurityQuestion2,
		u.SecurityQuestion3,
	}
}

// ValidatePasswordStrength checks if a password meets complexity requirements
// FR-002a: Password must be at least 8 characters and contain letters and numbers
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooWeak
	}

	hasLetter := false
	hasNumber := false

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z':
			hasLetter = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}

		if hasLetter && hasNumber {
			break
		}
	}

	if !hasLetter || !hasNumber {
		return ErrPasswordTooWeak
	}

	return nil
}

// IsActive returns true if the account is in active status
func (u *User) IsActive() bool {
	return u.AccountStatus == AccountStatusActive
}

// MarkInactive marks the user account as inactive
// Per FR-008, accounts are marked inactive after 12 months of no login
func (u *User) MarkInactive() {
	u.AccountStatus = AccountStatusInactive
}

// MarkSuspended marks the user account as suspended
func (u *User) MarkSuspended() {
	u.AccountStatus = AccountStatusSuspended
}

// UpdateLastLogin updates the last login timestamp to current time
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// GetFullName returns the user's full name (first + last)
func (u *User) GetFullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}
