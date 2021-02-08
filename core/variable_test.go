package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariable(t *testing.T) {
	ctx := NewContext()
	fooVar := NewSimpleVariable("foo", Cacheable, StaticValue{"foo"})
	assert.NotNil(t, fooVar)
	assert.Equal(t, "foo", fooVar.Name())
	assert.Equal(t, Cacheable, fooVar.Cacheable())
	assert.Equal(t, "foo", fooVar.Value(ctx))

	f := GetVariableFactory()
	f.Register(SingletonVariableCreator(fooVar), "foo", "my-foo")
	v1 := f.Create("foo")
	assert.Equal(t, fooVar, v1)
	v2 := f.Create("my-foo")
	assert.Equal(t, fooVar, v2)
	v3 := f.Create("your-foo")
	assert.Nil(t, v3)
}

func TestRegisterAlias(t *testing.T) {
	ctx := NewContext()
	f := GetVariableFactory()

	f.RegisterAlias("succ", "s")
	v := f.Create("s")
	assert.NotNil(t, v)
	assert.Equal(t, true, GetVariableValue(ctx, v))

	f.RegisterAlias("ctx.foo", "foo", "f")
	f.RegisterAlias("ctx.bar.", "bar.")

	ctx.Set("foo", "fv")
	ctx.Set("bar", map[string]interface{}{"baz": "bv"})
	v = f.Create("foo")
	assert.NotNil(t, v)
	assert.EqualValues(t, "fv", GetVariableValue(ctx, v))

	v = f.Create("f")
	assert.NotNil(t, v)
	assert.EqualValues(t, "fv", GetVariableValue(ctx, v))

	v = f.Create("bar.baz")
	assert.NotNil(t, v)
	assert.EqualValues(t, "bv", GetVariableValue(ctx, v))

	v = f.Create("bar")
	assert.Nil(t, v)
}

func TestGetVariableValue(t *testing.T) {
	ctx := NewContext()
	var1Val := &StaticValue{"var1"}
	var1 := NewSimpleVariable("var1", Uncacheable, var1Val)
	assert.Equal(t, "var1", GetVariableValue(ctx, var1))
	var1Val.Val = "var1-changed"
	assert.Equal(t, "var1-changed", GetVariableValue(ctx, var1))

	var2Val := &StaticValue{"var2"}
	var2 := NewSimpleVariable("var2", Cacheable, var2Val)
	assert.Equal(t, "var2", GetVariableValue(ctx, var2))
	var2Val.Val = "var2-channged"
	assert.Equal(t, "var2", GetVariableValue(ctx, var2))
}
