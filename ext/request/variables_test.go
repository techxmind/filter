package request

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/techxmind/filter/core"
)

func TestAll(t *testing.T) {
	url := `http://www.techxmind.com/config?a=1&b=1,2,3&c={"d":[{"e":{"f":4}}, 5]}`
	ua := "test"
	ip := "8.8.8.8"

	c := context.WithValue(
		context.Background(),
		REQUEST_URL,
		url,
	)
	c = context.WithValue(
		c,
		USER_AGENT,
		ua,
	)
	c = context.WithValue(
		c,
		CLIENT_IP,
		ip,
	)
	ctx := core.WithContext(c)
	f := core.GetVariableFactory()

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"url", url},
		{"request-url", url},
		{"ua", ua},
		{"user-agent", ua},
		{"ip", ip},
		{"client-ip", ip},
		{"get.a", "1"},
		{"get.b", "1,2,3"},
		{"get.b[0]", "1"},
		{"get.c{d.0.e.f}", 4},
	}

	for i, c := range tests {
		v := f.Create(c.input)
		require.NotNil(t, v)
		assert.EqualValues(t, c.expected, core.GetVariableValue(ctx, v), "case %d: %s = %v", i, c.input, c.expected)
	}
}
