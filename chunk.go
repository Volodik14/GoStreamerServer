package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

//var done = make(chan struct{})
//var buffer = make([]byte, 40000)

type chunk struct {
	mu     sync.RWMutex
	val    []byte
	buffer []byte
	done   chan struct{}
}

func (c *chunk) Load(f *os.File) {
	throttle := time.Tick(time.Second)
	for {
		<-throttle
		bytesread, err := f.Read(c.buffer)
		c.mu.Lock()
		c.val = c.buffer[:bytesread]
		c.mu.Unlock()

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			c.done <- struct{}{}
			break
		}
	}
}

func (c *chunk) Value() []byte {
	c.mu.RLock()
	val := c.val
	c.mu.RUnlock()
	return val
}
