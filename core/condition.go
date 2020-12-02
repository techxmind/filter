package core

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

//Condition interface
type Condition interface {
	Success(ctx *Context) bool
	String() string
}

//StdCondition
type StdCondition struct {
	expr      string
	variable  Variable
	operation Operation
	value     interface{}
}

func (c *StdCondition) Success(ctx *Context) bool {
	ok := c.operation.Run(ctx, c.variable, c.value)

	if trace := ctx.Trace(); trace != nil {
		trace.Log(
			c.String(),
			" => ",
			GetVariableValue(ctx, c.variable),
			c.operation.String(),
			c.value,
			" => ",
			ok,
		)
	}

	return ok
}

func (c *StdCondition) String() string {
	return c.expr
}

//ConditionGroup
type GROUP_LOGIC int

const (
	LOGIC_ALL GROUP_LOGIC = iota
	LOGIC_ANY
	LOGIC_NONE
	LOGIC_ANY_NOT
)

type ConditionGroup struct {
	logic      GROUP_LOGIC
	conditions []Condition
}

func (c *ConditionGroup) Success(ctx *Context) bool {
	result := c.logic != LOGIC_ANY_NOT
	for _, condition := range c.conditions {
		if ok := condition.Success(ctx); ok {
			if c.logic == LOGIC_ANY {
				result = true
				break
			}
			if c.logic == LOGIC_NONE {
				result = false
				break
			}
		} else {
			if c.logic == LOGIC_ALL {
				result = false
				break
			} else if c.logic == LOGIC_ANY_NOT {
				result = true
				break
			} else if c.logic == LOGIC_ANY {
				result = false
			}
		}
	}

	return result
}

func (c *ConditionGroup) String() string {
	var b strings.Builder
	for k, v := range groupConditionKeys {
		if v == c.logic {
			b.WriteString(strings.ToUpper(k))
			break
		}
	}
	b.WriteString("{")
	for _, c := range c.conditions {
		b.WriteString(c.String())
		//b.WriteString(", ")
	}
	b.WriteString("}")

	return b.String()
}

func (c *ConditionGroup) add(condition Condition) {
	c.conditions = append(c.conditions, condition)
}

func NewConditionGroup(logic GROUP_LOGIC) *ConditionGroup {
	return &ConditionGroup{
		logic:      logic,
		conditions: make([]Condition, 0),
	}
}

var groupConditionKeys = map[string]GROUP_LOGIC{
	"any?":  LOGIC_ANY,
	"not?":  LOGIC_ANY_NOT,
	"all?":  LOGIC_ALL,
	"none?": LOGIC_NONE,
}

func NewCondition(item []interface{}, groupLogic GROUP_LOGIC) (Condition, error) {
	if len(item) == 0 {
		return nil, errors.New("Empty")
	}

	if IsArray(item[0]) {
		group := NewConditionGroup(groupLogic)

		for _, subitem := range item {
			if !IsArray(subitem) {
				return nil, errors.Errorf("Sub item must be an array. -> %s", jstr(subitem))
			}
			if subCondition, err := NewCondition(ToArray(subitem), LOGIC_ALL); err != nil {
				return nil, err
			} else {
				group.add(subCondition)
			}
		}

		return group, nil
	}

	if len(item) != 3 {
		return nil, errors.Errorf("Item must contains 3 elements. -> %s", jstr(item))
	}

	key, ok := item[0].(string)
	if !ok {
		return nil, errors.Errorf("Item 1st element[%g] is not string. -> %s", item[0], jstr(item))
	}

	if logic, ok := groupConditionKeys[key]; ok {
		list, ok := item[2].([]interface{})
		if !ok {
			return nil, errors.Errorf("Logic[%s] item 3rd element must be an array. -> %s", key, jstr(item))
		}
		if group, err := NewCondition(list, logic); err != nil {
			return nil, err
		} else {
			return group, nil
		}
	}

	variable := _variableFactory.Create(key)
	if variable == nil {
		return nil, errors.Errorf("Unknown var[%s]. -> %s", key, jstr(item))
	}

	operationName, ok := item[1].(string)
	if !ok {
		return nil, errors.Errorf("Item 2nd element(operation) is not string. -> %s", jstr(item))
	}

	operation := _operationFactory.Get(operationName)

	if operation == nil {
		return nil, errors.Errorf("Unknown operation[%s]. -> %s", operationName, jstr(item))
	}

	pvalue, err := operation.PrepareValue(item[2])

	if err != nil {
		return nil, err
	}

	expr := fmt.Sprintf("%s %s %s", key, operationName, jstr(pvalue))

	condition := &StdCondition{
		expr:      expr,
		variable:  variable,
		operation: operation,
		value:     pvalue,
	}

	return condition, nil
}
