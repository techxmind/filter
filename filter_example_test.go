package filter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/techxmind/filter"
	"github.com/techxmind/filter/core"
)

func ExampleFilterTrace() {
	// your business context
	ctx := context.Background()

	filterJson := `
	[
		[
			["ctx.foo", "in", "a,b,c"],
			["ctx.bar", ">", 10],
			["result-1", "=", "result-1"]
		],
		[
			["any?", "=>", [
				["ctx.foo", "in", "a,b,c"],
				["ctx.bar", ">", 10]
			]],
			["result-2", "=", "result-2"]
		]
	]
	`

	fctx := core.WithContext(ctx, core.WithTrace(core.NewTrace(os.Stderr)))
	fctx.Set("bar", 11)

	var filterData []interface{}
	if err := json.Unmarshal([]byte(filterJson), &filterData); err == nil {
		if f, err := filter.New(filterData); err == nil {
			data := make(map[string]interface{})
			f.Run(fctx, data)
			fmt.Printf("%v", data)
			// Output: map[result-2:result-2]
		}
	}
}

func ExampleDoc() {

	dataStr := `
	{
		"banner" : {
			"type" : "image",
			"src" : "https://example.com/working-hard.png",
			"link" : "https://example.com/activity.html"
		},

		"filter" : [
			["time", "between", "18:00,23:00"],
			["ctx.user.group", "=", "programer"],
			["banner", "+", {
				"src" : "https://example.com/chat-with-beaty.png",
				"link" : "https://chat.com"
			}]
		]
	}
	`

	var data map[string]interface{}
	json.Unmarshal([]byte(dataStr), &data)

	// your business context
	ctx := context.Background()

	filterCtx := core.WithContext(ctx)

	// you can also set context value in your business ctx with context.WithValue("group", ...)
	filterCtx.Set("user", map[string]interface{}{"group": "programer"})

	f, _ := filter.New(data["filter"].([]interface{}))
	f.Run(filterCtx, data)

	fmt.Println(data["banner"])
}
