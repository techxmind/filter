package request

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"

	"github.com/techxmind/filter/core"
	"github.com/techxmind/go-utils/itype"
	"github.com/techxmind/go-utils/object"
)

const (
	REQUEST_URL = "request-url"
	USER_AGENT  = "user-agent"
	CLIENT_IP   = "client-ip"
)

func init() {
	f := core.GetVariableFactory()

	// variable: url
	// request url from context
	f.Register(
		core.SingletonVariableCreator(&VariableURL{REQUEST_URL}),
		"url", "request-url",
	)

	// variabe: ua
	// user-agent from context
	f.Register(
		core.SingletonVariableCreator(core.NewSimpleVariable("ua", core.Cacheable, &ContextValue{USER_AGENT})),
		"ua", "user-agent",
	)

	// variabe: ip
	// user-agent from context
	f.Register(
		core.SingletonVariableCreator(core.NewSimpleVariable("ip", core.Cacheable, &ContextValue{CLIENT_IP})),
		"ip", "client-ip",
	)

	// variable: get.xx
	// value from request-url query parameters
	f.Register(
		core.VariableCreatorFunc(VariableGetCreator),
		"get.", "query.",
	)
}

// ContextValue implements Valuer, get value from context
type ContextValue struct {
	name interface{}
}

func (v *ContextValue) Value(ctx *core.Context) interface{} {
	return ctx.Value(v.name)
}

// VariableURL
type VariableURL struct {
	name interface{}
}

func (v *VariableURL) Cacheable() bool { return true }
func (v *VariableURL) Name() string    { return itype.String(v.name) }
func (v *VariableURL) Value(ctx *core.Context) interface{} {
	return ctx.Value(v.name)
}
func (v *VariableURL) Query(ctx *core.Context) url.Values {
	us, ok := v.Value(ctx).(string)

	if !ok || us == "" {
		return nil
	}

	cache := ctx.Cache()
	cacheID := core.CacheID("values:" + us)
	val, ok := cache.Load(cacheID)
	if ok {
		if values, ok := val.(url.Values); ok {
			return values
		}
	}

	u, err := url.Parse(us)
	if err != nil {
		core.Logger.Printf("Parse url err:%s %v\n", us, err)
		return nil
	}

	values := u.Query()
	cache.Store(cacheID, values)

	return values
}

// VariableQueryStr get value from url query
type VariableQueryStr struct {
	name             string
	paramName        string
	queryValueGetter func(*core.Context, string) string
	listMode         bool
	listIndex        int
	jsonMode         bool
	jsonKey          string
}

func (self *VariableQueryStr) Cacheable() bool { return true }
func (self *VariableQueryStr) Name() string    { return self.name }
func (self *VariableQueryStr) Value(ctx *core.Context) interface{} {
	value := self.queryValueGetter(ctx, self.paramName)

	if value == "" || (!self.listMode && !self.jsonMode) {
		return value
	}

	var (
		ivalue interface{} = value
		cache              = ctx.Cache()
	)

	if self.jsonMode {
		var data interface{}
		if cacheData, ok := cache.Load("json." + self.name); ok {
			data = cacheData
		} else {
			if err := json.Unmarshal([]byte(value), &data); err != nil {
				core.Logger.Printf("json.Unmarshal url query variable[%s] err. value=%s err=%v\n", self.paramName, value, err)
				return ""
			}
		}
		if data == nil {
			core.Logger.Printf("Query variable[%s] json.Unmarshal get nil. value=%s", self.paramName, value)
			return nil
		}
		if v, ok := object.GetValue(data, self.jsonKey); ok {
			ivalue = v
		} else {
			return nil
		}
	}

	if self.listMode {
		arr := core.ToArray(ivalue)
		if self.listIndex < 0 || self.listIndex >= len(arr) {
			return nil
		}
		return arr[self.listIndex]
	}

	return ivalue
}

var _getRegexp = regexp.MustCompile("^(?:get|query).(.+?)(?:\\{([^\\}]+)\\})?(?:\\[(\\d+)\\])?$")

func queryValueGetter(ctx *core.Context, name string) string {
	urlVar := core.GetVariableFactory().Create("url")
	if urlVar == nil {
		return ""
	}

	url, ok := urlVar.(*VariableURL)
	if !ok {
		return ""
	}

	values := url.Query(ctx)

	if values == nil {
		return ""
	}

	return values.Get(name)
}

func VariableGetCreator(name string) core.Variable {
	if ma := _getRegexp.FindStringSubmatch(name); len(ma) == 4 {
		obj := &VariableQueryStr{
			name:             ma[0],
			paramName:        ma[1],
			queryValueGetter: queryValueGetter,
			listMode:         false,
			listIndex:        0,
			jsonMode:         false,
			jsonKey:          "",
		}
		if ma[2] != "" {
			obj.jsonMode = true
			obj.jsonKey = ma[2]
		}
		if ma[3] != "" {
			obj.listMode = true
			obj.listIndex, _ = strconv.Atoi(ma[3])
		}

		return obj
	}

	return nil
}
