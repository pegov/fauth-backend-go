package storage

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type MemoryCache struct {
	memory map[string]MemoryCacheRecord
}

func NewMemoryCache() CacheOps {
	return &MemoryCache{
		memory: make(map[string]MemoryCacheRecord),
	}
}

type MemoryCacheRecord struct {
	value string
	exp   time.Time
}

type MemoryCacheResultString struct {
	value string
	err   error
}

func (r *MemoryCacheResultString) Result() (string, error) {
	return r.value, r.err
}

func (r *MemoryCacheResultString) Err() error {
	return r.err
}

type MemoryCacheResultInt64 struct {
	value int64
	err   error
}

func (r *MemoryCacheResultInt64) Result() (int64, error) {
	return r.value, r.err
}

func (r *MemoryCacheResultInt64) Err() error {
	return r.err
}

var (
	ErrNil = errors.New("nil")
)

func (r *MemoryCache) Get(ctx context.Context, key string) CacheCmdResultString {
	record, ok := r.memory[key]
	if !ok {
		return &MemoryCacheResultString{
			err: ErrNil,
		}
	}

	if time.Now().After(record.exp) {
		return &MemoryCacheResultString{
			err: ErrNil,
		}
	}

	return &MemoryCacheResultString{
		value: record.value,
	}
}

func (r *MemoryCache) Set(
	ctx context.Context,
	key string,
	value interface{},
	expiration time.Duration,
) CacheCmdResultString {
	exp := time.Now().Add(expiration)
	s := fmt.Sprintf("%v", value)
	r.memory[key] = MemoryCacheRecord{
		value: s,
		exp:   exp,
	}

	return &MemoryCacheResultString{
		value: s,
	}
}

func (r *MemoryCache) Del(
	ctx context.Context,
	keys ...string,
) CacheCmdResultInt64 {
	var v int64
	for _, key := range keys {
		if _, ok := r.memory[key]; ok {
			v += 1
			delete(r.memory, key)
		}
	}
	return &MemoryCacheResultInt64{
		value: v,
	}
}
