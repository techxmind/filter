# filter
Simple and extensible data filtering framework. You can simply create filters from json/yaml data, filter definition can immbeded in business data. e.g.

*DATA*
```
{
    "banner" : {
        "type" : "image",
        "src" : "https://example.com/working-hard.png",
        "link" : "https://example.com/activity.html"
    },
    //... other config

    "filter" : [
        ["time", "between", "18:00,23:00"],
        ["ctx.user.group", "=", "programer"],
        ["banner", "+", {
            "src" : "https://example.com/chat-with-beaty.png",
            "link" : "https://chat.com"
        }]
    ]
}
```

*CODE*
```
import (
    "encoding/json"

	"github.com/techxmind/filter"
	"github.com/techxmind/filter/core"
)

func main() {
    var dataStr = `...` // data above

	var data map[string]interface{}

	json.Unmarshal([]byte(dataStr), &data)

	// your business context
	ctx := context.Background()

	filterCtx := core.WithContext(ctx)

	// you can also set context value in your business ctx with context.WithValue("group", ...)
	filterCtx.Set("user", map[string]interface{}{"group": "programer"})

	f, _ := filter.New(data["filter"].([]interface{}))
	f.Run(filterCtx, data)

    // if current time is between 18:00 ~ 23:00, output:
    // map[link:https://chat.com src:https://example.com/chat-with-beaty.png type:image]
	fmt.Println(data["banner"])
}
```

## Variables

Register your custom variable:

```
import (
	"github.com/techxmind/filter/core"
)

// Register variable "username" that fetch value from context
core.GetVariableFactory().Register(
    core.SingletonVariableCreator(core.NewSimpleVariable("username", core.Cacheable, &ContextValue{USER_AGENT})),
    "username"
)
```

Check `ext` folder to see more examples.

## Operations

Register your custom operation:

```
import (
    "errors"
    "strings"

	"github.com/techxmind/filter/core"
	"github.com/techxmind/go-utils/itype"
)

// Register operation "contains" to check if variable contains specified substring
// ["url", "contains", "something"]
core.GetOperationFactory().Register(&ContainsOperation{}, "contains")

type ContainsOperation struct {}

func (o \*ContainsOperation) String() string {
    return "contains"
}

func (o \*ContainsOperation) Run(ctx \*core.Context, variable core.Variable, value interface{}) bool {
	v := core.GetVariableValue(ctx, variable)

    return strings.Contains(itype.String(v), itype.String(value))
}

func (o \*ContainsOperation) PrepareValue(v interface{}) (interface{}, error) {
    if str, ok := v.(string); ok {
        return v, nil
    }

    return nil, errors.New("[contains] operation require value of type string")
}
```

## Assignments
...
