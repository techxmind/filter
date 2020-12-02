package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/techxmind/filter/core"
	"github.com/techxmind/go-utils/itype"
)

func TestBuildFilter(t *testing.T) {
	f, err := buildFilter(arr(
		arr("time", "between", "10:00:00,19:00:00"),
		arr("hit", "=", true),
	))
	assert.NoError(t, err)
	fl, ok := f.(*singleFilter)
	require.True(t, ok)
	t.Logf("filter name:%s", fl.name)
	assert.NotNil(t, fl.condition)
	assert.NotNil(t, fl.executor)

	f, err = buildFilter(arr(
		"filter-name",
		arr("time", "between", "10:00:00,19:00:00"),
		arr("hit", "=", true),
	))
	require.NoError(t, err)
	assert.Equal(t, "filter-name", f.Name())

	f, err = buildFilter(arr(
		arr("time", "between", "10:00:00,19:00:00"),
	))
	assert.Error(t, err)

	f, err = buildFilter(arr(
		arr("time", "between", "10:00:00,19:00:00"),
		nil,
	))
	require.NoError(t, err)
	assert.Nil(t, f.(*singleFilter).executor)

	f, err = buildFilter(arr(
		arr("time", "between", "10:00:00,19:00:00"),
		nil,
	), Name("filter-name"))
	require.NoError(t, err)
	assert.Equal(t, "filter-name", f.Name())

	f, err = buildFilter(arr(
		arr("ctx.foo", "=", "bar"),
		arr("hello", "=", "world"),
	), Name("filter-name"), NamePrefix("my-"))
	require.NoError(t, err)
	assert.Equal(t, "my-filter-name", f.Name())

	ctx := core.NewContext()
	ctx.Set("foo", "bar")
	data := make(map[string]interface{})
	ok = f.Run(ctx, data)
	assert.True(t, ok)
	assert.Equal(t, "world", data["hello"])
}

func TestNew(t *testing.T) {
	ctx := core.NewContext()
	ctx.Set("foo", "bar")
	ctx.Set("bar", "zap")
	ctx.Set("zap", "baz")
	data := make(map[string]interface{})

	f, err := New(arr(
		arr("ctx.foo", "=", "bar"),
		arr("a", "=", 1),
	))
	require.NoError(t, err)
	require.True(t, f.Run(ctx, data))
	assert.Equal(t, 1, data["a"])

	getItems := func(v interface{}) []interface{} {
		return arr(
			arr(
				arr("ctx.foo", "=", "bar"),
				arr("a", "=", v),
			),

			arr(
				arr("ctx.bar", "=", "zap"),
				arr("b", "=", v),
			),
		)
	}

	f, err = New(getItems(2))
	require.NoError(t, err)
	require.True(t, f.Run(ctx, data))
	assert.Equal(t, 2, data["a"])
	assert.Equal(t, 2, data["b"])

	f, err = New(getItems(3), ShortMode(true))
	require.NoError(t, err)
	require.True(t, f.Run(ctx, data))
	assert.Equal(t, 3, data["a"])
	assert.Equal(t, 2, data["b"])

	f1, err := New(arr(
		arr("succ", "=", true),
		arr("a", "=", 1),
	))
	f2, err := New(arr(
		arr("succ", "=", true),
		arr("a", "=", 2),
	))
	f3, err := New(arr(
		arr("succ", "=", true),
		arr("a", "=", 3),
	))
	g := NewFilterGroup(EnableRank(true))
	g.Add(f1, Weight(10), Priority(3))
	g.Add(f2, Weight(5), Priority(3))
	g.Add(f3, Weight(100), Priority(1))

	hit := make(map[int]int)
	for i := 1; i < 10000; i++ {
		data := make(map[string]interface{})
		g.Run(ctx, data)
		v := int(itype.Int(data["a"]))
		hit[v] += 1
	}

	require.Equal(t, 2, len(hit), "hit.size = 2")
	require.True(t, hit[2] < hit[1], "hit.2 < hit.1")
	t.Log("hit:", hit)
}
