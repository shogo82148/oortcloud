package oortcloud

import (
	"sync"

	"github.com/Songmu/strrand"
)

type ConnectionMap struct {
	conMap map[string]Connection
	mu     sync.RWMutex
}

func (c *ConnectionMap) New(con Connection) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conMap == nil {
		c.conMap = map[string]Connection{}
	}

	for i := 0; i < 1000; i++ {
		id, err := strrand.RandomString("[0-9A-F]{32}")
		if err != nil {
			panic(err)
		}
		if _, ok := c.conMap[id]; !ok {
			c.conMap[id] = con
			return id
		}
	}

	panic("cannot create id")
}

func (c *ConnectionMap) Get(id string) (Connection, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conMap == nil {
		return nil, false
	}

	con, ok := c.conMap[id]
	return con, ok
}

func (c *ConnectionMap) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conMap == nil {
		return
	}

	if _, ok := c.conMap[id]; ok {
		delete(c.conMap, id)
	}
}
