package main

import (
	"io/fs"
	"iter"
	"os"
)

var (
	filterFunc func(a fs.FileInfo) bool
	sortFunc   func(a, b fs.FileInfo) int
)

func filter[S any](s []S, fn func(S) bool) iter.Seq[S] {
	return func(yield func(s S) bool) {
		for _, v := range s {
			if fn(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

type nodeRef struct {
	path string
	dir  bool
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	/*sortFunc = func(a, b fs.FileInfo) int {
		if a.IsDir() && !b.IsDir() {
			return -1
		}
		return strings.Compare(a.Name(), b.Name())
	}*/
	filterFunc = func(a fs.FileInfo) bool {
		return a.Name()[0] != '.'
	}

	app := newApp()
	app.setup()
	app.setRoot(dir)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
