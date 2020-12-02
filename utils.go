package filter

import (
	"encoding/json"
	"math/rand"
	"strconv"

	"github.com/techxmind/filter/core"
)

func generateFilterName(filter interface{}) string {
	v, _ := json.Marshal(filter)
	return strconv.FormatUint(core.HashID(string(v)), 36)
}

type Weighter interface {
	Weight() int64
}

func PickIndexByWeight(items []Weighter, totalWeight int64) int {
	if totalWeight == 0 {
		for _, item := range items {
			totalWeight += item.Weight()
		}
	}

	if totalWeight == 0 {
		return 0
	}

	choose := rand.Int63n(totalWeight) + 1
	line := int64(0)

	for i, b := range items {
		line += b.Weight()
		if choose <= line {
			return i
		}
	}

	return 0
}

// help func to create []interface{} for unit test
func arr(vals ...interface{}) []interface{} {
	return core.ToArray(vals)
}
