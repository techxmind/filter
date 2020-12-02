package core

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OperationTestSuite struct {
	suite.Suite
	ctx *Context
}

type opTestCase struct {
	input         []interface{}
	expected      bool
	expectedError bool
}

func (s *OperationTestSuite) SetupTest() {
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

func (s *OperationTestSuite) TestEqual() {
	tests := []opTestCase{
		{[]interface{}{"succ", "=", true}, true, false},
		{[]interface{}{"succ", "=", 1}, true, false},
		{[]interface{}{"data.area.zipcode", "=", 200211}, true, false},
		{[]interface{}{"data.area.zipcode", "=", "200211"}, true, false},
		{[]interface{}{"data.age", "=", 25}, true, false},
		{[]interface{}{"data.age", "=", 24}, false, false},
		{[]interface{}{"data.age", "=", "foo"}, false, false},
		{[]interface{}{"data.area", "=", true}, true, false},
	}

	s.testCases(tests)

	s.testCases(s.getOppositeCases(tests, map[string]string{"=": "!="}))
}

func (s *OperationTestSuite) TestCompare() {
	tests := []opTestCase{
		{[]interface{}{"data.age", ">", 24}, true, false},
		{[]interface{}{"data.age", ">", 25}, false, false},
		{[]interface{}{"data.age", ">=", 25}, true, false},
		{[]interface{}{"data.age", "<", 26}, true, false},
		{[]interface{}{"data.age", "<=", 25}, true, false},
		{[]interface{}{"data.age", "<", 25}, false, false},
		{[]interface{}{"data.age", "<", 25}, false, false},
		{[]interface{}{"data.age", "<", "foo"}, false, false},
		{[]interface{}{"data.height", "<", 177}, false, false},
		{[]interface{}{"data.height", ">", 179}, false, false},
	}

	s.testCases(tests)

	s.testCases(s.getOppositeCases(tests, map[string]string{
		">":  "<=",
		">=": "<",
		"<":  ">=",
		"<=": ">",
	}))
}

func (s *OperationTestSuite) TestBetween() {
	tests := []opTestCase{
		{[]interface{}{"data.age", "between", "24,26"}, true, false},
		{[]interface{}{"data.age", "between", []interface{}{24, 26}}, true, false},
		{[]interface{}{"data.age", "between", "24,26,27"}, false, true},
		{[]interface{}{"data.birthday", "between", "1984-11-11,1986-11-11"}, true, false},
	}

	s.testCases(tests)
}

func (s *OperationTestSuite) TestIn() {
	tests := []opTestCase{
		{[]interface{}{"data.age", "in", "24,25 ,26"}, true, false},
		{[]interface{}{"data.age", "in", "24,26"}, false, false},
		{[]interface{}{"data.age", "in", []interface{}{24, 25, 26}}, true, false},
		{[]interface{}{"data.age", "in", []interface{}{24, 26}}, false, false},
	}

	s.testCases(tests)
	s.testCases(s.getOppositeCases(tests, map[string]string{"in": "not in"}))
}

func (s *OperationTestSuite) TestMatch() {
	tests := []opTestCase{
		{[]interface{}{"data.area.city", "~", "shang"}, true, false},
		{[]interface{}{"data.area.city", "~", "hai"}, true, false},
		{[]interface{}{"data.area.city", "~", "h.i"}, false, false},
		{[]interface{}{"data.area.city", "~", "/h.i/"}, true, false},
		{[]interface{}{"data.area.city", "~", "/^s.+i$/"}, true, false},
		{[]interface{}{"data.area.city", "~", "/^(s.+i$/"}, false, true},
	}

	s.testCases(tests)
	s.testCases(s.getOppositeCases(tests, map[string]string{"~": "!~"}))
}

func (s *OperationTestSuite) TestListMatch() {
	tests := []opTestCase{
		{[]interface{}{"data.fav_books", "has", "book1,book3"}, true, false},
		{[]interface{}{"data.fav_books", "has", "book1,book3,book4"}, false, false},
		{[]interface{}{"data.pets", "any", "cat,pig"}, true, false},
		{[]interface{}{"data.pets", "any", "rabbit,fly"}, false, false},
		{[]interface{}{"data.pets", "none", []interface{}{"pig", "rabbit"}}, true, false},
		{[]interface{}{"data.pets", "none", []interface{}{"dog", "rabbit"}}, false, false},
	}

	s.testCases(tests)
	s.testCases(s.getOppositeCases(tests, map[string]string{"any": "none", "none": "any"}))
}

func (s *OperationTestSuite) getOppositeCases(tests []opTestCase, oppositeOpMap map[string]string) []opTestCase {
	cases := make([]opTestCase, 0, len(tests))
	for _, c := range tests {
		op, ok := oppositeOpMap[c.input[1].(string)]
		if !ok {
			continue
		}
		c.input[1] = op
		c.expected = !c.expected
		cases = append(cases, c)
	}
	return cases
}

func (s *OperationTestSuite) testCases(tests []opTestCase) {
	for _, c := range tests {
		op := _operationFactory.Get(c.input[1].(string))
		require.NotNil(s.T(), op)
		v := _variableFactory.Create(c.input[0].(string))
		require.NotNil(s.T(), v)
		pv, err := op.PrepareValue(c.input[2])
		if c.expectedError {
			s.Error(err)
			continue
		}
		require.NoError(s.T(), err)
		s.Equal(c.expected, op.Run(s.ctx, v, pv), c.input)
	}
}

func TestOperationTestSuite(t *testing.T) {
	suite.Run(t, new(OperationTestSuite))
}
