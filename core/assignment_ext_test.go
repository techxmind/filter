package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/techxmind/go-utils/itype"
	"github.com/techxmind/go-utils/object"
)

func TestProbabilitySet(t *testing.T) {
	ctx := NewContext()
	a := _assignmentFactory.Get("*=")
	v := []interface{}{
		[]interface{}{10, 10},
		[]interface{}{30, 30},
		[]interface{}{60, 60},
	}
	data := make(map[string]interface{})
	hit := make(map[int]int)
	val, err := a.PrepareValue(v)
	require.NoError(t, err)
	for i := 0; i < 10000; i++ {
		a.Run(ctx, data, "a.b", val)
		bi, _ := object.GetValue(data, "a.b")
		b := int(itype.Int(bi))
		assert.Contains(t, []int{10, 30, 60}, b)
		hit[b] += 1
	}
	t.Log("hits:", hit)
	assert.Equal(t, 3, len(hit), "hits.size = 3")
	assert.True(t, hit[60] > hit[30], "hits.60 > hits.30")
	assert.True(t, hit[30] > hit[10], "hits.60 > hits.30")
}

func TestGroupAssign(t *testing.T) {
	ctx := NewContext()
	data := make(map[string]interface{})
	a := _assignmentFactory.Get("=>")

	val1, err := a.PrepareValue([]interface{}{"a", "=", "a"})
	require.NoError(t, err)

	val2, err := a.PrepareValue([]interface{}{
		[]interface{}{"b", "=", "b"},
		[]interface{}{"c", "=", "c"},
	})
	require.NoError(t, err)

	a.Run(ctx, data, "set", val1)
	assert.Equal(t, map[string]interface{}{"a": "a"}, data)

	a.Run(ctx, data, "set", val2)
	assert.Equal(t, map[string]interface{}{"a": "a", "b": "b", "c": "c"}, data)
}
