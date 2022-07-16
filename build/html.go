package build

import (
	"bytes"
	"os"

	"github.com/tdewolff/minify/v2/minify"
	"golang.org/x/net/html"
)

type HtmlInstance struct {
	dist      string
	root      *html.Node
	head      *html.Node
	body      *html.Node
	cssChunks map[string]bool
	jsChunks  map[string]bool
	io        IoWriter
}

func InitHtml(entry string, dist string, io IoWriter) *HtmlInstance {
	res, readErr := os.ReadFile(entry + "/index.html")
	content := ""
	if readErr == nil {
		content = string(res)
	} else {
		panic(readErr)
	}

	nodes, parseErr := html.Parse(bytes.NewBufferString(content))

	if parseErr == nil {
		htmlInstance := &HtmlInstance{
			dist:      dist,
			root:      nodes,
			head:      getHead(nodes),
			body:      getBody(nodes),
			cssChunks: make(map[string]bool),
			jsChunks:  make(map[string]bool),
			io:        io,
		}

		return htmlInstance
	} else {
		panic(parseErr)
	}
}

func (context *HtmlInstance) InsertCss(css string) {
	if !context.cssChunks[css] {
		context.cssChunks[css] = true
		context.head.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "link",
			Attr: [](html.Attribute){
				html.Attribute{
					Key: "rel",
					Val: "stylesheet",
				},
				html.Attribute{
					Key: "href",
					Val: css,
				},
			},
		})
	}
}

func (context *HtmlInstance) InsertJs(js string) {
	if !context.jsChunks[js] {
		context.jsChunks[js] = true
		context.body.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "script",
			Attr: [](html.Attribute){
				html.Attribute{
					Key: "src",
					Val: js,
				},
				html.Attribute{
					Key: "type",
					Val: "text/javascript",
				},
			},
		})
	}
}

func (context *HtmlInstance) Generate() {
	buffer := bytes.NewBufferString("")
	html.Render(buffer, context.root)
	minifyHtml, htmlErr := minify.HTML(buffer.String())
	if htmlErr == nil {
		context.io.Write("index.html", []byte(minifyHtml))
	} else {
		panic(htmlErr)
	}
}

func getHead(root *html.Node) *html.Node {
	cur := root.FirstChild

	// 获取doc节点
	for cur != nil {
		if cur.Type == html.ElementNode {
			if cur.Data == "html" {
				cur = cur.FirstChild
				break
			}
		}
		cur = cur.NextSibling
	}

	if cur == nil {
		panic("Not found html tag")
	}

	// 获取head节点
	for cur != nil {
		if cur.Type == html.ElementNode {
			if cur.Data == "head" {
				break
			}
		}
		cur = cur.NextSibling
	}

	if cur == nil {
		panic("Not found head tag")
	}

	return cur
}

func getBody(root *html.Node) *html.Node {
	cur := root.FirstChild

	// 获取doc节点
	for cur != nil {
		if cur.Type == html.ElementNode {
			if cur.Data == "html" {
				cur = cur.FirstChild
				break
			}
		}
		cur = cur.NextSibling
	}

	if cur == nil {
		panic("Not found html tag")
	}

	// 获取head节点
	for cur != nil {
		if cur.Type == html.ElementNode {
			if cur.Data == "body" {
				break
			}
		}
		cur = cur.NextSibling
	}

	if cur == nil {
		panic("Not found body tag")
	}

	return cur
}
