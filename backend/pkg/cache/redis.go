// Package cache provides a small cache-aside wrapper over Redis. Callers pass
// plain strings (usually pre-marshaled JSON) in and out; encoding stays the
// caller's concern so this package has zero knowledge of domain types.
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(addr, password string, db int) *Cache {
	return &Cache{client: redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})}
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error { return c.client.Close() }

func (c *Cache) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// IncrWithExpiry atomically increments key and, only on the first increment
// (count == 1), sets its TTL — implementing a fixed-window counter suitable
// for rate limiting without a read-then-write race.
func (c *Cache) IncrWithExpiry(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		c.client.Expire(ctx, key, ttl)
	}
	return count, nil
}

// DeleteByPrefix scans (never KEYS, which blocks the server on a large keyspace)
// for keys starting with prefix and deletes them. Used for coarse invalidation
// like "every cached search result" after a bulk import changes verse content.
func (c *Cache) DeleteByPrefix(ctx context.Context, prefix string) error {
	iter := c.client.Scan(ctx, 0, prefix+"*", 200).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 500 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}
