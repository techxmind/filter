package core

import (
	"io/ioutil"
	"log"
)

type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

var (
	Logger StdLogger = log.New(ioutil.Discard, "[filter]", log.LstdFlags)
)
