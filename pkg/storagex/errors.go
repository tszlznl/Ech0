package storagex

import "errors"

var (
	ErrInvalidPath   = errors.New("storagex: invalid path")
	ErrNotFound      = errors.New("storagex: not found")
	ErrAlreadyExists = errors.New("storagex: already exists")
	ErrUnsupported   = errors.New("storagex: unsupported operation")
)
