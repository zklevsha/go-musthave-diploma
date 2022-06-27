package interfaces

import "github.com/zklevsha/go-musthave-diploma/internal/structs"

type Storage interface {
	Close()
	Init() error
	Register(login string, password string) error
	GetUserID(creds structs.Credentials) (int, error)
	CreateOrder(userid int, orderid int) (bool, error)
	GetOrders(userid int) ([]structs.Order, error)
	GetUnprocessedOrders() ([]int, error)
	SetOrderStatus(id int, status string) (int64, error)
	SetOrderAccrual(id int, accrual int) (int64, error)
}
