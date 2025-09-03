package file

import "io"

type Reader struct {
	io.ReadCloser
	Size int64
	Name string
}
