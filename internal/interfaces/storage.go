package interfaces

type Storage interface {
	Close()
	Init() error
	Register(login string, password string) error
}
