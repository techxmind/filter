package core

import (
	"math/rand"
	"strings"
	"time"

	"github.com/techxmind/go-utils/object"
)

// register core varaiables
//   succ     : bool, true
//   rand     : int, random value in range 1 ~ 100
//   datetime : string, current date time with format 2006-01-02 15:04: 05
//   date     : string, current date with format 2006-01-02
//   time     : string, current time with format 15:04:05
//   year     : int, current year, e.g. 2020
//   month    : int, current month in range 1 ~ 12
//   day      : int, current day in range 1 ~ 31
//   hour     : int, current hour in range 0 ~ 23
//   minute   : int, current minute in range 0 ~ 59
//   second   : int, current second in range 0 ~ 59
//   unixtime : int, number of seconds since the Epoch
//   wday     : int, the day of the week, range 1 ~ 7, Monday = 1 ...
//   data.xx  : mixed, xx is key path to the value in data being filtered. e.g. data.foo.bar means data['foo']['bar']
//   ctx.xx   : mixed, like data.xx. Search value in data["ctx"] or context
//              e.g. ctx.foo.bar search in order:
//               1. check data["ctx"]["foo"]["bar"]
//               2. check context data setted by ctx.Set("foo", fooValue), check fooValue["bar"]
//               3. check context data setted by context.WithValue("foo", fooValue); check fooValue["bar"]
//
func init() {
	// variable: succ
	_variableFactory.Register(
		SingletonVariableCreator(NewSimpleVariable("succ", Cacheable, &StaticValue{true})),
		"succ",
	)

	rand.Seed(time.Now().UnixNano())
	// variable: rand
	_variableFactory.Register(
		SingletonVariableCreator(
			NewSimpleVariable("rand", Uncacheable, ValueFunc(func(ctx *Context) interface{} {
				return rand.Intn(100) + 1
			})),
		),
		"rand",
	)

	// variable: time group...
	names := []string{
		"datetime", "date", "time", "year", "month", "day",
		"hour", "minute", "second", "unixtime", "wday",
	}
	for _, name := range names {
		_variableFactory.Register(SingletonVariableCreator(&variableTime{name}), name)
	}

	// variable: data.xx
	_variableFactory.Register(VariableCreatorFunc(variableDataCreator), "data.")

	// variable: ctx.xx
	_variableFactory.Register(VariableCreatorFunc(variableCtxCreator), "ctx.")
}

var (
	// Mock it for testing
	_currentTime = time.Now
)

// variable: time
type variableTime struct {
	name string
}

func (v *variableTime) Cacheable() bool { return false }
func (v *variableTime) Name() string    { return v.name }
func (v *variableTime) Value(ctx *Context) interface{} {
	now := _currentTime()

	switch v.name {
	case "unixtime":
		return now.Unix()
	case "hour":
		return now.Hour()
	case "minute":
		return now.Minute()
	case "second":
		return now.Second()
	case "year":
		return now.Year()
	case "month":
		return int(now.Month())
	case "day":
		return now.Day()
	case "wday":
		wday := int(now.Weekday())
		// Set Sunday to 7
		if wday == 0 {
			wday = 7
		}
		return wday
	case "date":
		return now.Format("2006-01-02")
	case "time":
		return now.Format("15:04:05")
	case "datetime":
		fallthrough
	default:
		return now.Format("2006-01-02 15:04:05")
	}
}

// variableData access the data being filtered
type variableData struct {
	name string
	key  string
}

func (self *variableData) Cacheable() bool { return false }
func (self *variableData) Name() string    { return self.name }
func (self *variableData) Value(ctx *Context) interface{} {
	if v, ok := object.GetValue(ctx.Data(), self.key); ok {
		return v
	}

	return nil
}

func variableDataCreator(name string) Variable {
	key := strings.TrimPrefix(name, "data.")

	if key == "" {
		return nil
	}

	return &variableData{
		name: name,
		key:  key,
	}
}

// variableCtx access context value
type variableCtx struct {
	name string
	key  string
}

func (self *variableCtx) Cacheable() bool { return false }

func (self *variableCtx) Name() string {
	return self.name
}

// Context variable search value with order:
//   1. check data["ctx"] data map
//   2. check context data, both setted by WithValue or Set method
//
func (self *variableCtx) Value(ctx *Context) interface{} {

	// First priority: data["ctx"][key...]
	if value, ok := object.GetValue(ctx.Data(), "ctx."+self.key); ok {
		return value
	}

	// Secondary priority: from Context.Set(topKey, value)
	if value, ok := object.GetValue(ctx.GetAll(), self.key); ok {
		return value
	}

	paths := strings.Split(self.key, ".")

	// Default from Context.WithValue(topKey, value)
	if v := ctx.Value(paths[0]); v != nil {
		if len(paths) == 1 {
			return v
		}

		v, _ := object.GetValue(v, strings.Join(paths[1:], "."))

		return v
	}

	return nil
}

func variableCtxCreator(name string) Variable {
	key := strings.TrimPrefix(name, "ctx.")
	if key == "" {
		return nil
	}
	return &variableCtx{
		name: name,
		key:  key,
	}
}
