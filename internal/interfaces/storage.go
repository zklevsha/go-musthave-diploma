package interfaces

import "github.com/zklevsha/go-musthave-diploma/internal/structs"

type Storage interface {
	Close()
	Init() error
	Register(login string, password string) (int, error)
	GetUserID(creds structs.Credentials) (int, error)
	CreateOrder(userid int, orderid int) (bool, error)
	GetOrders(userid int) ([]structs.Order, error)
	GetUnprocessedOrders() ([]int, error)
	SetOrderStatus(id int, status string) (int64, error)
	SetOrderAccrual(id int, accrual float64) (int64, error)
	GetUserBalance(id int) (structs.Balance, error)
	Withdraw(userid int, winthdraw structs.Withdraw) error
	GetWithdrawls(userid int) ([]structs.Withdraw, error)
}
