package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"web-demo/enterprise/config"
)

const (
	TaskCacheKeyPrefix = "task:"
	AllTasksCacheKey   = "tasks:all"
)

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired 检查是否过期
func (ci *CacheItem) IsExpired() bool {
	return time.Now().After(ci.ExpiresAt)
}

// Cache 缓存系统（L1 本地内存 + L2 Redis）
type Cache struct {
	mu          sync.RWMutex
	store       map[string]*CacheItem
	redisClient *redis.Client
	l1TTL       time.Duration
	l2TTL       time.Duration
	log         zerolog.Logger
}

// NewForTest 创建测试用的缓存（不连接 Redis）
func NewForTest(log zerolog.Logger) *Cache {
	return &Cache{
		store: make(map[string]*CacheItem),
		l1TTL: 30 * time.Second,
		l2TTL: 5 * time.Minute,
		log:   log,
	}
}

// New 创建缓存系统
func New(cfg *config.Config, log zerolog.Logger) *Cache {
	c := &Cache{
		store: make(map[string]*CacheItem),
		l1TTL: cfg.Cache.L1TTL,
		l2TTL: cfg.Cache.L2TTL,
		log:   log,
	}

	log.Info().Msg("初始化 Redis 客户端...")
	c.redisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试 Redis 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.redisClient.Ping(ctx).Err(); err != nil {
		log.Warn().Err(err).Msg("Redis 连接失败，仅使用本地缓存")
		c.redisClient = nil
	} else {
		log.Info().Msg("Redis 连接成功")
	}

	// 启动清理过期本地缓存的后台任务
	log.Info().Msg("启动缓存清理后台任务...")
	go c.cleanup()

	return c
}

// cleanup 定期清理过期的本地缓存
func (c *Cache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		count := 0
		for key, item := range c.store {
			if item.IsExpired() {
				delete(c.store, key)
				count++
			}
		}
		c.mu.Unlock()

		if count > 0 {
			c.log.Debug().Int("count", count).Msg("清理过期本地缓存")
		}
	}
}

// SetL1 设置本地缓存
func (c *Cache) SetL1(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(c.l1TTL),
	}
	c.log.Debug().Str("key", key).Dur("ttl", c.l1TTL).Msg("CACHE SET: L1")
}

// GetL1 获取本地缓存
func (c *Cache) GetL1(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.store[key]
	if !exists || item.IsExpired() {
		return nil, false
	}
	return item.Value, true
}

// DeleteL1 删除本地缓存
func (c *Cache) DeleteL1(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.store[key]; exists {
		delete(c.store, key)
		c.log.Debug().Str("key", key).Msg("CACHE DELETE: L1")
	}
}

// SetL2 设置 Redis 缓存（使用默认 L2 TTL）
func (c *Cache) SetL2(ctx context.Context, key string, value interface{}) error {
	return c.SetL2WithTTL(ctx, key, value, c.l2TTL)
}

// SetL2WithTTL 设置 Redis 缓存（使用自定义 TTL）
func (c *Cache) SetL2WithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.redisClient == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		c.log.Error().Err(err).Msg("Redis 序列化失败")
		return err
	}

	if err := c.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		c.log.Error().Err(err).Msg("Redis SET 失败")
		return err
	}

	c.log.Debug().Str("key", key).Dur("ttl", ttl).Msg("CACHE SET: L2")
	return nil
}

// GetL2 获取 Redis 缓存
func (c *Cache) GetL2(ctx context.Context, key string, dest interface{}) (bool, error) {
	if c.redisClient == nil {
		return false, nil
	}

	val, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		c.log.Error().Err(err).Msg("Redis GET 失败")
		return false, err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.log.Error().Err(err).Msg("Redis 反序列化失败")
		return false, err
	}
	return true, nil
}

// DeleteL2 删除 Redis 缓存
func (c *Cache) DeleteL2(ctx context.Context, key string) error {
	if c.redisClient == nil {
		return nil
	}
	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		c.log.Error().Err(err).Msg("Redis DELETE 失败")
		return err
	}
	c.log.Debug().Str("key", key).Msg("CACHE DELETE: L2")
	return nil
}

// InvalidateTask 清除指定任务的所有缓存
func (c *Cache) InvalidateTask(ctx context.Context, id uint) {
	key := fmt.Sprintf("%s%d", TaskCacheKeyPrefix, id)
	c.DeleteL1(key)
	c.DeleteL2(ctx, key)

	// 同时清除列表缓存
	c.DeleteL1(AllTasksCacheKey)
	c.DeleteL2(ctx, AllTasksCacheKey)
}

// InvalidateAllTasks 清除所有任务列表缓存
func (c *Cache) InvalidateAllTasks(ctx context.Context) {
	c.DeleteL1(AllTasksCacheKey)
	c.DeleteL2(ctx, AllTasksCacheKey)
}

// L1TTL 返回 L1 缓存 TTL
func (c *Cache) L1TTL() time.Duration {
	return c.l1TTL
}

// L2TTL 返回 L2 缓存 TTL
func (c *Cache) L2TTL() time.Duration {
	return c.l2TTL
}

// Ping 检查 Redis 连接
func (c *Cache) Ping(ctx context.Context) error {
	if c.redisClient == nil {
		return fmt.Errorf("redis not available")
	}
	return c.redisClient.Ping(ctx).Err()
}

// Close 关闭 Redis 连接
func (c *Cache) Close() {
	if c.redisClient != nil {
		c.log.Info().Msg("关闭 Redis 连接...")
		c.redisClient.Close()
	}
}
