// Package types provides some implementation-agnostic types to use throughout treefiddy.
package types

// FileReference stores information about a file for use in tree node generation.
type FileReference struct {
	Name string
	Path string
	Dir  bool
}
