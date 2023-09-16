package main

import (
	"fmt"
	"sync"
	"time"
	"net/http"
	"net/url"
	"io"
)

const (
	DefaultExpiration time.Duration = 0
	NoExpireTime time.Duration = -1
)


type Items struct {
	Value any
	ExpirationTime int64 // This is set to int64 since the default type for time.Duration is an int64 in nanoseconds. This is also why DefaultExpiration is time.Duration, so we can compare them.
	Expired bool
}

func (item Items) isExpired() bool {
	if item.ExpirationTime <= time.Now().UnixNano() && item.ExpirationTime > 0 {
		return true
	}

	return false
}

type Cache struct {
	items map[string]Items
	mu sync.Mutex
	Expiration time.Duration
}

func (c *Cache) SetItem(k string, d any, ex time.Duration) {
	c.mu.Lock() // Ensure no one causes a race condition
	defer c.mu.Unlock()
	var newExpireTime int64
	
	if ex == DefaultExpiration {
		ex = c.Expiration
	}

	if ex > -1 {
		newExpireTime = time.Now().Add(ex).UnixNano()
	}

	c.items[k] = Items{
		Value: d,
		ExpirationTime: newExpireTime, 
	}
}

func (c *Cache) GetItem(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Println(c.items[k])
}

func (c *Cache) DeleteItem(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, exists := c.items[k]
	if !exists {
		fmt.Println("Error, item does not exist and cannot be deleted.")
	}

	delete(c.items, k)
}

func (c *Cache) UpdateItem(k string, d any, ex time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.CheckExists(k) {
		c.setItem(k, d, ex)
	} else {
		fmt.Println("Key does not exist so unable to update item.")
	}
}

func (c *Cache) CheckExists(k string) bool {
	_, exists := c.items[k]
	if !exists {
		fmt.Println("Error: Key doesn't exist.")
		return false
	}

	return true
}

/* Below is non-exported functions used to create the same functionality while there are mutex locks to avoid deadlock. */

func (c *Cache) setItem(k string, d any, ex time.Duration) bool {
	var newExpireTime int64
	
	if ex == DefaultExpiration {
		ex = c.Expiration
	}

	if ex > -1 {
		newExpireTime = time.Now().Add(ex).UnixNano()
	}

	c.items[k] = Items{
		Value: d,
		ExpirationTime: newExpireTime, 
	}

	return true
}

/*func (c *Cache) retrieveItem(k string) {
	fmt.Println(c.items[k])
}*/

func (c *Cache) deleteItem(k string) bool {
	_, exists := c.items[k]
	if !exists {
		fmt.Println("Error, item does not exist and cannot be deleted.")
	}

	delete(c.items, k)
	return true
}

func (c *Cache) maintainCache(ch chan<- bool) {
	c.mu.Lock()
	for k, v := range c.items {
		if v.isExpired() {
			c.deleteItem(k)
		}
	}
	c.mu.Unlock()
	ch <- true
}

func (c *Cache) RetrievePage(page string) {
	url, err := url.ParseRequestURI(page)
	if err != nil {
		fmt.Println("URL is not valid.")
	}
	nurl := url.String()
	resp, err := http.Get(url.String())
	if err != nil {
		fmt.Println("error getting page")
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
	}

	c.SetItem(nurl, body, DefaultExpiration)

}

func main() {
	it := make(map[string]Items)

	cache := &Cache{
		items: it,
		Expiration: DefaultExpiration,
	}

	cache.RetrievePage("https://google.com")

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	done := make(chan bool)

	fmt.Println(NoExpireTime)

	cache.SetItem("test", 12341234, time.Second * 20)
	cache.SetItem("test2", 234523634574, NoExpireTime)
	cache.SetItem("test5", "asfgdfshsdfhshsh", time.Second * 10)
	cache.GetItem("test")
	cache.GetItem("test2")
	cache.GetItem("test5")
	cache.UpdateItem("test1034234", 234123324, NoExpireTime)

	for {
		select {
		case <-done:
			fmt.Println("Cleaned cache.")
		case <-ticker.C:
			go cache.maintainCache(done)
			fmt.Println("Tick")
			cache.GetItem("https://google.com")
			cache.GetItem("test")
			cache.GetItem("test2")
			cache.GetItem("test5")
		}
	}
}


