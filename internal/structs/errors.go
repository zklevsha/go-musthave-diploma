package structs

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserAuth = errors.New("authentication failed")
var ErrOrderIdAlreadyUsed = errors.New("order id already used by another user")
