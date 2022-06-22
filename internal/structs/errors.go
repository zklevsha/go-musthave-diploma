package structs

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserAuth = errors.New("authentication failed")
