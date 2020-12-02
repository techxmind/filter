package core

import (
	"math"
	"math/rand"

	"github.com/pkg/errors"

	"github.com/techxmind/go-utils/itype"
)

func init() {
	_assignmentFactory.Register(&GroupAssign{}, "=>")
	_assignmentFactory.Register(&ProbabilitySet{}, "*=")
}

// ProbabilitySet set value with specified probability.
// e.g. :
//  ["key", "*=", [ [10, "value1"], [10, "value2"], [10, "value3"] ]]
//  value1,value2,value3 has the same weight 10, each of them has a probability 10/(10+10+10) = 1/3 to been chosen.
//
type ProbabilitySet struct {
}

type probabilityItem struct {
	linePoint int64
	value     interface{}
}

func (a *ProbabilitySet) PrepareValue(value interface{}) (val interface{}, err error) {
	if !IsArray(value) {
		err = errors.Errorf("assignment[*=] value must be array.")
		return
	}

	var (
		linePoint int64 = 0
		setter          = _assignmentFactory.Get("=")
		items           = make([]*probabilityItem, 0)
	)

	for _, item := range ToArray(value) {
		if !IsArray(item) {
			err = errors.Errorf("assignment[*=] value element must be array [#weight, #value].")
			return
		}

		pitem := ToArray(item)

		if len(pitem) != 2 {
			err = errors.Errorf("assignment[*=] value element must be array [#weight, #value].")
			return
		}

		if itype.GetType(pitem[0]) != itype.NUMBER {
			err = errors.Errorf("assignment[*=] value element 1st item(weight) must be number and gt 0")
		}

		weight := int64(math.Round(itype.Float(pitem[0]) * 1000))
		if weight < 0 {
			err = errors.Errorf("assignment[*=] value element 1st item(weight) must be number and gt 0")
			return
		}

		linePoint += weight

		pvalue, err := setter.PrepareValue(pitem[1])
		if err != nil {
			return nil, err
		}

		items = append(items, &probabilityItem{
			linePoint: linePoint,
			value:     pvalue,
		})
	}

	return items, nil
}

func (a *ProbabilitySet) Run(ctx *Context, data interface{}, key string, value interface{}) {
	items, ok := value.([]*probabilityItem)
	if !ok {
		return
	}
	n := len(items)
	if n == 0 {
		return
	}
	max := items[n-1].linePoint
	choose := rand.Intn(int(max)) + 1
	for _, item := range items {
		if choose <= int(item.linePoint) {
			_assignmentFactory.Get("=").Run(ctx, data, key, item.value)
			break
		}
	}
}

type GroupAssign struct{}

func (a *GroupAssign) PrepareValue(value interface{}) (val interface{}, err error) {
	if !IsArray(value) {
		return nil, errors.New("assignment[=>] value must be array")
	}

	executor, err := NewExecutor(ToArray(value))
	if err != nil {
		return nil, errors.Wrap(err, "assignment[=>]")
	}

	return executor, nil
}

func (a *GroupAssign) Run(ctx *Context, data interface{}, key string, value interface{}) {
	executor, ok := value.(Executor)

	if !ok {
		return
	}

	executor.Execute(ctx, data)
}
