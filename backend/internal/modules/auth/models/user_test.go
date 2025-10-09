package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	t.Run("SetPassword and VerifyPassword work correctly", func(t *testing.T) {
		user := &User{}
		password := "SecurePassword123!"

		// Set password
		err := user.SetPassword(password)
		require.NoError(t, err, "SetPassword should not return an error")
		assert.NotEmpty(t, user.PasswordHash, "Password hash should be set")

		// Verify correct password
		err = user.VerifyPassword(password)
		assert.NoError(t, err, "VerifyPassword should succeed with correct password")

		// Verify incorrect password
		err = user.VerifyPassword("WrongPassword123!")
		assert.Error(t, err, "VerifyPassword should fail with incorrect password")
		assert.Equal(t, ErrInvalidPassword, err, "Should return ErrInvalidPassword")
	})

	t.Run("Hash format is correct", func(t *testing.T) {
		user := &User{}
		password := "TestPassword123!"

		err := user.SetPassword(password)
		require.NoError(t, err)

		// Hash should contain a $ separator
		assert.Contains(t, user.PasswordHash, "$", "Hash should contain $ separator")

		// Should have exactly 2 parts (salt$hash)
		parts := len(user.PasswordHash)
		assert.Greater(t, parts, 10, "Hash should be non-trivial length")
	})

	t.Run("Different passwords produce different hashes", func(t *testing.T) {
		user1 := &User{}
		user2 := &User{}
		password := "SamePassword123!"

		err := user1.SetPassword(password)
		require.NoError(t, err)

		err = user2.SetPassword(password)
		require.NoError(t, err)

		// Even with the same password, hashes should differ due to random salt
		assert.NotEqual(t, user1.PasswordHash, user2.PasswordHash, "Same password should produce different hashes due to random salt")
	})

	t.Run("Empty password is rejected", func(t *testing.T) {
		user := &User{}
		err := user.SetPassword("")
		assert.Error(t, err, "Empty password should be rejected")
	})

	t.Run("Weak password is rejected", func(t *testing.T) {
		user := &User{}
		err := user.SetPassword("weak")
		assert.Error(t, err, "Weak password should be rejected")
	})
}

func TestArgon2Functions(t *testing.T) {
	t.Run("hashWithArgon2 produces valid hash", func(t *testing.T) {
		password := "TestPassword123!"
		hash, err := hashWithArgon2(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Contains(t, hash, "$", "Hash should contain $ separator")
	})

	t.Run("verifyArgon2Hash validates correct password", func(t *testing.T) {
		password := "TestPassword123!"
		hash, err := hashWithArgon2(password)
		require.NoError(t, err)

		valid, err := verifyArgon2Hash(password, hash)
		require.NoError(t, err)
		assert.True(t, valid, "Correct password should validate")
	})

	t.Run("verifyArgon2Hash rejects incorrect password", func(t *testing.T) {
		password := "TestPassword123!"
		hash, err := hashWithArgon2(password)
		require.NoError(t, err)

		valid, err := verifyArgon2Hash("WrongPassword!", hash)
		require.NoError(t, err)
		assert.False(t, valid, "Incorrect password should not validate")
	})

	t.Run("verifyArgon2Hash handles invalid hash format", func(t *testing.T) {
		password := "TestPassword123!"

		// Test with no $ separator
		valid, err := verifyArgon2Hash(password, "invalidhashformat")
		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "invalid hash format")

		// Test with too many parts
		valid, err = verifyArgon2Hash(password, "part1$part2$part3")
		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "invalid hash format")

		// Test with invalid base64
		valid, err = verifyArgon2Hash(password, "!!!invalid!!!$base64")
		assert.Error(t, err)
		assert.False(t, valid)
	})
}
