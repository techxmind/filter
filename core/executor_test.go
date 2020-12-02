package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func _e(k, op, val interface{}) []interface{} {
	return []interface{}{k, op, val}
}

func TestExecutor(t *testing.T) {
	data := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": []interface{}{1, 2},
		},
	}

	tests := []struct {
		input    []interface{}
		expected interface{}
		hasError bool
	}{
		{
			_e("foo.bar", "=", 1),
			map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": 1,
				},
			},
			false,
		},
		{
			[]interface{}{
				_e("foo.bar.0", "=", 2),
				_e("foo.zap", "=", 3),
			},
			map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": []interface{}{2, 2},
					"zap": 3,
				},
			},
			false,
		},
		{
			[]interface{}{
				"a", "*", "b",
			},
			nil,
			true,
		},
		{
			[]interface{}{
				"a", "*=", "b",
			},
			nil,
			true,
		},
	}

	ctx := NewContext()

	for i, c := range tests {
		e, err := NewExecutor(c.input)
		if c.hasError {
			t.Log(err)
			assert.Error(t, err)
			continue
		}
		d := Clone(data)
		e.Execute(ctx, d)
		assert.Equal(t, c.expected, d, "case %d: %v", i, c)
	}
}
