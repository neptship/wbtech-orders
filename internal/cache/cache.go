package cache

import (
	"hash/fnv"
	"runtime"
	"sync"

	"github.com/neptship/wbtech-orders/internal/models"
)

type Cache interface {
	Set(id string, order models.Order)
	Get(id string) (models.Order, bool)
}

type shard struct {
	mu    sync.RWMutex
	store map[string]models.Order
}

type MyCache struct {
	shards []shard
}

func NewCache(_ int) Cache {
	n := runtime.NumCPU() * 2
	if n < 4 {
		n = 4
	}
	ss := make([]shard, n)
	for i := range ss {
		ss[i] = shard{store: make(map[string]models.Order)}
	}
	return &MyCache{shards: ss}
}

func (c *MyCache) Set(id string, order models.Order) {
	s := c.shardFor(id)
	s.mu.Lock()
	s.store[id] = order
	s.mu.Unlock()
}

func (c *MyCache) Get(id string) (models.Order, bool) {
	s := c.shardFor(id)
	s.mu.RLock()
	v, ok := s.store[id]
	s.mu.RUnlock()
	return v, ok
}

func (c *MyCache) shardFor(id string) *shard {
	if len(c.shards) == 1 {
		return &c.shards[0]
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	idx := int(h.Sum32() % uint32(len(c.shards)))
	return &c.shards[idx]
}
