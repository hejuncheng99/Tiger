package TIGER

import "errors"

var (
	DtsNotPointerError = errors.New("dts not a pointer")
	DtsNotSlice        = errors.New("dst not a pointer to slice")
)
