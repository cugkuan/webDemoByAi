package cache

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func newTestCache(t *testing.T) *Cache {
	t.Helper()
	log := zerolog.New(zerolog.NewTestWriter(t))
	return &Cache{
		store: make(map[string]*CacheItem),
		l1TTL: 30 * time.Second,
		l2TTL: 5 * time.Minute,
		log:   log,
	}
}

func TestCache_SetAndGetL1(t *testing.T) {
	c := newTestCache(t)

	c.SetL1("key1", "value1")
	val, ok := c.GetL1("key1")
	if !ok {
		t.Error("GetL1 should return true for existing key")
	}
	if val != "value1" {
		t.Errorf("GetL1 = %v, want value1", val)
	}
}

func TestCache_GetL1_Missing(t *testing.T) {
	c := newTestCache(t)

	_, ok := c.GetL1("nonexistent")
	if ok {
		t.Error("GetL1 should return false for missing key")
	}
}

func TestCache_GetL1_Expired(t *testing.T) {
	c := newTestCache(t)
	c.l1TTL = -1 * time.Second // 立即过期

	c.SetL1("key", "value")
	time.Sleep(10 * time.Millisecond)

	_, ok := c.GetL1("key")
	if ok {
		t.Error("GetL1 should return false for expired key")
	}
}

func TestCache_DeleteL1(t *testing.T) {
	c := newTestCache(t)

	c.SetL1("key", "value")
	c.DeleteL1("key")

	_, ok := c.GetL1("key")
	if ok {
		t.Error("GetL1 should return false after DeleteL1")
	}
}

func TestCache_DeleteL1_NonExistent(t *testing.T) {
	c := newTestCache(t)
	// 删除不存在的 key 不应 panic
	c.DeleteL1("nonexistent")
}

func TestCache_InvalidateTask(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()

	// 设置任务缓存和列表缓存
	c.SetL1("task:1", &struct{ ID uint }{ID: 1})
	c.SetL1(AllTasksCacheKey, []interface{}{})

	// 清除任务缓存
	c.InvalidateTask(ctx, 1)

	// 验证任务缓存和列表缓存都被清除
	if _, ok := c.GetL1("task:1"); ok {
		t.Error("task:1 should be deleted")
	}
	if _, ok := c.GetL1(AllTasksCacheKey); ok {
		t.Error("tasks:all should be deleted")
	}
}

func TestCache_InvalidateAllTasks(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()

	c.SetL1(AllTasksCacheKey, []interface{}{})
	c.InvalidateAllTasks(ctx)

	if _, ok := c.GetL1(AllTasksCacheKey); ok {
		t.Error("tasks:all should be deleted")
	}
}

func TestCache_L1TTLandL2TTL(t *testing.T) {
	c := newTestCache(t)

	if c.L1TTL() != 30*time.Second {
		t.Errorf("L1TTL = %v, want 30s", c.L1TTL())
	}
	if c.L2TTL() != 5*time.Minute {
		t.Errorf("L2TTL = %v, want 5m", c.L2TTL())
	}
}

func TestCache_Ping_NoRedis(t *testing.T) {
	c := newTestCache(t)
	// 没有 redisClient，Ping 应该返回错误
	err := c.Ping(context.Background())
	if err == nil {
		t.Error("Ping should return error when redis is not available")
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := newTestCache(t)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := "key"
			c.SetL1(key, n)
			c.GetL1(key)
			c.DeleteL1(key)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCacheItem_IsExpired(t *testing.T) {
	tests := []struct {
		name string
		item *CacheItem
		want bool
	}{
		{
			name: "not expired",
			item: &CacheItem{ExpiresAt: time.Now().Add(1 * time.Hour)},
			want: false,
		},
		{
			name: "expired",
			item: &CacheItem{ExpiresAt: time.Now().Add(-1 * time.Hour)},
			want: true,
		},
		{
			name: "just expired",
			item: &CacheItem{ExpiresAt: time.Now().Add(-1 * time.Millisecond)},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
