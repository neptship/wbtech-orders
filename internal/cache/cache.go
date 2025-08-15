package cache

import (
	"sync"

	"github.com/neptship/wbtech-orders/internal/models"
)

type Cache interface {
	Set(id string, order models.Order)
	Get(id string) (models.Order, bool)
}

type MyCache struct {
	mu    *sync.RWMutex
	store map[string]models.Order
	order []string
	size  int
}

func NewCache(size int) Cache {
	return &MyCache{
		store: make(map[string]models.Order),
		order: make([]string, 0, size),
		size:  size,
		mu:    &sync.RWMutex{},
	}
}

func (c *MyCache) Set(id string, order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.store[id]; ok {
		c.store[id] = order
		return
	}
	if c.size > 0 && len(c.order) >= c.size {
		old := c.order[0]
		delete(c.store, old)
		c.order = c.order[1:]
	}
	c.store[id] = order
	c.order = append(c.order, id)
}

func (c *MyCache) Get(id string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[id]
	return val, ok
}
