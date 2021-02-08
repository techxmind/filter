package core

import (
	"strings"
)

var (
	Cacheable   = true
	Uncacheable = false

	_variableFactory = &stdVariableFactory{
		creators: make(map[string]VariableCreator),
		aliases:  make(map[string]string),
	}
)

type Variable interface {
	Cacheable() bool
	Name() string
	Valuer
}

type Valuer interface {
	Value(*Context) interface{}
}

type VariableCreator interface {
	Create(string) Variable
}

type VariableCreatorFunc func(string) Variable

func (f VariableCreatorFunc) Create(name string) Variable {
	return f(name)
}

type VariableFactory interface {
	VariableCreator
	Register(VariableCreator, ...string)
	// RegisterAlias register aliases for destination variable
	// RegisterAlias("time", "t") t => t
	// RegisterAlias("ctx.something", "s") s.l => ctx.something.l
	RegisterAlias(dest string, aliases ...string) error
}

// GetVariableFactory return VariableFactory
func GetVariableFactory() VariableFactory {
	return _variableFactory
}

// stdVariableFactory default VariableFactory
type stdVariableFactory struct {
	creators map[string]VariableCreator
	aliases  map[string]string
}

func (f *stdVariableFactory) Register(creator VariableCreator, names ...string) {
	for _, name := range names {
		f.creators[name] = creator
	}
}

func (f *stdVariableFactory) RegisterAlias(dest string, aliases ...string) error {
	for _, alias := range aliases {
		f.aliases[alias] = dest
	}
	return nil
}

func (f *stdVariableFactory) Create(name string) Variable {
	segments := strings.Split(name, ".")
	if v, ok := f.aliases[segments[0]]; ok {
		n := strings.Split(v, ".")
		segments = append(n, segments[1:]...)
	} else if v, ok := f.aliases[segments[0]+"."]; ok && len(segments) > 1 {
		n := strings.Split(strings.TrimRight(v, "."), ".")
		segments = append(n, segments[1:]...)
	}
	if len(segments) == 1 {
		if creator, ok := f.creators[segments[0]]; ok {
			return creator.Create(segments[0])
		}
	} else {
		if creator, ok := f.creators[segments[0]+"."]; ok {
			return creator.Create(strings.Join(segments, "."))
		}
	}

	return nil
}

// GetVariableValue get value of variable, also handlers variable cacheing, hooking logic
//
func GetVariableValue(ctx *Context, v Variable) interface{} {
	if v == nil {
		return ""
	}

	if v.Cacheable() {
		if value, ok := ctx.Cache().Load(v.Name()); ok {
			return value
		}
	}

	value := v.Value(ctx)

	if v.Cacheable() {
		ctx.Cache().Store(v.Name(), value)
	}

	return value
}

// ValueFunc implements Valuer interface
type ValueFunc func(*Context) interface{}

func (f ValueFunc) Value(ctx *Context) interface{} {
	return f(ctx)
}

// StaticValue static value implements Valuer interface
type StaticValue struct {
	Val interface{}
}

func (v StaticValue) Value(_ *Context) interface{} {
	return v.Val
}

// SimpleVariable implements Variable interface
type SimpleVariable struct {
	cacheable bool
	name      string
	value     Valuer
}

func NewSimpleVariable(name string, cacheable bool, value Valuer) Variable {
	return &SimpleVariable{
		name:      name,
		cacheable: cacheable,
		value:     value,
	}
}

func (v *SimpleVariable) Name() string {
	return v.name
}

func (v *SimpleVariable) Cacheable() bool {
	return v.cacheable
}

func (v *SimpleVariable) Value(ctx *Context) interface{} {
	return v.value.Value(ctx)
}

func SingletonVariableCreator(instance Variable) VariableCreatorFunc {
	return func(name string) Variable {
		return instance
	}
}
