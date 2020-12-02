package core

import (
	"fmt"
	"io"
	"strings"
)

type Trace interface {
	Enter(name string) Trace
	Leave(name string) Trace
	Log(a ...interface{}) Trace
}

type stdTrace struct {
	level        int
	levelPadding string
	w            io.Writer
}

func NewTrace(w io.Writer) Trace {
	return &stdTrace{
		w: w,
	}
}

func (t *stdTrace) Enter(name string) Trace {
	t.Log(name)
	t.level++
	t.setPadding()

	return t
}

func (t *stdTrace) Leave(name string) Trace {
	t.level--
	t.setPadding()
	return t
}

func (t *stdTrace) setPadding() {
	if t.level <= 0 {
		t.levelPadding = ""
		return
	}

	t.levelPadding = strings.Repeat("  ", t.level)
}

func (t *stdTrace) Log(a ...interface{}) Trace {
	if t.levelPadding != "" {
		t.w.Write([]byte(t.levelPadding))
	}

	// color bool vars
	for i := 0; i < len(a); i++ {
		if b, ok := a[i].(bool); ok {
			if b {
				a[i] = "\033[0;32mtrue\033[0m"
			} else {
				a[i] = "\033[0;31mfalse\033[0m"
			}
		}
	}

	t.w.Write([]byte(fmt.Sprintln(a...)))
	return t
}
