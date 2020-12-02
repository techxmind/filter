package core

import (
	"fmt"

	"github.com/pkg/errors"
)

//Executor
type Executor interface {
	Execute(*Context, interface{})
}

//StdExecutor
type StdExecutor struct {
	expr       string
	key        string
	assignment Assignment
	value      interface{}
}

func (e *StdExecutor) Execute(ctx *Context, data interface{}) {
	if trace := ctx.Trace(); trace != nil {
		trace.Log(
			e.expr,
		)
	}

	e.assignment.Run(ctx, data, e.key, e.value)
}

//ExecutorGroup
type ExecutorGroup struct {
	executors []Executor
}

func (e *ExecutorGroup) Execute(ctx *Context, data interface{}) {
	for _, executor := range e.executors {
		executor.Execute(ctx, data)
	}
}

func (e *ExecutorGroup) add(executor Executor) {
	e.executors = append(e.executors, executor)
}

func NewExecutorGroup() *ExecutorGroup {
	return &ExecutorGroup{
		executors: make([]Executor, 0),
	}
}

func NewExecutor(item []interface{}) (Executor, error) {
	if len(item) == 0 {
		return nil, errors.New("Executor is empty")
	}

	if IsArray(item[0]) {
		group := NewExecutorGroup()

		for _, subitem := range item {
			if !IsArray(subitem) {
				return nil, errors.Errorf("Executor child item is not array. -> %s", jstr(item))
			}
			if subExecutor, err := NewExecutor(ToArray(subitem)); err != nil {
				return nil, err
			} else {
				group.add(subExecutor)
			}
		}

		return group, nil
	}

	if len(item) != 3 {
		return nil, errors.Errorf("Executor item must contains 3 elements. -> %s", jstr(item))
	}

	key, ok := item[0].(string)
	if !ok {
		return nil, errors.Errorf("Executor item 1st element (%v) is not string. -> %s", item[0], jstr(item))
	}

	assignmentName, ok := item[1].(string)
	if !ok {
		return nil, errors.Errorf("Executor item 2nd element (%v) is not string. -> %s", item[1], jstr(item))
	}

	assignment := _assignmentFactory.Get(assignmentName)
	if assignment == nil {
		return nil, errors.Errorf("Executor with invalid assignment[%s]", assignmentName)
	}

	value, err := assignment.PrepareValue(item[2])
	if err != nil {
		return nil, errors.Errorf("Executor assignment[%s] prepare value err:%s", assignmentName, err)
	}

	expr := fmt.Sprintf("%s %s %s", key, assignmentName, jstr(value))

	executor := &StdExecutor{
		expr:       expr,
		key:        key,
		assignment: assignment,
		value:      value,
	}

	return executor, nil
}
