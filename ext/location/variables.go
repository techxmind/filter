// Define location variables from client ip
// So you should import package that defines ip var, e.g. "github.com/techxmind/vars/request"
// Variables:
//   country, province, city
package location

import (
	"github.com/techxmind/filter/core"
	"github.com/techxmind/ip2location"
)

func init() {
	f := core.GetVariableFactory()

	// variabe: country
	f.Register(
		core.SingletonVariableCreator(core.NewSimpleVariable("country", core.Cacheable, &VariableLocation{"country"})),
		"country",
	)

	// variabe: province
	f.Register(
		core.SingletonVariableCreator(core.NewSimpleVariable("province", core.Cacheable, &VariableLocation{"province"})),
		"province",
	)

	// variabe: city
	f.Register(
		core.SingletonVariableCreator(core.NewSimpleVariable("city", core.Cacheable, &VariableLocation{"city"})),
		"city",
	)
}

var (
	// mock it in unit test
	_getLocation = ip2location.Get
	_getIpVar    = func() core.Variable {
		return core.GetVariableFactory().Create("ip")
	}
)

type VariableLocation struct {
	name string
}

func (v *VariableLocation) Cacheable() bool { return true }
func (v *VariableLocation) Name() string    { return v.name }
func (v *VariableLocation) Value(ctx *core.Context) interface{} {
	ipVar := _getIpVar()
	if ipVar == nil {
		return nil
	}
	ip, ok := core.GetVariableValue(ctx, ipVar).(string)
	if !ok {
		return nil
	}
	loc, err := _getLocation(ip)
	if err != nil {
		return nil
	}

	if v.name == "country" {
		return loc.Country
	} else if v.name == "province" {
		return loc.Province
	} else {
		return loc.City
	}
}
