package interfaces

import "github.com/zklevsha/go-musthave-diploma/internal/structs"

type Storage interface {
	Close()
	Init() error
	Register(login string, password string) error
	GetUserID(creds structs.Credentials) (int, error)
}
