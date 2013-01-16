package main

import (
	"flag"
	"fmt"
)

// TODO(jwall): Move these to a common flags library.
type stringEnum struct {
	val string
	e   map[string]struct{}
}

func (e *stringEnum) Set(v string) error {
	if _, ok := e.e[v]; ok {
		e.val = v
		return nil
	}
	return fmt.Errorf("Value %q not a valid enum value")
}

func (e *stringEnum) String() string {
	var list []string
	for k, _ := range e.e {
		list = append(list, k)
	}
	return e.val + fmt.Sprintf(" from %v", list)
}

func (e *stringEnum) Value() string {
	return e.val
}

func StringEnum(name string, e map[string]struct{}, d, doc string) *stringEnum {
	val := &stringEnum{
		val: d,
		e:   e,
	}
	flag.Var(val, name, doc)
	return val
}
