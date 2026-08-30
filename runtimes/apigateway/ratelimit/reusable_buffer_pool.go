package ratelimiter

import (
	"bytes"
	"sync"
)

type ReusableBufferPool struct {
	pool sync.Pool
}

func NewReusableBufferPool() *ReusableBufferPool {
	return &ReusableBufferPool{}
}

func (p *ReusableBufferPool) Get() *bytes.Buffer {
	buffer := p.pool.Get()
	if buffer == nil {
		return &bytes.Buffer{}
	}
	return buffer.(*bytes.Buffer)
}

func (p *ReusableBufferPool) Put(buffer *bytes.Buffer) {
	p.pool.Put(buffer)
}
