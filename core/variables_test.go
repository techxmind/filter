package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/techxmind/go-utils/itype"
)

func TestVariableSucc(t *testing.T) {
	v := _variableFactory.Create("succ")
	require.NotNil(t, v)

	val := GetVariableValue(NewContext(), v)
	assert.Equal(t, true, val)
}

func TestVariableRand(t *testing.T) {
	ctx := NewContext()
	v := _variableFactory.Create("rand")
	require.NotNil(t, v)
	hit := map[int64]bool{}
	for i := 0; i < 100; i++ {
		val := GetVariableValue(ctx, v)
		assert.GreaterOrEqual(t, val, 1)
		assert.LessOrEqual(t, val, 100)
		hit[itype.Int(val)] = true
	}
	assert.Greater(t, len(hit), 1)
}

func TestVariableTime(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2020-11-11T18:59:59Z")
	_currentTime = func() time.Time {
		return tm
	}
	defer func() {
		_currentTime = time.Now
	}()

	ctx := NewContext()
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"datetime", "2020-11-11 18:59:59"},
		{"date", "2020-11-11"},
		{"year", 2020},
		{"month", 11},
		{"day", 11},
		{"hour", 18},
		{"minute", 59},
		{"second", 59},
		{"unixtime", tm.Unix()},
		{"wday", 3},
	}

	for i, test := range tests {
		v := _variableFactory.Create(test.input)
		require.NotNil(t, v)
		assert.Equal(t, test.expected, GetVariableValue(ctx, v), "case %d: %s", i, test.input)
	}
}

func TestVariableData(t *testing.T) {
	ctx := NewContext()
	ctx = WithData(ctx, map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": []interface{}{1, "2", map[string]interface{}{
				"zap": true,
			}},
		},
	})

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"data.foo.bar.0", 1},
		{"data.foo.bar.2.zap", true},
		{"data.baz", nil},
	}

	for i, test := range tests {
		v := _variableFactory.Create(test.input)
		require.NotNil(t, v, test.input)
		assert.Equal(t, test.expected, GetVariableValue(ctx, v), "case %d: %s", i, test.input)
	}
}

func TestVariableCtx(t *testing.T) {
	ctx := WithContext(context.WithValue(context.Background(), "zap", "zap-ctx-value"))
	ctx = WithData(ctx, map[string]interface{}{
		"ctx": map[string]interface{}{
			"baz": "baz-in-data",
		},
	})
	ctx.Set("foo", map[string]interface{}{
		"bar": "bar-ctx",
	})
	ctx.Set("baz", "baz-ctx")

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"ctx.foo.bar", "bar-ctx"},
		{"ctx.baz", "baz-in-data"},
		{"ctx.zap", "zap-ctx-value"},
		{"ctx.other", nil},
	}

	for i, c := range tests {
		v := _variableFactory.Create(c.input)
		require.NotNil(t, v, c.input)
		assert.Equal(t, c.expected, GetVariableValue(ctx, v), "case %d : %s", i, c.input)
	}
}
