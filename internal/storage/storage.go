package storage

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrPermissionDenied = errors.New("permission denied")
	ErrAppNotFound      = errors.New("app not found")
)

// TODO: Вынести isAdmin в отдельную сущность
