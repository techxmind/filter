package core

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type ctxKey string

const (
	filterDataCtxKey ctxKey = "data"
	cacheCtxKey      ctxKey = "cache"
	ctxDataCtxKey    ctxKey = "ctx"
	traceCtxKey      ctxKey = "trace"
)

type Context struct {
	ctx context.Context
	mu  sync.Mutex
}

func NewContext() *Context {
	return WithContext(context.Background())
}

type ContextOption interface {
	apply(*Context)
}

type ContextOptionFunc func(*Context)

func (f ContextOptionFunc) apply(c *Context) {
	f(c)
}

// WithCache ContextOption
func WithCache(cache Cache) ContextOption {
	return ContextOptionFunc(func(c *Context) {
		c.ctx = context.WithValue(c.ctx, cacheCtxKey, cache)
	})
}

// WithTrace ContextOption
func WithTrace(trace Trace) ContextOption {
	return ContextOptionFunc(func(c *Context) {
		c.ctx = context.WithValue(c.ctx, traceCtxKey, trace)
	})
}

func WithContext(ctx context.Context, opts ...ContextOption) *Context {
	if ctx == nil {
		ctx = context.Background()
	}

	// check context data
	if ctxData := ctx.Value(ctxDataCtxKey); ctxData == nil {
		ctx = context.WithValue(ctx, ctxDataCtxKey, newContextData())
	}

	c := &Context{
		ctx: ctx,
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}

// WithData return new *Context contains data.
// call every time the filter runs, to make data thread-safe
func WithData(ctx context.Context, data interface{}) *Context {
	ctx = context.WithValue(ctx, filterDataCtxKey, data)

	return WithContext(ctx)
}

// Data return filter data
func (c *Context) Data() interface{} {
	var data interface{}
	if data = c.ctx.Value(filterDataCtxKey); data == nil {
		c.mu.Lock()
		data = c.ctx.Value(filterDataCtxKey)
		if data == nil {
			data = make(map[string]interface{})
			c.ctx = context.WithValue(c.ctx, filterDataCtxKey, data)
		}
		c.mu.Unlock()
	}

	return data
}

// Cache return filter Cache object
func (c *Context) Cache() Cache {
	var cache interface{}

	if cache = c.ctx.Value(cacheCtxKey); cache == nil {
		c.mu.Lock()
		cache = c.ctx.Value(cacheCtxKey)
		if cache == nil {
			cache = NewCache()
			c.ctx = context.WithValue(c.ctx, cacheCtxKey, cache)
		}
		c.mu.Unlock()
	}

	return cache.(Cache)
}

// Trace return Trace
func (c *Context) Trace() Trace {
	if t := c.ctx.Value(traceCtxKey); t != nil {
		return t.(Trace)
	}

	return nil
}

// Set set context data
func (c *Context) Set(key string, value interface{}) {
	if data := c.ctx.Value(ctxDataCtxKey); data != nil {
		data.(*contextData).Set(key, value)
	}
}

// Delete delte context data
func (c *Context) Delete(key string) {
	if data := c.ctx.Value(ctxDataCtxKey); data != nil {
		data.(*contextData).Delete(key)
	}
}

// Get get context data
func (c *Context) Get(key string) (value interface{}, exists bool) {
	if data := c.ctx.Value(ctxDataCtxKey); data != nil {
		return data.(*contextData).Get(key)
	}

	return nil, false
}

// GetAll return a map contains all context data
func (c *Context) GetAll() map[string]interface{} {
	if data := c.ctx.Value(ctxDataCtxKey); data != nil {
		return data.(*contextData).GetAll()
	}

	return nil
}

// Implements context.Context interface
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Context) Err() error {
	return c.ctx.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// contextData
type contextData struct {
	mu sync.Mutex

	// context data
	m *sync.Map
	// for checking if readonly is old
	amended int32
	// context data readonly
	readonly map[string]interface{}
}

func newContextData() *contextData {
	return &contextData{
		m:        new(sync.Map),
		readonly: make(map[string]interface{}),
	}
}

func (d *contextData) Set(key string, value interface{}) {
	d.m.Store(key, value)
	atomic.AddInt32(&d.amended, 1)
}

func (d *contextData) Delete(key string) {
	d.m.Delete(key)
	atomic.AddInt32(&d.amended, 1)
}

func (d *contextData) Get(key string) (value interface{}, exists bool) {
	return d.m.Load(key)
}

func (d *contextData) GetAll() map[string]interface{} {
	if atomic.LoadInt32(&d.amended) == 0 {
		return d.readonly
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	old := atomic.LoadInt32(&d.amended)
	if old == 0 {
		return d.readonly
	}

	readonly := make(map[string]interface{})
	d.m.Range(func(key, val interface{}) bool {
		readonly[key.(string)] = val
		return true
	})
	atomic.CompareAndSwapInt32(&d.amended, old, 0)
	d.readonly = readonly

	return d.readonly
}
