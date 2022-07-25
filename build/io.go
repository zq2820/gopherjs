package build

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/gorilla/websocket"
)

type IoWriter interface {
	Write(path string, content []byte)
}

type Packet struct {
	Js  [][2]string `json:"js"`
	Css [][2]string `json:"css"`
}

func newPacket() *Packet {
	return &Packet{
		Js:  make([][2]string, 0),
		Css: make([][2]string, 0),
	}
}

type DevServer struct {
	Port   int
	files  map[string][]byte
	socket *websocket.Conn
	packet *Packet
}

func (server *DevServer) Write(path string, content []byte) {
	if path == "index.html" {
		path = ""
	}
	server.files[path] = content
}

func NewDevServer(port int) *DevServer {
	server := &DevServer{
		Port:   port,
		files:  make(map[string][]byte),
		packet: newPacket(),
	}
	copyDir("./helloworld/public", "", server)

	go (func() {
		upgrade := websocket.Upgrader{}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			url := r.URL.Path[1:]
			for suffix, prop := range contentType {
				if TestExpr(fmt.Sprintf(`\.%s$`, suffix), url) {
					w.Header().Set("content-type", prop)
					break
				}
			}

			if content := server.files[url]; content != nil {
				w.Write(content)
			} else {
				pwd, _ := os.Getwd()
				if file, err := ioutil.ReadFile(path.Join(pwd, url)); err == nil {
					w.Write(file)
				} else {
					w.Write([]byte("404 Not Found"))
				}
			}
		})
		http.HandleFunc("/update/ws", func(w http.ResponseWriter, r *http.Request) {
			connect, connectErr := upgrade.Upgrade(w, r, nil)
			if connectErr == nil {
				server.socket = connect
			} else {
				w.Write([]byte(connectErr.Error()))
			}
		})
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			panic(err)
		}
	})()
	return server
}

type FileWriter struct {
	dist string
}

func (writer *FileWriter) Write(path string, content []byte) {
	file, fileErr := os.Create(writer.dist + "/" + path)
	if fileErr == nil {
		_, writeErr := file.Write(content)
		file.Close()
		if writeErr != nil {
			panic(writeErr)
		}
	} else {
		panic(fileErr)
	}
}

func NewFileWriter(dist string) *FileWriter {
	dir := os.DirFS(dist)
	_, err := fs.ReadDir(dir, ".")

	if err == nil {
		os.RemoveAll(dist)
	}
	os.Mkdir(dist, os.FileMode(0777))
	copyDir("./public", "./dist", nil)
	return &FileWriter{
		dist: dist,
	}
}

const (
	Js  = 1
	Css = 2
)

func (context *DevServer) Stash(packetType int, key string, value []byte) {
	if packetType == Js {
		context.packet.Js = append(context.packet.Js, [2]string{key, string(value)})
	} else if packetType == Css {
		context.packet.Css = append(context.packet.Css, [2]string{key, string(value)})
	}
}

func (context *DevServer) Send() {
	defer (func() {
		context.packet = newPacket()
	})()

	if context.socket == nil || (len(context.packet.Js) == 0 && len(context.packet.Css) == 0) {
		return
	}

	message, _ := json.Marshal(context.packet)

	if context.socket != nil {
		context.socket.WriteMessage(websocket.BinaryMessage, []byte(message))
	}
	context.packet = newPacket()
}

func copyDir(dir, dist string, server *DevServer) {
	if files, err := os.ReadDir(dir); err == nil {
		for _, file := range files {
			if !(file.IsDir() && file.Name() != "index.html") {
				if server != nil {
					if file.IsDir() {
						copyDir(file.Name(), dist, server)
					} else {
						if content, err := os.ReadFile(path.Join(dir, file.Name())); err == nil {
							server.Write(file.Name(), content)
						} else {
							panic(err)
						}
					}
				} else {
					exec.Command(fmt.Sprintf("cp -r %s %s", file.Name(), dist))
				}
			}
		}
	} else {
		panic(err)
	}
}
