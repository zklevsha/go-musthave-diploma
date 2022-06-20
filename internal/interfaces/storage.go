package interfaces

type Storage interface {
	Close()
	Init() error
}
