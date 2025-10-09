package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: These tests require a running Redis instance
// For CI/CD, consider using testcontainers or miniredis

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "6379", config.Port)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, 0, config.DB)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, 2, config.MinIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxIdleTime)
	assert.Equal(t, 5*time.Second, config.DialTimeout)
	assert.Equal(t, 3*time.Second, config.ReadTimeout)
	assert.Equal(t, 3*time.Second, config.WriteTimeout)
}

func TestNewClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	// Test health check
	err = client.Health(context.Background())
	assert.NoError(t, err)
}

func TestSetAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key := "test:key"
	value := "test_value"

	// Clean up before test
	_ = client.Delete(ctx, key)

	// Set value
	err = client.Set(ctx, key, value, 1*time.Minute)
	require.NoError(t, err)

	// Get value
	retrieved, err := client.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, retrieved)

	// Clean up after test
	err = client.Delete(ctx, key)
	require.NoError(t, err)
}

func TestGetNonExistentKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key := "test:nonexistent"

	_, err = client.Get(ctx, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSetAndGetJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key := "test:json"

	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	original := TestStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	// Clean up before test
	_ = client.Delete(ctx, key)

	// Set JSON
	err = client.SetJSON(ctx, key, original, 1*time.Minute)
	require.NoError(t, err)

	// Get JSON
	var retrieved TestStruct
	err = client.GetJSON(ctx, key, &retrieved)
	require.NoError(t, err)
	assert.Equal(t, original, retrieved)

	// Clean up after test
	err = client.Delete(ctx, key)
	require.NoError(t, err)
}

func TestDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key1 := "test:delete1"
	key2 := "test:delete2"

	// Set values
	_ = client.Set(ctx, key1, "value1", 1*time.Minute)
	_ = client.Set(ctx, key2, "value2", 1*time.Minute)

	// Delete multiple keys
	err = client.Delete(ctx, key1, key2)
	require.NoError(t, err)

	// Verify deletion
	exists1, _ := client.Exists(ctx, key1)
	exists2, _ := client.Exists(ctx, key2)
	assert.False(t, exists1)
	assert.False(t, exists2)
}

func TestExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key := "test:exists"

	// Clean up before test
	_ = client.Delete(ctx, key)

	// Key should not exist
	exists, err := client.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Set value
	_ = client.Set(ctx, key, "value", 1*time.Minute)

	// Key should exist
	exists, err = client.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Clean up after test
	_ = client.Delete(ctx, key)
}

func TestExpireAndTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key := "test:expire"

	// Clean up before test
	_ = client.Delete(ctx, key)

	// Set value with long expiration
	_ = client.Set(ctx, key, "value", 10*time.Minute)

	// Change expiration
	err = client.Expire(ctx, key, 1*time.Minute)
	require.NoError(t, err)

	// Check TTL
	ttl, err := client.TTL(ctx, key)
	require.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= 1*time.Minute)

	// Clean up after test
	_ = client.Delete(ctx, key)
}

func TestIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key := "test:counter"

	// Clean up before test
	_ = client.Delete(ctx, key)

	// Increment from 0
	val, err := client.Increment(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// Increment again
	val, err = client.Increment(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(2), val)

	// Clean up after test
	_ = client.Delete(ctx, key)
}

func TestIncrementBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key := "test:counter_by"

	// Clean up before test
	_ = client.Delete(ctx, key)

	// Increment by 5
	val, err := client.IncrementBy(ctx, key, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), val)

	// Increment by 10
	val, err = client.IncrementBy(ctx, key, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(15), val)

	// Clean up after test
	_ = client.Delete(ctx, key)
}

func TestSetNX(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key := "test:setnx"

	// Clean up before test
	_ = client.Delete(ctx, key)

	// First SetNX should succeed
	success, err := client.SetNX(ctx, key, "value1", 1*time.Minute)
	require.NoError(t, err)
	assert.True(t, success)

	// Second SetNX should fail (key exists)
	success, err = client.SetNX(ctx, key, "value2", 1*time.Minute)
	require.NoError(t, err)
	assert.False(t, success)

	// Value should be the first one
	val, err := client.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Clean up after test
	_ = client.Delete(ctx, key)
}

func TestSessionStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(nil)
	require.NoError(t, err)
	defer client.Close()

	sessionStorage := NewSessionStorage(client, "test:session:", 1*time.Hour)

	ctx := context.Background()
	sessionID := "test_session_123"

	type SessionData struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	originalData := SessionData{
		UserID:   "user123",
		Username: "testuser",
		Role:     "admin",
	}

	// Clean up before test
	_ = sessionStorage.DeleteSession(ctx, sessionID)

	// Set session
	err = sessionStorage.SetSession(ctx, sessionID, originalData)
	require.NoError(t, err)

	// Get session
	var retrievedData SessionData
	err = sessionStorage.GetSession(ctx, sessionID, &retrievedData)
	require.NoError(t, err)
	assert.Equal(t, originalData, retrievedData)

	// Check session exists
	exists, err := sessionStorage.SessionExists(ctx, sessionID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Refresh session (extend TTL)
	err = sessionStorage.RefreshSession(ctx, sessionID)
	require.NoError(t, err)

	// Delete session
	err = sessionStorage.DeleteSession(ctx, sessionID)
	require.NoError(t, err)

	// Verify deletion
	exists, err = sessionStorage.SessionExists(ctx, sessionID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFlushDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use a separate DB for this test to avoid affecting other tests
	config := DefaultConfig()
	config.DB = 15 // Use a different DB number

	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	// Set some test data
	_ = client.Set(ctx, "test:flush1", "value1", 1*time.Minute)
	_ = client.Set(ctx, "test:flush2", "value2", 1*time.Minute)

	// Flush DB
	err = client.FlushDB(ctx)
	require.NoError(t, err)

	// Verify all keys are gone
	exists1, _ := client.Exists(ctx, "test:flush1")
	exists2, _ := client.Exists(ctx, "test:flush2")
	assert.False(t, exists1)
	assert.False(t, exists2)
}
