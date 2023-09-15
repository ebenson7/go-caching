package main

import (
	"sync"
	"time"
	"fmt"
)

const DefaultExpiration time.Duration = -1 // This is bad practice normally, but for local development this will be fine until I add expiration features.

type Items struct {
	Value any
	ExpirationTime int64 // This is set to int64 since the default type for time.Duration is an int64 in nanoseconds. This is also why DefaultExpiration is time.Duration, so we can compare them.
}

/*func (item Items) isExpired() bool {
	if item.ExpirationTime == 0 {
		return false
	}

	return time.Now().UnixNano() > item.ExpirationTime
}*/

type Cache struct {
	items map[string]Items
	mu sync.Mutex
	Expiration time.Duration
}

func (c *Cache) SetItem(k string, v any, ex time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var newExpireTime int64
	
	if ex == DefaultExpiration {
		ex = c.Expiration
	}

	if ex > -1 {
		newExpireTime = time.Now().Add(ex).UnixNano()
	}

	c.items[k] = Items{
		Value: v,
		ExpirationTime: newExpireTime, 
	}
}

func (c *Cache) GetItem(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	fmt.Println(c.items[k])
}

func main() {
	it := make(map[string]Items)
	
	cache := &Cache{
		items: it,
		Expiration: DefaultExpiration,
	}

	cache.SetItem("test", 12341234, -1)
	cache.SetItem("test2", 234523634574, -1)
	cache.SetItem("test5", "asfgdfshsdfhshsh", -1)
	cache.GetItem("test")
	cache.GetItem("test2")
	cache.GetItem("test5")
}


