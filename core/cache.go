package core

import (
	"sync"
)

type Cache interface {
	Load(key interface{}) (interface{}, bool)
	Store(key, value interface{})
	Delete(key interface{})
}

func NewCache() Cache {
	return new(sync.Map)
}
