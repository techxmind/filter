package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	assert.Implements(t, (*context.Context)(nil), NewContext())
}

func TestWithContext(t *testing.T) {
	pctx := context.WithValue(context.Background(), "foo", "bar")
	ctx := WithContext(pctx)

	assert.Equal(t, "bar", ctx.Value("foo"))
}

func TestData(t *testing.T) {
	assert.NotNil(t, NewContext().Data())
}

func TestCache(t *testing.T) {
	assert.NotNil(t, NewContext().Cache())
}

func TestGetSet(t *testing.T) {
	ctx := NewContext()

	ctx.Set("foo", "bar")
	v, exists := ctx.Get("foo")
	assert.Equal(t, "bar", v)
	assert.True(t, exists)
	v, exists = ctx.Get("bar")
	assert.Nil(t, v)
	assert.False(t, exists)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, ctx.GetAll())

	ctx.Set("bar", "zap")
	assert.Equal(t, map[string]interface{}{"foo": "bar", "bar": "zap"}, ctx.GetAll())

	ctx.Delete("foo")
	assert.Equal(t, map[string]interface{}{"bar": "zap"}, ctx.GetAll())
}

func Benchmark_Context(b *testing.B) {
	ctx := NewContext()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx.Data()
			ctx.Cache()
			ctx.Set("foo", "bar")
			ctx.Get("foo")
			ctx.Get("bar")
		}
	})
}
