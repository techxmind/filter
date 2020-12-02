package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	ctx := NewContext()
	ctx.Set("a", map[string]interface{}{
		"b": 1,
		"c": []interface{}{1, 2},
		"d": 5,
	})

	tests := []struct {
		item     []interface{}
		logic    GROUP_LOGIC
		expected bool
		hasError bool
	}{
		{
			item:     []interface{}{"ctx.a.b", "=", 1},
			logic:    LOGIC_ALL,
			expected: true,
		},
		{
			item:     []interface{}{"ctx.a.b", "=", 2},
			logic:    LOGIC_ALL,
			expected: false,
		},
		{
			item: []interface{}{
				[]interface{}{"ctx.a.b", "=", 1},
				[]interface{}{"ctx.a.c.0", "=", 1},
				[]interface{}{"ctx.a.d", "=", 5},
			},
			logic:    LOGIC_ALL,
			expected: true,
		},
		{
			item: []interface{}{
				[]interface{}{"ctx.a.b", "=", 2},
				[]interface{}{"ctx.a.c.0", "=", 1},
				[]interface{}{"ctx.a.d", "=", 5},
			},
			logic:    LOGIC_ALL,
			expected: false,
		},
		{
			item: []interface{}{
				[]interface{}{"ctx.a.b", "=", 2},
				[]interface{}{"ctx.a.c.0", "=", 1},
				[]interface{}{"ctx.a.d", "=", 5},
			},
			logic:    LOGIC_ANY,
			expected: true,
		},
		{
			item: []interface{}{
				[]interface{}{"ctx.a.b", "=", 2},
				[]interface{}{"ctx.a.c.0", "=", 22},
				[]interface{}{"ctx.a.d", "=", 55},
			},
			logic:    LOGIC_ANY,
			expected: false,
		},
		{
			item: []interface{}{
				[]interface{}{"ctx.a.b", "=", 1},
				[]interface{}{"ctx.a.c.0", "=", 1},
			},
			logic:    LOGIC_NONE,
			expected: false,
		},
		{
			item: []interface{}{
				[]interface{}{"ctx.a.b", "=", 2},
			},
			logic:    LOGIC_NONE,
			expected: true,
		},
		{
			item: []interface{}{
				[]interface{}{"ctx.a.b", "=", 2},
				[]interface{}{"ctx.a.c.0", "=", 1},
			},
			logic:    LOGIC_NONE,
			expected: false,
		},
		{
			item: []interface{}{
				[]interface{}{"ctx.a.b", "=", 2},
				[]interface{}{"ctx.a.c.0", "=", 1},
			},
			logic:    LOGIC_ANY_NOT,
			expected: true,
		},
		{
			item: []interface{}{
				[]interface{}{"ctx.a.b", "=", 1},
				[]interface{}{"ctx.a.c.0", "=", 1},
			},
			logic:    LOGIC_ANY_NOT,
			expected: false,
		},
		{
			item: []interface{}{
				"any?", "=>", []interface{}{
					[]interface{}{"ctx.a.b", "=", 2},
					[]interface{}{
						"none?", "=>", []interface{}{
							[]interface{}{"ctx.a.c.0", "=", 2},
						},
					},
				},
			},
			logic:    LOGIC_ALL,
			expected: true,
		},
	}

	for i, c := range tests {
		cond, err := NewCondition(c.item, c.logic)
		if c.hasError {
			assert.Error(t, err)
			continue
		}
		require.NoError(t, err, err)
		assert.Equal(t, c.expected, cond.Success(ctx), "case %d: %s", i, cond.String())
	}
}
