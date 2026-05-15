package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// 缓存过期时间
	CacheL1TTL = 30 * time.Second  // 本地缓存 30 秒
	CacheL2TTL = 5 * time.Minute   // Redis 缓存 5 分钟
	
	// 缓存键前缀
	TaskCacheKeyPrefix = "task:"
	AllTasksCacheKey   = "tasks:all"
)

// LocalCache 本地内存缓存（L1）
type LocalCache struct {
	mu    sync.RWMutex
	store map[string]*CacheItem
}

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired 检查是否过期
func (ci *CacheItem) IsExpired() bool {
	return time.Now().After(ci.ExpiresAt)
}

var (
	// 本地缓存实例
	localCache = &LocalCache{
		store: make(map[string]*CacheItem),
	}
	
	// Redis 客户端
	redisClient *redis.Client
)

// InitCache 初始化缓存系统
func InitCache() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	// 测试 Redis 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := redisClient.Ping(ctx).Err(); err != nil {
		fmt.Println("⚠️  Redis 连接失败，仅使用本地缓存:", err.Error())
		redisClient = nil
	} else {
		fmt.Println("✓ Redis 连接成功")
	}
	
	// 启动清理过期本地缓存的后台任务
	go cleanupExpiredLocalCache()
}

// cleanupExpiredLocalCache 定期清理过期的本地缓存
func cleanupExpiredLocalCache() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		localCache.mu.Lock()
		for key, item := range localCache.store {
			if item.IsExpired() {
				delete(localCache.store, key)
			}
		}
		localCache.mu.Unlock()
	}
}

// setL1 设置本地缓存
func setL1(key string, value interface{}) {
	localCache.mu.Lock()
	defer localCache.mu.Unlock()
	
	localCache.store[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(CacheL1TTL),
	}
}

// getL1 获取本地缓存
func getL1(key string) (interface{}, bool) {
	localCache.mu.RLock()
	defer localCache.mu.RUnlock()
	
	item, exists := localCache.store[key]
	if !exists || item.IsExpired() {
		return nil, false
	}
	return item.Value, true
}

// deleteL1 删除本地缓存
func deleteL1(key string) {
	localCache.mu.Lock()
	defer localCache.mu.Unlock()
	delete(localCache.store, key)
}

// setL2 设置 Redis 缓存
func setL2(ctx context.Context, key string, value interface{}) error {
	if redisClient == nil {
		return nil // Redis 不可用，跳过
	}
	
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return redisClient.Set(ctx, key, data, CacheL2TTL).Err()
}

// getL2 获取 Redis 缓存
func getL2(ctx context.Context, key string, dest interface{}) (bool, error) {
	if redisClient == nil {
		return false, nil // Redis 不可用
	}
	
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // 缓存不存在
	}
	if err != nil {
		return false, err
	}
	
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, err
	}
	return true, nil
}

// deleteL2 删除 Redis 缓存
func deleteL2(ctx context.Context, key string) error {
	if redisClient == nil {
		return nil
	}
	return redisClient.Del(ctx, key).Err()
}

// invalidateTaskCache 清除指定任务的所有缓存
func invalidateTaskCache(ctx context.Context, id uint) {
	key := fmt.Sprintf("%s%d", TaskCacheKeyPrefix, id)
	deleteL1(key)
	deleteL2(ctx, key)
	
	// 同时清除列表缓存
	deleteL1(AllTasksCacheKey)
	deleteL2(ctx, AllTasksCacheKey)
}

// invalidateAllTasksCache 清除所有任务列表缓存
func invalidateAllTasksCache(ctx context.Context) {
	deleteL1(AllTasksCacheKey)
	deleteL2(ctx, AllTasksCacheKey)
}
