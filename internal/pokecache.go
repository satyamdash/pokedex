package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	mu       sync.Mutex
	mpp      map[string]cacheEntry
	interval time.Duration
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		mpp:      make(map[string]cacheEntry),
		interval: interval,
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			c.reapLoop()
		}
	}()

	return c
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mpp[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.mpp[key].val) == 0 {
		return c.mpp[key].val, false
	}
	return c.mpp[key].val, true
}

func (c *Cache) reapLoop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.mpp {
		if time.Since(v.createdAt) >= c.interval {
			delete(c.mpp, k)
		}
	}
}
