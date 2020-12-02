package core

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/techxmind/go-utils/object"
)

type AssignmentTestSuite struct {
	suite.Suite
	ctx *Context
}

type assignTestCase struct {
	input        []interface{}
	expected     []interface{}
	prepareError bool
}

func (s *AssignmentTestSuite) SetupTest() {
	ctx := NewContext()
	data := map[string]interface{}{
		"area": map[string]interface{}{
			"zipcode": 200211,
			"city":    "shanghai",
		},
		"birthday":  "1985-11-21",
		"height":    "178",
		"age":       25,
		"fav_books": "book1,book2,book3",
		"pets":      []interface{}{"dog", "cat"},
	}
	ctx = WithData(ctx, data)
	s.ctx = ctx
}

func (s *AssignmentTestSuite) TestEqualAssignment() {
	tests := []assignTestCase{
		{
			input: []interface{}{"area.province", "=", "shanghai"},
			expected: []interface{}{"area", map[string]interface{}{
				"zipcode":  200211,
				"city":     "shanghai",
				"province": "shanghai",
			}},
		},
		{
			input: []interface{}{"pets.0", "=", "rabbit"},
			expected: []interface{}{"pets", []interface{}{
				"rabbit",
				"cat",
			}},
		},
		{
			input: []interface{}{"assets.house.area", "=", 100.4},
			expected: []interface{}{"assets", map[string]interface{}{
				"house": map[string]interface{}{
					"area": 100.4,
				},
			}},
		},
	}

	s.testCases(tests)
}

func (s *AssignmentTestSuite) TestDeleteAssignment() {
	tests := []assignTestCase{
		{
			input:    []interface{}{"area", "-", "city,zipcode"},
			expected: []interface{}{"area", map[string]interface{}{}},
		},
		{
			input:    []interface{}{"$", "-", []interface{}{"height"}},
			expected: []interface{}{"height", nil},
		},
		{
			input:        []interface{}{"area", "-", 0},
			prepareError: true,
		},
	}

	s.testCases(tests)
}

func (s *AssignmentTestSuite) TestMergeAssignment() {
	tests := []assignTestCase{
		{
			input: []interface{}{"area", "+", map[string]interface{}{
				"province": "shanghai",
			}},
			expected: []interface{}{"area", map[string]interface{}{
				"zipcode":  200211,
				"city":     "shanghai",
				"province": "shanghai",
			}},
		},
		{
			input: []interface{}{"assets.house", "+", map[string]interface{}{
				"area": 100,
			}},
			expected: []interface{}{"assets", map[string]interface{}{
				"house": map[string]interface{}{
					"area": 100,
				},
			}},
		},
		{
			input:        []interface{}{"assets.car", "+", "lexus"},
			prepareError: true,
		},
	}

	s.testCases(tests)
}

func (s *AssignmentTestSuite) testCases(tests []assignTestCase) {
	for i, c := range tests {
		a := _assignmentFactory.Get(c.input[1].(string))
		v, err := a.PrepareValue(c.input[2])
		if c.prepareError {
			s.Error(err)
			continue
		}
		require.NoError(s.T(), err)
		a.Run(s.ctx, s.ctx.Data(), c.input[0].(string), v)

		cv, _ := object.GetValue(s.ctx.Data(), c.expected[0].(string))

		s.Equal(c.expected[1], cv, "case %d: %v", i, c.input)
	}
}

func TestAssignmentTestSuite(t *testing.T) {
	suite.Run(t, new(AssignmentTestSuite))
}
