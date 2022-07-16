package build

import (
	"fmt"
	"os"
	"path/filepath"
)

func mustAbs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		panic(fmt.Errorf("failed to get absolute path to %s", p))
	}
	return a
}

// makeWritable attempts to make the given path writable by its owner.
func makeWritable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	err = os.Chmod(path, info.Mode()|0700)
	if err != nil {
		return err
	}
	return nil
}

var contentType = (func() (ret map[string]string) {
	ret = make(map[string]string)

	ret["ico"] = "image/x-icon"
	ret["svg"] = "image/svg+xml"
	ret["png"] = "image/png"
	ret["gif"] = "image/gif"
	ret["jpeg"] = "image/jpeg"
	ret["jpg"] = "image/jpeg"
	ret["jpe"] = "image/jpeg"
	ret["html"] = "text/html;charset=utf-8"
	ret["js"] = "text/javascript;charset=utf-8"
	ret["css"] = "text/css; charset=utf-8"
	ret["mp3"] = "audio/mp3"
	ret["mp4"] = "video/mpeg4"

	return
})()

func isCss(filename string) bool {
	return TestExpr(`\.s[ca]ss$`, filename) || TestExpr(`\.css$`, filename)
}
