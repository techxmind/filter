package core

import (
	"errors"
	"strconv"
	"strings"

	"github.com/techxmind/go-utils/object"
)

var _assignmentFactory AssignmentFactory

func init() {
	_assignmentFactory = &stdAssignmentFactory{
		assignments: map[string]Assignment{
			"=": &EqualAssignment{},
			"+": &MergeAssignment{},
			"-": &DeleteAssignment{},
		},
	}
}

type Assignment interface {
	Run(ctx *Context, data interface{}, key string, val interface{})
	PrepareValue(value interface{}) (interface{}, error)
}

type BaseAssignmentPrepareValue struct{}

func (self *BaseAssignmentPrepareValue) PrepareValue(value interface{}) (interface{}, error) {
	return value, nil
}

type AssignmentFactory interface {
	Get(string) Assignment
	Register(Assignment, ...string)
}

// GetAssignmentFactory return AssignmentFactory for registering new Assigment
//
func GetAssignmentFactory() AssignmentFactory {
	return _assignmentFactory
}

type stdAssignmentFactory struct {
	assignments map[string]Assignment
}

func (self *stdAssignmentFactory) Get(name string) Assignment {
	if assignment, ok := self.assignments[name]; ok {
		return assignment
	}

	return nil
}

func (self *stdAssignmentFactory) Register(op Assignment, names ...string) {
	for _, name := range names {
		self.assignments[name] = op
	}
}

//["key", "=", "val"]
type Setter interface {
	AssignmentSet(key string, value interface{}) bool
}

type EqualAssignment struct{ BaseAssignmentPrepareValue }

func (self *EqualAssignment) Run(_ *Context, data interface{}, key string, value interface{}) {
	if v, ok := data.(Setter); ok {
		if v.AssignmentSet(key, value) {
			return
		}
	}

	keys := strings.Split(key, ".")
	lastKey := keys[len(keys)-1]

	var obj interface{}
	if len(keys) > 1 {
		obj, _ = object.GetObject(data, strings.Join(keys[:len(keys)-1], "."), true)
	} else {
		obj = data
	}

	if obj == nil {
		return
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		if IsScalar(value) {
			v[lastKey] = value
		} else {
			v[lastKey] = Clone(value)
		}
	case []interface{}:
		if index, err := strconv.ParseInt(lastKey, 10, 32); err == nil {
			if int(index) >= len(v) {
				return
			}
			if IsScalar(value) {
				v[int(index)] = value
			} else {
				v[int(index)] = Clone(value)
			}
		}
	}
}

//["key", "+", {}]
type Merger interface {
	AssignmentMerge(key string, value interface{}) bool
}
type MergeAssignment struct{}

func (self *MergeAssignment) Run(_ *Context, data interface{}, key string, value interface{}) {
	if v, ok := data.(Merger); ok {
		if v.AssignmentMerge(key, value) {
			return
		}
	}

	keys := strings.Split(key, ".")
	lastKey := keys[len(keys)-1]

	var obj interface{}

	if len(keys) > 1 {
		obj, _ = object.GetObject(data, strings.Join(keys[:len(keys)-1], "."), true)
	} else {
		obj = data
	}

	if obj == nil {
		return
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		if robj, ok := v[lastKey]; !ok {
			v[lastKey] = value
		} else {
			if robj2, ok := robj.(map[string]interface{}); ok {
				for ikey, ivalue := range value.(map[string]interface{}) {
					if IsScalar(ivalue) {
						robj2[ikey] = ivalue
					} else {
						robj2[ikey] = Clone(ivalue)
					}
				}
			}
		}
	}
}

func (self *MergeAssignment) PrepareValue(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, errors.New("assignment[+] value must not be nil")
	}

	if _, ok := value.(map[string]interface{}); !ok {
		return nil, errors.New("assignment[+] value must be map[string]interface{}")
	}

	return value, nil
}

//["key", "-", ["key"]]
type Deleter interface {
	AssignmentDelete(key string, value interface{}) bool
}
type DeleteAssignment struct{}

func (self *DeleteAssignment) Run(_ *Context, data interface{}, key string, value interface{}) {
	// can use '$' or '.' to specify root path
	// ["$", "-", "key1,key2.."]
	// [".", "-", "key1,key2.."]
	key = strings.TrimLeft(strings.TrimLeft(key, "$"), ".")

	if v, ok := data.(Deleter); ok {
		if v.AssignmentDelete(key, value) {
			return
		}
	}

	obj, _ := object.GetObject(data, key, false)

	if obj == nil {
		return
	}

	if v, ok := obj.(map[string]interface{}); ok {
		for _, key := range value.([]interface{}) {
			if _, ok := v[key.(string)]; ok {
				delete(v, key.(string))
			}
		}
	}
}

func (self *DeleteAssignment) PrepareValue(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, errors.New("assignment[-] value must be list")
	}

	list := ToArray(value)

	if len(list) == 0 {
		return nil, errors.New("assignment[-] value must be list")
	}

	for _, item := range list {
		if _, ok := item.(string); !ok {
			return nil, errors.New("assignment[-] value must be string or []string")
		}
	}

	return list, nil
}
