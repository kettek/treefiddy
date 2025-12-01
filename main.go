package main

import (
	"iter"
	"os"

	"github.com/kettek/treefiddy/types"
)

var filterFunc func(a types.FileReference) bool

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

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	filterFunc = func(a types.FileReference) bool {
		return a.Name[0] != '.'
	}

	app := newApp()
	app.setup(dir)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
