package main

import (
	"log"
	"sync"
	"time"
)

type CooldownContainer struct {
	mu        sync.RWMutex
	container map[string]map[string]time.Time
}

func (c *CooldownContainer) Test(channel, command string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.container[channel][command]; !ok {
		return true
	}

	return time.Now().After(c.container[channel][command])
}

func (c *CooldownContainer) Set(channel, command, dutation string) {
	dut, err := time.ParseDuration(dutation)
	if err != nil {
		log.Println(err)
	}

	c.mu.Lock()
	c.container[channel][command] = time.Now().Add(dut)
	c.mu.Unlock()
}
