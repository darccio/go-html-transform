package main

import (
	"flag"
	"fmt"
)

type stringEnum struct {
	val string
	e map[string]struct{}
}

func (e *stringEnum) Set(v string) error {
	if _, ok := e.e[v]; ok {
		e.val = v
		return nil
	}
	return fmt.Errorf("Value %q not a valid enum value")
}

func(e *stringEnum) String() string {
	return e.val
}

func StringEnum(name string, e map[string]struct{}, d, doc string) flag.Value {
	val := &stringEnum{
	val: d,
	e: e,
	}
	flag.Var(val, name, doc)
	return val
}