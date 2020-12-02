package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/techxmind/go-utils/compare"
	"github.com/techxmind/go-utils/itype"
)

var _operationFactory OperationFactory

func init() {
	_operationFactory = &stdOperationFactory{
		operations: map[string]Operation{
			"=":       &EqualOperation{stringer: stringer("=")},
			"!=":      &NotEqualOperation{stringer: stringer("!=")},
			">":       &GtOperation{stringer: stringer(">")},
			">=":      &GeOperation{stringer: stringer(">=")},
			"<":       &LtOperation{stringer: stringer("<")},
			"<=":      &LeOperation{stringer: stringer("<=")},
			"between": &BetweenOperation{stringer: stringer("between")},
			"in":      &InOperation{stringer: stringer("in")},
			"not in":  &NotInOperation{stringer: stringer("not in")},
			"~":       &MatchOperation{stringer: stringer("~")},
			"!~":      &NotMatchOperation{stringer: stringer("!~")},
			"any":     &AnyOperation{stringer: stringer("any")},
			"has":     &HasOperation{stringer: stringer("has")},
			"none":    &NoneOperation{stringer: stringer("none")},
		},
	}
}

type Operation interface {
	Run(ctx *Context, variable Variable, value interface{}) bool
	PrepareValue(value interface{}) (interface{}, error)
	String() string
}

type OperationFactory interface {
	Get(string) Operation
	Register(Operation, ...string)
}

func GetOperationFactory() OperationFactory {
	return _operationFactory
}

// stdOperationFactory is default OperationFactory
type stdOperationFactory struct {
	operations map[string]Operation
}

func (f *stdOperationFactory) Get(name string) Operation {
	if value, ok := f.operations[name]; ok {
		return value
	}

	return nil
}

func (f *stdOperationFactory) Register(op Operation, names ...string) {
	for _, name := range names {
		f.operations[name] = op
	}
}

//----------------------------------------------------------------------------------
// helper functions
//----------------------------------------------------------------------------------

type operationBase struct{}

func (o operationBase) PrepareValue(value interface{}) (interface{}, error) {
	return value, nil
}

//----------------------------------------------------------------------------------
// core operations define
//----------------------------------------------------------------------------------

//----------------------------------------------------------------------------------
type EqualOperation struct {
	stringer
	operationBase
}

func (o *EqualOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	cmpValue := GetVariableValue(ctx, variable)

	if b, ok := value.(bool); ok {
		return itype.Bool(cmpValue) == b
	}

	return compare.Object(cmpValue, value) == 0
}

//----------------------------------------------------------------------------------
type NotEqualOperation struct {
	stringer
	EqualOperation
}

func (o *NotEqualOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	return !o.EqualOperation.Run(ctx, variable, value)
}

//----------------------------------------------------------------------------------
type GtOperation struct {
	stringer
	operationBase
}

func (o *GtOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	cmpValue := GetVariableValue(ctx, variable)
	return compare.Object(cmpValue, value) == 1
}

//----------------------------------------------------------------------------------
type LeOperation struct {
	stringer
	GtOperation
}

func (o *LeOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	return !o.GtOperation.Run(ctx, variable, value)
}

//----------------------------------------------------------------------------------
type LtOperation struct {
	stringer
	operationBase
}

func (o *LtOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	cmpValue := GetVariableValue(ctx, variable)

	return compare.Object(cmpValue, value) == -1
}

//----------------------------------------------------------------------------------
type GeOperation struct {
	stringer
	LtOperation
}

func (o *GeOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	return !o.LtOperation.Run(ctx, variable, value)
}

//----------------------------------------------------------------------------------
type BetweenOperation struct{ stringer }

func (o *BetweenOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	cmpValue := GetVariableValue(ctx, variable)
	startAndEnd := value.([]interface{})
	return compare.Object(cmpValue, startAndEnd[0]) >= 0 && compare.Object(cmpValue, startAndEnd[1]) <= 0
}
func (o *BetweenOperation) PrepareValue(value interface{}) (interface{}, error) {
	startAndEnd := ToArray(value)

	if len(startAndEnd) != 2 {
		return nil, errors.New(fmt.Sprintf("[between] operation value must be a list with 2 elements"))
	}

	return startAndEnd, nil
}

