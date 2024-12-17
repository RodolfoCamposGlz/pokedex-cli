package cache

import (
	"sync"
	"time"
	"fmt"
)

type cacheEntry struct {
	createdAt time.Time
	val []byte
}

type Cache struct {
	mu    sync.Mutex              // Protects access to the map
	store map[string]cacheEntry   // Cache storage
	interval  time.Duration         // Time duration for cleaning stale entries
}

func NewCache(cleanupInterval time.Duration) *Cache{
	cache := &Cache {
		store: make(map[string]cacheEntry),
		interval: cleanupInterval,
	}
	go cache.reapLoop()
	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = cacheEntry {
		createdAt: time.Now(),
		val: val,
	} 
	fmt.Printf("Added key: %s\n", key)
}

func (c *Cache) Get(key string) ([]byte, bool){
	if entry, exists := c.store[key]; exists {
		fmt.Println("Found:")
		return entry.val, exists
	} else {
		fmt.Println("Key does not exist")
		return nil, exists
	}
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now:= time.Now()
		for key, entry := range c.store {
			if now.Sub(entry.createdAt) > c.interval {
				delete(c.store, key)
			}
		}
		c.mu.Unlock()
	}
}