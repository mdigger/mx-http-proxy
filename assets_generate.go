// +build ignore

package main

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/shurcooL/httpfs/filter"
	"github.com/shurcooL/vfsgen"
)

func main() {
	var fs = filter.Keep(http.Dir("docs"),
		FilesWithExtensions(
			".yaml", ".html", ".ico", ".png", ".xml", ".json", ".txt"))
	if err := vfsgen.Generate(fs, vfsgen.Options{BuildTags: "!dev"}); err != nil {
		log.Fatalln(err)
	}
}

// FilesWithExtensions returns a filter func that selects files (and directories)
// that have any of the given extensions. For example:
//
// 	filter.FilesWithExtensions(".go", ".html")
//
// Would select both .go and .html files and any directories.
func FilesWithExtensions(exts ...string) filter.Func {
	return func(name string, fi os.FileInfo) bool {
		if fi.IsDir() {
			return true // fix dir response
		}
		for _, ext := range exts {
			if path.Ext(name) == ext {
				return true
			}
		}
		return false
	}
}
