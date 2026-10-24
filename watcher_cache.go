package goralens

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/cache"
)

func WrapCache(c cache.Cache) cache.Cache {
	if c == nil {
		return nil
	}
	if _, ok := c.(*watchedCache); ok {
		return c
	}
	return &watchedCache{Cache: c, driver: watchedDriver{Driver: c, store: "default"}}
}

type watchedCache struct {
	cache.Cache
	driver watchedDriver
}

func (c *watchedCache) Store(name string) cache.Driver {
	return watchedDriver{Driver: c.Cache.Store(name), store: name}
}

func (c *watchedCache) Add(key string, value any, t time.Duration) bool {
	return c.driver.Add(key, value, t)
}
func (c *watchedCache) Decrement(key string, value ...int64) (int64, error) {
	return c.driver.Decrement(key, value...)
}
func (c *watchedCache) Forever(key string, value any) bool { return c.driver.Forever(key, value) }
func (c *watchedCache) Forget(key string) bool             { return c.driver.Forget(key) }
func (c *watchedCache) Flush() bool                        { return c.driver.Flush() }
func (c *watchedCache) Get(key string, def ...any) any     { return c.driver.Get(key, def...) }
func (c *watchedCache) GetBool(key string, def ...bool) bool {
	return c.driver.GetBool(key, def...)
}
func (c *watchedCache) GetInt(key string, def ...int) int { return c.driver.GetInt(key, def...) }
func (c *watchedCache) GetInt64(key string, def ...int64) int64 {
	return c.driver.GetInt64(key, def...)
}
func (c *watchedCache) GetString(key string, def ...string) string {
	return c.driver.GetString(key, def...)
}
func (c *watchedCache) Has(key string) bool { return c.driver.Has(key) }
func (c *watchedCache) Increment(key string, value ...int64) (int64, error) {
	return c.driver.Increment(key, value...)
}
func (c *watchedCache) Put(key string, value any, t time.Duration) error {
	return c.driver.Put(key, value, t)
}
func (c *watchedCache) Pull(key string, def ...any) any { return c.driver.Pull(key, def...) }
func (c *watchedCache) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	return c.driver.Remember(key, ttl, callback)
}
func (c *watchedCache) RememberForever(key string, callback func() (any, error)) (any, error) {
	return c.driver.RememberForever(key, callback)
}
func (c *watchedCache) WithContext(ctx context.Context) cache.Driver {
	return c.driver.WithContext(ctx)
}

type watchedDriver struct {
	cache.Driver
	store string
	ctx   context.Context
}

func (d watchedDriver) record(kind, key string, value any, ttl time.Duration) {
	r := Instance()
	if r == nil || !r.config.watching(TypeCache) {
		return
	}
	content := map[string]any{"type": kind, "key": key, "store": d.store}
	if kind == "set" || kind == "hit" {
		if matchesAny(key, r.config.HiddenCacheKeys) {
			content["value"] = hiddenValue
		} else {
			content["value"] = exportValue(value)
		}
	}
	if ttl > 0 {
		content["expiration"] = int64(ttl.Seconds())
	}
	ctx := d.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	r.Record(ctx, newEntry(TypeCache, content))
}

func (d watchedDriver) recordRead(key string, found bool, value any) {
	if found {
		d.record("hit", key, value, 0)
	} else {
		d.record("missed", key, nil, 0)
	}
}

func (d watchedDriver) Add(key string, value any, t time.Duration) bool {
	ok := d.Driver.Add(key, value, t)
	if ok {
		d.record("set", key, value, t)
	}
	return ok
}

func (d watchedDriver) Decrement(key string, value ...int64) (int64, error) {
	n, err := d.Driver.Decrement(key, value...)
	if err == nil {
		d.record("set", key, n, 0)
	}
	return n, err
}

func (d watchedDriver) Forever(key string, value any) bool {
	ok := d.Driver.Forever(key, value)
	if ok {
		d.record("set", key, value, 0)
	}
	return ok
}

func (d watchedDriver) Forget(key string) bool {
	ok := d.Driver.Forget(key)
	d.record("forget", key, nil, 0)
	return ok
}

func (d watchedDriver) Flush() bool {
	ok := d.Driver.Flush()
	d.record("flush", "*", nil, 0)
	return ok
}

func (d watchedDriver) Get(key string, def ...any) any {
	found := d.Driver.Has(key)
	v := d.Driver.Get(key, def...)
	d.recordRead(key, found, v)
	return v
}

func (d watchedDriver) GetBool(key string, def ...bool) bool {
	found := d.Driver.Has(key)
	v := d.Driver.GetBool(key, def...)
	d.recordRead(key, found, v)
	return v
}

func (d watchedDriver) GetInt(key string, def ...int) int {
	found := d.Driver.Has(key)
	v := d.Driver.GetInt(key, def...)
	d.recordRead(key, found, v)
	return v
}

func (d watchedDriver) GetInt64(key string, def ...int64) int64 {
	found := d.Driver.Has(key)
	v := d.Driver.GetInt64(key, def...)
	d.recordRead(key, found, v)
	return v
}

func (d watchedDriver) GetString(key string, def ...string) string {
	found := d.Driver.Has(key)
	v := d.Driver.GetString(key, def...)
	d.recordRead(key, found, v)
	return v
}

func (d watchedDriver) Has(key string) bool {
	found := d.Driver.Has(key)
	d.recordRead(key, found, nil)
	return found
}

func (d watchedDriver) Increment(key string, value ...int64) (int64, error) {
	n, err := d.Driver.Increment(key, value...)
	if err == nil {
		d.record("set", key, n, 0)
	}
	return n, err
}

func (d watchedDriver) Put(key string, value any, t time.Duration) error {
	err := d.Driver.Put(key, value, t)
	if err == nil {
		d.record("set", key, value, t)
	}
	return err
}

func (d watchedDriver) Pull(key string, def ...any) any {
	found := d.Driver.Has(key)
	v := d.Driver.Pull(key, def...)
	d.recordRead(key, found, v)
	if found {
		d.record("forget", key, nil, 0)
	}
	return v
}

func (d watchedDriver) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	found := d.Driver.Has(key)
	v, err := d.Driver.Remember(key, ttl, callback)
	d.recordRead(key, found, v)
	if !found && err == nil {
		d.record("set", key, v, ttl)
	}
	return v, err
}

func (d watchedDriver) RememberForever(key string, callback func() (any, error)) (any, error) {
	found := d.Driver.Has(key)
	v, err := d.Driver.RememberForever(key, callback)
	d.recordRead(key, found, v)
	if !found && err == nil {
		d.record("set", key, v, 0)
	}
	return v, err
}

func (d watchedDriver) WithContext(ctx context.Context) cache.Driver {
	return watchedDriver{Driver: d.Driver.WithContext(ctx), store: d.store, ctx: ctx}
}
