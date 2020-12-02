package core

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/mohae/deepcopy"
	"github.com/spaolacci/murmur3"
	"github.com/techxmind/go-utils/itype"
)

// ToArray convert value to []interface{}
func ToArray(value interface{}) []interface{} {
	var ret []interface{}

	if value == nil {
		return ret
	}

	if v, ok := value.([]interface{}); ok {
		return v
	}

	tp := itype.GetType(value)

	// split string with seperator ','
	if tp == itype.STRING {
		v := value.(string)
		if v == "" {
			return []interface{}{}
		}
		sarr := strings.Split(v, ",")
		ret = make([]interface{}, 0, len(sarr))
		for _, e := range sarr {
			e = strings.TrimSpace(e)
			if e != "" {
				ret = append(ret, e)
			}
		}
		return ret
	}

	if tp != itype.ARRAY {
		return []interface{}{value}
	}

	va := reflect.ValueOf(value)

	ret = make([]interface{}, va.Len())
	for i := 0; i < va.Len(); i++ {
		ret[i] = va.Index(i).Interface()
	}

	return ret
}

func jstr(v interface{}) string {
	var s strings.Builder
	encoder := json.NewEncoder(&s)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "")
	encoder.Encode(v)

	return strings.TrimSpace(s.String())
}

func IsArray(v interface{}) bool {
	return itype.GetType(v) == itype.ARRAY
}

func IsScalar(v interface{}) bool {
	tp := itype.GetType(v)

	return tp == itype.NUMBER || tp == itype.BOOL || tp == itype.STRING
}

func Clone(v interface{}) interface{} {
	return deepcopy.Copy(v)
}

func HashID(v string) uint64 {
	return murmur3.Sum64([]byte(v))
}

func CacheID(v string) uint64 {
	return HashID(v)
}

// help struct for String() method
type stringer string

func (s stringer) String() string {
	return string(s)
}
