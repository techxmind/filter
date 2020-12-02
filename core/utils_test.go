package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToArray(t *testing.T) {
	tests := []struct {
		in    interface{}
		check []interface{}
	}{
		{
			1,
			[]interface{}{1},
		},
		{
			"a",
			[]interface{}{"a"},
		},
		{
			"a,b ,c,",
			[]interface{}{"a", "b", "c"},
		},
		{
			[]interface{}{"a", "b"},
			[]interface{}{"a", "b"},
		},
		{
			[]int{1, 2, 3},
			[]interface{}{1, 2, 3},
		},
		{
			//invalid
			map[int]int{1: 1},
			[]interface{}{map[int]int{1: 1}},
		},
	}

	for _, test := range tests {
		a := ToArray(test.in)
		assert.Equal(t, test.check, a)
	}
}

func TestIsArray(t *testing.T) {
	tests := []struct {
		in    interface{}
		check bool
	}{
		{1, false},
		{"a,b,c", false},
		{[]int{1}, true},
		{[1]int{1}, true},
		{[]interface{}{1}, true},
		{map[string]string{"a": "b"}, false},
		{nil, false},
	}

	for i, test := range tests {
		assert.Equal(t, test.check, IsArray(test.in), "case %d : %v", i, test.in)
	}
}
