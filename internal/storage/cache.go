package storage

import (
	"context"
	"time"
)

type CacheCmdResultString interface {
	Result() (string, error)
	Err() error
}

type CacheCmdResultInt64 interface {
	Result() (int64, error)
	Err() error
}

type CacheOps interface {
	Get(context.Context, string) CacheCmdResultString
	Set(context.Context, string, interface{}, time.Duration) CacheCmdResultString
	Del(context.Context, ...string) CacheCmdResultInt64
}
