package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPickIndexByWeight(t *testing.T) {
	items := []Weighter{
		rank{weight: 10},
		rank{weight: 30},
		rank{weight: 60},
	}

	hit := make(map[int64]int)

	for i := 0; i < 10000; i++ {
		index := PickIndexByWeight(items, 0)
		hit[items[index].Weight()] += 1
	}

	assert.Equal(t, 3, len(hit), "hit.size = 3")
	assert.True(t, hit[60] > hit[30], "hit.60 > hit.30")
	assert.True(t, hit[30] > hit[10], "hit.30 > hit.10")
	t.Log("hit:", hit)
}
