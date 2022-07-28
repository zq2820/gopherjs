package es

import (
	"regexp"

	"github.com/speps/go-hashids"
)

func Import(path string) func(...string) string {
	testModule, regexErr := regexp.Compile(`\.module\.`)
	preId := ""
	if regexErr == nil {
		if testModule.MatchString(path) {
			ints := make([]int, len(path))
			for i, val := range path {
				ints[i] = int(val)
			}

			hashId := hashids.NewData()
			hashId.Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			h := hashids.NewWithData(hashId)
			res, encodeError := h.Encode(ints)
			if encodeError == nil {
				preId = res[0:8] + "_"
			} else {
				panic(encodeError)
			}
		}
	} else {
		panic(regexErr)
	}

	return func(classname ...string) string {
		if len(classname) > 0 {
			return preId + classname[0]
		} else {
			return path
		}
	}
}
