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

    // if current time is between 18:00 ~ 23:00, outputs:
    // map[link:https://chat.com src:https://example.com/chat-with-beaty.png type:image]
	fmt.Println(data["banner"])
}
```