//----------------------------------------------------------------------------------
type InOperation struct{ stringer }

func (o *InOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	cmpValue := GetVariableValue(ctx, variable)

	for _, elem := range value.([]interface{}) {
		if compare.Object(elem, cmpValue) == 0 {
			return true
		}
	}

	return false
}

func (o *InOperation) PrepareValue(value interface{}) (interface{}, error) {
	elems := ToArray(value)

	if len(elems) == 0 {
		return nil, errors.New(fmt.Sprintf("[in/not in] operation value must be a list"))
	}

	return elems, nil
}

//----------------------------------------------------------------------------------
type NotInOperation struct {
	stringer
	InOperation
}

func (o *NotInOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	return !o.InOperation.Run(ctx, variable, value)
}

//----------------------------------------------------------------------------------
type MatchOperation struct{ stringer }

func (o *MatchOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	cmpValue := GetVariableValue(ctx, variable)

	cmpValueStr, ok := cmpValue.(string)

	if !ok {
		return false
	}

	if regexpObj, ok := value.(*regexp.Regexp); ok {

		return regexpObj.MatchString(cmpValueStr)
	} else if strObj, ok := value.(string); ok {

		return strings.Contains(strings.ToLower(cmpValueStr), strObj)
	}

	return false
}

func (o *MatchOperation) PrepareValue(value interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil, errors.New(fmt.Sprintf("[match] operation value must be a string"))
	}

	if !(strings.HasPrefix(str, "/") && strings.HasSuffix(str, "/")) {
		return strings.ToLower(str), nil
	}

	str = strings.Trim(str, "/")
	if str == "" {
		return nil, errors.New(fmt.Sprintf("[match] operation value is not a valid regexp expression[%s]", str))
	}

	if robj, err := regexp.Compile("(?i)" + str); err != nil {
		return nil, errors.New(fmt.Sprintf("[match] operation value is not a valid regexp expression[%s].err:%s", str, err))
	} else {
		return robj, nil
	}
}

//----------------------------------------------------------------------------------
type NotMatchOperation struct {
	stringer
	MatchOperation
}

func (o *NotMatchOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	return !o.MatchOperation.Run(ctx, variable, value)
}

//----------------------------------------------------------------------------------
type AnyOperation struct{ stringer }

func (o *AnyOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	cmpValue := GetVariableValue(ctx, variable)

	cmpElems := ToArray(cmpValue)
	elems := ToArray(value)

	if len(elems) == 0 || len(cmpElems) == 0 {
		return false
	}

	for _, elem := range elems {
		for _, cmpElem := range cmpElems {
			if compare.Object(elem, cmpElem) == 0 {
				return true
			}
		}
	}

	return false
}

func (o *AnyOperation) PrepareValue(value interface{}) (interface{}, error) {
	elems := ToArray(value)

	if len(elems) == 0 {
		return nil, errors.New(fmt.Sprintf("[any] operation value must be a list"))
	}

	return elems, nil
}

//----------------------------------------------------------------------------------
type HasOperation struct{ stringer }

func (o *HasOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	cmpValue := GetVariableValue(ctx, variable)

	cmpElems := ToArray(cmpValue)
	elems := ToArray(value)

	if len(elems) == 0 || len(cmpElems) == 0 {
		return false
	}

	for _, elem := range elems {
		ok := false
		for _, cmpElem := range cmpElems {
			if compare.Object(elem, cmpElem) == 0 {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}

	return true
}
func (o *HasOperation) PrepareValue(value interface{}) (interface{}, error) {
	elems := ToArray(value)

	if len(elems) == 0 {
		return nil, errors.New(fmt.Sprintf("[has] operation value must be a list"))
	}

	return elems, nil
}

//----------------------------------------------------------------------------------
type NoneOperation struct {
	stringer
	AnyOperation
}

func (o *NoneOperation) Run(ctx *Context, variable Variable, value interface{}) bool {
	return !o.AnyOperation.Run(ctx, variable, value)
}
