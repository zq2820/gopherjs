package build

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"os"
	"strings"

	resolve "path"

	"github.com/gorilla/css/scanner"
	"github.com/speps/go-hashids/v2"
	"github.com/vanng822/css"
	"github.com/wellington/go-libsass"
)

type ScssCompiler struct {
	Entry        string
	Dist         string
	Session      *Session
	tasks        *StringSet
	cssChunks    map[string][]byte
	dependencies map[string]*StringSet
	depended     map[string]*StringSet
}

func NewScssCompiler(entry, dist string) *ScssCompiler {
	compiler := &ScssCompiler{
		Entry:        entry,
		Dist:         dist,
		tasks:        NewStringSet(),
		cssChunks:    make(map[string][]byte),
		dependencies: make(map[string]*StringSet),
		depended:     make(map[string]*StringSet),
	}

	return compiler
}

func (context *ScssCompiler) BeforeCompile(path string) {
	if context.Session.WatchReady() {
		if context.dependencies[path] != nil {
			set := context.dependencies[path]
			for key := range set.value {
				context.depended[key].Remove(path)
			}
			set.Clear()
		}
	}
}

func (context *ScssCompiler) Add(path string, container string) {
	if context.Session.options.Watch && container != "" {
		if context.depended[path] == nil {
			context.depended[path] = NewStringSet()
		}
		context.depended[path].Add(container)
		if context.dependencies[container] == nil {
			context.dependencies[container] = NewStringSet()
		}
		context.dependencies[container].Add(path)
	}
	if context.cssChunks[path] != nil {
		delete(context.cssChunks, path)
	}
	context.tasks.Add(path)
}

func (context *ScssCompiler) BuildScss() {
	for path := range context.tasks.value {
		context.compileFile(path)
		context.tasks.Remove(path)
		if context.Session.WatchReady() {
			if devServer := context.Session.io.(*DevServer); devServer != nil {
				devServer.Stash(Css, path, context.cssChunks[path])
			}
		}
	}
	context.tasks = NewStringSet()

	if context.Session.WatchReady() {
		for path, set := range context.depended {
			if set.Len() == 0 {
				delete(context.depended, path)
				delete(context.cssChunks, path)
				if devServer := context.Session.io.(*DevServer); devServer != nil {
					devServer.Stash(Css, path, []byte{})
				}
			}
		}
	}
}

func (context *ScssCompiler) compileFile(path string) {
	if context.cssChunks[path] == nil {
		content, readErr := os.ReadFile(path)
		if readErr == nil {
			buffer := bytes.NewBuffer(content)
			writer := bytes.NewBufferString("")
			scssCompiler, newErr := libsass.New(writer, buffer)
			scssCompiler.Option(libsass.OutputStyle(libsass.Style["compressed"]))
			splits := strings.Split(path, "/")
			scssCompiler.Option(libsass.IncludePaths([]string{
				strings.Join(splits[0:len(splits)-1], "/"),
			}))
			if newErr == nil {
				compileErr := scssCompiler.Run()
				if compileErr == nil {
					content := writer.String()
					if getIsModule(path) {
						content = context.moduleTransform(path, content)
					}
					if context.Session.options.Watch {
						content = strings.ReplaceAll(content, "\n", "")
					}
					context.cssChunks[path] = []byte(content)
					context.writeToFile(path, context.cssChunks[path])
				} else {
					panic(compileErr)
				}
			} else {
				panic(newErr)
			}
		} else {
			panic(readErr)
		}
	}
}

func (context *ScssCompiler) writeToFile(path string, content []byte) {
	hash := md5.New()
	hash.Write([]byte(path))
	tag := hash.Sum(nil)
	filename := fmt.Sprintf("%x", tag) + ".css"

	context.Session.io.Write(filename, content)
	if !context.Session.options.Watch {
		context.Session.htmlInstance.InsertCss("./" + filename)
	}
}

func getIsModule(filename string) bool {
	return TestExpr(`\.module\.`, filename)
}

func isGlobal(tokens []*scanner.Token) bool {
	test := []string{":", "global", " "}
	for i := 0; i < 3; i++ {
		if tokens[i].Value != test[i] {
			return false
		}
	}
	return true
}

func (context *ScssCompiler) moduleTransform(path string, content string) string {
	res := ""
	classes := css.Parse(content)
	for _, class := range classes.CssRuleList {
		res += context.transformRule(class, path)
	}

	return res
}

func (context *ScssCompiler) transformRule(rule *css.CSSRule, path string) string {
	res := ""
	res += rule.Type.Text() + " "

	for i := 1; i < len(rule.Style.Selector.Tokens); i++ {
		if isGlobal(rule.Style.Selector.Tokens) {
			rule.Style.Selector.Tokens = rule.Style.Selector.Tokens[3:]
		} else {
			if rule.Style.Selector.Tokens[i-1].Value == "." {
				rule.Style.Selector.Tokens[i].Value = getModuleName(path) + "_" + rule.Style.Selector.Tokens[i].Value
			}
		}
	}

	res += rule.Style.Selector.Text()
	if rule.Type == css.IMPORT_RULE {
		value := rule.Style.Selector.Tokens[1].Value
		pathSplit := strings.Split(path, "/")
		cssPath := "./" + resolve.Join(strings.Join(pathSplit[1:len(pathSplit)-1], "/"), value[4:len(value)-1])
		context.tasks.Add(cssPath)

		return ""
	}
	res += "{"
	for i := 0; i < len(rule.Style.Styles); i++ {
		res += rule.Style.Styles[i].Property + ":" + rule.Style.Styles[i].Value.Text() + ";"
	}
	for _, subRule := range rule.Rules {
		res += context.transformRule(subRule, path)
	}
	res += "}"

	return res
}

func getModuleName(path string) string {
	ints := make([]int, len(path))
	for i, val := range path {
		ints[i] = int(val)
	}

	hashId := hashids.NewData()
	hashId.Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	h, newErr := hashids.NewWithData(hashId)
	if newErr == nil {
		res, encodeError := h.Encode(ints)
		if encodeError == nil {
			return res[0:8]
		} else {
			panic(encodeError)
		}
	} else {
		panic(newErr)
	}
}
