package build

import (
	"regexp"
)

type StringSet struct {
	value map[string]bool
}

func NewStringSet() *StringSet {
	return &StringSet{
		value: make(map[string]bool),
	}
}

func (set *StringSet) Add(key string) {
	set.value[key] = true
}

func (set *StringSet) Remove(key string) {
	delete(set.value, key)
}

func (set *StringSet) Len() int {
	return len(set.value)
}

func (set *StringSet) Has(key string) bool {
	return set.value[key]
}

func (set *StringSet) Clear() {
	set.value = make(map[string]bool)
}

func TestExpr(expr string, str string) bool {
	test, regexError := regexp.Compile(expr)
	if regexError == nil {
		return test.MatchString(str)
	} else {
		panic(regexError)
	}
}
