package structs

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserAuth = errors.New("authentication failed")
var ErrOrderIDAlreadyUsed = errors.New("order id already used by another user")
var ErrToManyRequests = errors.New("to many request to the remote system")
