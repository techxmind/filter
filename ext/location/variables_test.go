package location

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/techxmind/filter/core"
	"github.com/techxmind/ip2location"
)

func TestLocation(t *testing.T) {
	f := core.GetVariableFactory()
	originalGetLocation := _getLocation
	_getLocation = func(_ string) (*ip2location.Location, error) {
		return &ip2location.Location{
			Country:  "中国",
			Province: "江苏省",
			City:     "南京市",
		}, nil
	}
	originalGetIpVar := _getIpVar
	_getIpVar = func() core.Variable {
		return core.NewSimpleVariable("ip", core.Cacheable, &core.StaticValue{"8.8.8.8"})
	}
	defer func() {
		_getLocation = originalGetLocation
		_getIpVar = originalGetIpVar
	}()

	ctx := core.NewContext()
	v := f.Create("country")
	require.NotNil(t, v)
	assert.Equal(t, "中国", core.GetVariableValue(ctx, v))

	v = f.Create("province")
	require.NotNil(t, v)
	assert.Equal(t, "江苏省", core.GetVariableValue(ctx, v))

	v = f.Create("city")
	require.NotNil(t, v)
	assert.Equal(t, "南京市", core.GetVariableValue(ctx, v))
}
