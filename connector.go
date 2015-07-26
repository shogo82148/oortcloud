package oortcloud

import (
	"sync"

	"github.com/Songmu/strrand"
)

type idGenerator struct {
	chanMap map[string]chan Event
	mu      sync.RWMutex
}

// newId creates id and channel
func (c *idGenerator) newId() (string, chan Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := make(chan Event, 8)

	for i := 0; i < 1000; i++ {
		id, err := strrand.RandomString("[0-9A-F]{32}")
		if err != nil {
			panic(err)
		}
		if _, ok := c.chanMap[id]; !ok {
			c.chanMap[id] = ch
			return id, ch
		}
	}

	panic("cannot create id")
}

// deleteId deletes id from chanMap
func (c *idGenerator) deleteId(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ch, ok := c.chanMap[id]; ok {
		close(ch)
	}
	delete(c.chanMap, id)
}
