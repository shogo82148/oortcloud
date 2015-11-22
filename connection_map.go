package oortcloud

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ConnectionMap struct {
	rand   *rand.Rand
	conMap map[string]Connection
	mu     sync.RWMutex
}

func NewConnectionMap() *ConnectionMap {
	return &ConnectionMap{
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		conMap: map[string]Connection{},
	}
}

func (c *ConnectionMap) New(con Connection) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conMap == nil {
		panic("cannot create id")
	}

	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("%08X%08X%08X%08X", c.rand.Uint32(), c.rand.Uint32(), c.rand.Uint32(), c.rand.Uint32())
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
