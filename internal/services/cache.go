package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
	"marketplace-app/internal/config"
)

type CacheService struct {
	redisClient *redis.Client
	memoryCache *cache.Cache
	config      *config.Config
	ctx         context.Context
}

func NewCacheService(cfg *config.Config) *CacheService {
	ctx := context.Background()

	// Initialize Redis client
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		// Fallback to default Redis configuration
		opt = &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		// If Redis is not available, we'll rely on memory cache only
		redisClient = nil
	}

	// Initialize in-memory cache as fallback
	memoryCache := cache.New(time.Duration(cfg.CacheExpiration)*time.Minute, 10*time.Minute)

	return &CacheService{
		redisClient: redisClient,
		memoryCache: memoryCache,
		config:      cfg,
		ctx:         ctx,
	}
}

// Set stores a value in cache with expiration
func (cs *CacheService) Set(key string, value interface{}, expiration time.Duration) error {
	// Try Redis first
	if cs.redisClient != nil {
		jsonData, err := json.Marshal(value)
		if err == nil {
			err = cs.redisClient.Set(cs.ctx, key, jsonData, expiration).Err()
			if err == nil {
				return nil
			}
		}
	}

	// Fallback to memory cache
	cs.memoryCache.Set(key, value, expiration)
	return nil
}

// Get retrieves a value from cache
func (cs *CacheService) Get(key string) interface{} {
	// Try Redis first
	if cs.redisClient != nil {
		val, err := cs.redisClient.Get(cs.ctx, key).Result()
		if err == nil {
			// Try to unmarshal as generic interface{}
			var result interface{}
			if json.Unmarshal([]byte(val), &result) == nil {
				return result
			}
			// Return raw string if unmarshal fails
			return val
		}
	}

	// Fallback to memory cache
	if value, found := cs.memoryCache.Get(key); found {
		return value
	}

	return nil
}

// GetString retrieves a string value from cache
func (cs *CacheService) GetString(key string) (string, bool) {
	// Try Redis first
	if cs.redisClient != nil {
		val, err := cs.redisClient.Get(cs.ctx, key).Result()
		if err == nil {
			return val, true
		}
	}

	// Fallback to memory cache
	if value, found := cs.memoryCache.Get(key); found {
		if str, ok := value.(string); ok {
			return str, true
		}
	}

	return "", false
}

// Delete removes a value from cache
func (cs *CacheService) Delete(key string) error {
	// Delete from Redis
	if cs.redisClient != nil {
		cs.redisClient.Del(cs.ctx, key)
	}

	// Delete from memory cache
	cs.memoryCache.Delete(key)
	return nil
}

// DeletePattern deletes all keys matching a pattern (Redis only)
func (cs *CacheService) DeletePattern(pattern string) error {
	if cs.redisClient == nil {
		return fmt.Errorf("Redis not available for pattern deletion")
	}

	keys, err := cs.redisClient.Keys(cs.ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return cs.redisClient.Del(cs.ctx, keys...).Err()
	}

	return nil
}

// Exists checks if a key exists in cache
func (cs *CacheService) Exists(key string) bool {
	// Check Redis first
	if cs.redisClient != nil {
		count, err := cs.redisClient.Exists(cs.ctx, key).Result()
		if err == nil && count > 0 {
			return true
		}
	}

	// Check memory cache
	_, found := cs.memoryCache.Get(key)
	return found
}

// Increment increments a numeric value in cache
func (cs *CacheService) Increment(key string, delta int64) (int64, error) {
	// Try Redis first
	if cs.redisClient != nil {
		val, err := cs.redisClient.IncrBy(cs.ctx, key, delta).Result()
		if err == nil {
			return val, nil
		}
	}

	// Fallback to memory cache (basic implementation)
	if value, found := cs.memoryCache.Get(key); found {
		if intVal, ok := value.(int64); ok {
			newVal := intVal + delta
			cs.memoryCache.Set(key, newVal, cache.DefaultExpiration)
			return newVal, nil
		}
	}

	// If key doesn't exist, set it to delta
	cs.memoryCache.Set(key, delta, cache.DefaultExpiration)
	return delta, nil
}

// SetExpiration sets expiration for an existing key (Redis only)
func (cs *CacheService) SetExpiration(key string, expiration time.Duration) error {
	if cs.redisClient == nil {
		return fmt.Errorf("Redis not available for setting expiration")
	}

	return cs.redisClient.Expire(cs.ctx, key, expiration).Err()
}

// GetTTL gets the time to live for a key
func (cs *CacheService) GetTTL(key string) (time.Duration, error) {
	if cs.redisClient == nil {
		return 0, fmt.Errorf("Redis not available for TTL")
	}

	return cs.redisClient.TTL(cs.ctx, key).Result()
}

// FlushAll clears all cache entries
func (cs *CacheService) FlushAll() error {
	// Flush Redis
	if cs.redisClient != nil {
		if err := cs.redisClient.FlushAll(cs.ctx).Err(); err != nil {
			return err
		}
	}

	// Flush memory cache
	cs.memoryCache.Flush()
	return nil
}

// GetStats returns cache statistics
func (cs *CacheService) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Memory cache stats
	stats["memory_items"] = cs.memoryCache.ItemCount()

	// Redis stats (if available)
	if cs.redisClient != nil {
		info, err := cs.redisClient.Info(cs.ctx, "memory").Result()
		if err == nil {
			stats["redis_info"] = info
		}
		stats["redis_available"] = true
	} else {
		stats["redis_available"] = false
	}

	return stats
}

// Close closes the cache connections
func (cs *CacheService) Close() {
	if cs.redisClient != nil {
		cs.redisClient.Close()
	}
}

// Health checks the health of cache services
func (cs *CacheService) Health() map[string]bool {
	health := make(map[string]bool)

	// Check memory cache (always healthy if service exists)
	health["memory_cache"] = true

	// Check Redis
	if cs.redisClient != nil {
		_, err := cs.redisClient.Ping(cs.ctx).Result()
		health["redis"] = err == nil
	} else {
		health["redis"] = false
	}

	return health
}