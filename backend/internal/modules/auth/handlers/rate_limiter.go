package handlers

import (
	"context"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/cache"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// RateLimiter defines the behaviour required for throttling authentication endpoints.
type RateLimiter interface {
	// Allow increments the counter for the provided key and returns whether another request
	// is permitted within the configured limit/window. When the limit is exceeded, the
	// remaining duration until the counter resets is returned in retryAfter.
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (allow bool, retryAfter time.Duration, err error)
}

// noopRateLimiter is used when no external cache is available. It always allows the request.
type noopRateLimiter struct{}

func (noopRateLimiter) Allow(context.Context, string, int64, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}

// RedisRateLimiter implements RateLimiter using Redis for request counters.
type RedisRateLimiter struct {
	client *cache.Client
	prefix string
	log    *logger.Logger
}

// NewRedisRateLimiter constructs a Redis-backed rate limiter. If client is nil, a noop limiter is returned.
func NewRedisRateLimiter(client *cache.Client, prefix string, log *logger.Logger) RateLimiter {
	if client == nil {
		return noopRateLimiter{}
	}
	if log == nil {
		log = logger.Get()
	}
	return &RedisRateLimiter{
		client: client,
		prefix: prefix,
		log:    log,
	}
}

// Allow implements the RateLimiter interface using an increment-and-expire strategy.
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, time.Duration, error) {
	namespacedKey := r.prefix + key

	count, err := r.client.Increment(ctx, namespacedKey)
	if err != nil {
		r.log.WithContext(ctx).Warnf("rate limiter failed for key %s: %v", namespacedKey, err)
		return true, 0, err
	}

	if count == 1 {
		if expireErr := r.client.Expire(ctx, namespacedKey, window); expireErr != nil {
			r.log.WithContext(ctx).Warnf("failed to set TTL for key %s: %v", namespacedKey, expireErr)
		}
	}

	if count <= limit {
		return true, 0, nil
	}

	ttl, err := r.client.TTL(ctx, namespacedKey)
	if err != nil {
		r.log.WithContext(ctx).Warnf("failed to get TTL for key %s: %v", namespacedKey, err)
		return false, window, nil
	}

	if ttl < 0 {
		ttl = window
	}

	return false, ttl, nil
}
