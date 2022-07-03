package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zklevsha/go-musthave-diploma/internal/structs"
)

type DBConnector struct {
	DSN        string
	Ctx        context.Context
	Pool       *pgxpool.Pool
	initalized bool
}

func (d *DBConnector) checkInit() error {
	if !d.initalized {
		err := fmt.Errorf("DbConnector is not initiliazed (run DBConnector.Init() to initilize)")
		return err
	}
	return nil
}

func (d *DBConnector) Register(login string, password string) (int, error) {
	err := d.checkInit()
	if err != nil {
		return -1, err
	}

	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	// Check if user don`t exists
	var counter int
	sql := `select count(id) from users where login=$1;`
	err = conn.QueryRow(d.Ctx, sql, login).Scan(&counter)
	if err != nil {
		return -1, fmt.Errorf("failed to query users table: %s", err.Error())
	}
	if counter != 0 {
		return -1, structs.ErrUserAlreadyExists
	}

	// adding new user
	var id int
	sql = `INSERT INTO users (login, password)
		   VALUES($1, $2)
		   RETURNING id;`
	err = conn.QueryRow(d.Ctx, sql, login, password).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("failed to create user id DB: %s", err.Error())
	}
	return id, nil
}

func (d *DBConnector) GetUserID(creds structs.Credentials) (int, error) {
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	var id int
	var password string
	sql := `SELECT id, password FROM users WHERE login=$1;`
	row := conn.QueryRow(d.Ctx, sql, creds.Login)

	switch err := row.Scan(&id, &password); err {
	case pgx.ErrNoRows:
		return -1, structs.ErrUserAuth
	case nil:
		return id, nil
	default:
		e := fmt.Errorf("unknown error while authenticating user: %s", err.Error())
		return -1, e
	}

}

func (d *DBConnector) Init() error {
	p, err := pgxpool.Connect(d.Ctx, d.DSN)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	d.Pool = p
	err = d.CreateTables()
	if err != nil {
		return err
	}
	d.initalized = true
	return nil
}

func (d *DBConnector) Close() {
	if d.initalized {
		d.Pool.Close()
		d.initalized = false
	}
}

func (d *DBConnector) CreateOrder(userid int, orderid int) (bool, error) {
	err := d.checkInit()
	if err != nil {
		return false, err
	}

	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return false, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	// Check if order already exists
	var uid int
	sql := `select userid from orders where id=$1;`
	row := conn.QueryRow(d.Ctx, sql, orderid)

	switch err := row.Scan(&uid); err {
	case pgx.ErrNoRows:
		break
	case nil:
		// user already sent this order
		if uid == userid {
			return false, nil
		} else {
			// orderid was used by different user
			return false, structs.ErrOrderIDAlreadyUsed
		}
	default:
		e := fmt.Errorf("unknown error while creating order: %s", err.Error())
		return false, e
	}

	// creating new order
	now := time.Now().Unix()
	sql = `INSERT INTO orders (id, created_ts, userid)
		   VALUES($1, $2, $3);`
	_, err = conn.Exec(d.Ctx, sql, orderid, now, userid)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DBConnector) GetOrders(userid int) ([]structs.Order, error) {
	err := d.checkInit()
	if err != nil {
		return nil, err
	}

	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	sql := `SELECT id, status, accrual, created_ts 
			FROM orders
			WHERE userid=$1`
	rows, err := conn.Query(d.Ctx, sql, userid)
	if err != nil {
		e := fmt.Errorf("failed to query orders table: %s", err.Error())
		return nil, e
	}
	defer rows.Close()

	var orders []structs.Order
	for rows.Next() {
		var orderNumber int
		var status string
		var accrual *float32
		var createdTS int64

		if err := rows.Scan(&orderNumber, &status, &accrual, &createdTS); err != nil {
			e := fmt.Errorf("failed to scan row from orders table: %s", err.Error())
			return nil, e
		}
		order := structs.Order{Number: fmt.Sprint(orderNumber),
			Status:     status,
			Accrual:    accrual,
			UploadedAt: time.Unix(createdTS, 0).Format("2006-01-02T15:04:05-07:00")}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		e := fmt.Errorf("error(s) occured during orders table scanning: %s", err.Error())
		return nil, e
	}

	return orders, nil
}

func (d *DBConnector) GetUnprocessedOrders() ([]int, error) {
	err := d.checkInit()
	if err != nil {
		return nil, err
	}

	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	sql := `SELECT id
			FROM orders
			WHERE status='NEW' OR status='PROCESSING'`
	rows, err := conn.Query(d.Ctx, sql)
	if err != nil {
		e := fmt.Errorf("failed to query orders table: %s", err.Error())
		return nil, e
	}
	defer rows.Close()

	var orders []int
	for rows.Next() {
		var orderNumber int
		if err := rows.Scan(&orderNumber); err != nil {
			e := fmt.Errorf("failed to scan row from orders table: %s", err.Error())
			return nil, e
		}
		orders = append(orders, orderNumber)
	}

	if err := rows.Err(); err != nil {
		e := fmt.Errorf("error(s) occured during orders table scanning: %s", err.Error())
		return nil, e
	}
	return orders, nil
}

func (d *DBConnector) SetOrderStatus(id int, status string) (int64, error) {
	err := d.checkInit()
	if err != nil {
		return -1, err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	sql := `UPDATE orders SET status = $2 WHERE id = $1;`
	res, err := conn.Exec(d.Ctx, sql, id, status)
	if err != nil {
		return -1, fmt.Errorf("failed exec sql: %s", err.Error())
	}
	count := res.RowsAffected()
	return count, nil
}

func (d *DBConnector) SetOrderAccrual(id int, accrual float32) (int64, error) {
	err := d.checkInit()
	if err != nil {
		return -1, err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	sql := `UPDATE orders SET accrual = $2 WHERE id = $1;`
	res, err := conn.Exec(d.Ctx, sql, id, accrual)
	if err != nil {
		return -1, fmt.Errorf("failed exec sql: %s", err.Error())
	}
	count := res.RowsAffected()
	return count, nil
}

func (d *DBConnector) GetUserBalance(id int) (structs.Balance, error) {
	err := d.checkInit()
	if err != nil {
		return structs.Balance{}, err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return structs.Balance{}, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	// geting user accrual
	sql := `SELECT COALESCE(SUM(accrual),0) AS acc_total
			FROM orders
			WHERE userid = $1;`
	var accTotal float32
	row := conn.QueryRow(d.Ctx, sql, id)
	err = row.Scan(&accTotal)
	if err != nil {
		return structs.Balance{}, fmt.Errorf("failed to query orders table: %s", err.Error())
	}

	// geting user withdrawals
	sql = `SELECT COALESCE(SUM(amount),0) AS withdrawals_total
		   FROM withdrawals
		   WHERE userid = $1`
	var wdTotal float32
	row = conn.QueryRow(d.Ctx, sql, id)
	err = row.Scan(&wdTotal)
	if err != nil {
		return structs.Balance{}, fmt.Errorf("failed to query withdrawals table: %s", err.Error())
	}

	balance := structs.Balance{Current: accTotal - wdTotal, Withdrawn: wdTotal}
	return balance, nil
}

func (d *DBConnector) Withdraw(userid int, winthdraw structs.Withdraw) error {
	err := d.checkInit()
	if err != nil {
		return err
	}

	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	sql := `INSERT INTO withdrawals (userid, orderid, amount, processed_at)
		   VALUES($1, $2, $3, $4);`
	_, err = conn.Exec(d.Ctx, sql, userid, winthdraw.Order, winthdraw.Sum, time.Now().Unix())

	return err
}

func (d *DBConnector) GetWithdrawls(userid int) ([]structs.Withdraw, error) {
	err := d.checkInit()
	if err != nil {
		return nil, err
	}

	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	sql := `SELECT amount, orderid, processed_at
			FROM withdrawals
			WHERE userid=$1`
	rows, err := conn.Query(d.Ctx, sql, userid)
	if err != nil {
		e := fmt.Errorf("failed to query withdrawals table: %s", err.Error())
		return nil, e
	}
	defer rows.Close()

	var withdrawals []structs.Withdraw
	for rows.Next() {
		var orderid int
		var amount float32
		var processedAt int64

		if err := rows.Scan(&amount, &orderid, &processedAt); err != nil {
			e := fmt.Errorf("failed to scan row from withdrawals table: %s", err.Error())
			return nil, e
		}
		withdraw := structs.Withdraw{
			Order:       fmt.Sprint(orderid),
			Sum:         amount,
			ProcessedAt: time.Unix(processedAt, 0).Format("2006-01-02T15:04:05-07:00")}
		withdrawals = append(withdrawals, withdraw)
	}

	if err := rows.Err(); err != nil {
		e := fmt.Errorf("error(s) occured during withdrawals table scanning: %s", err.Error())
		return nil, e
	}

	return withdrawals, nil

}

func (d *DBConnector) CreateTables() error {
	conn, err := d.Pool.Acquire(d.Ctx)
	defer conn.Release()
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}

	usersSQL := `CREATE TABLE IF NOT EXISTS users (
		id serial PRIMARY KEY,
		login VARCHAR ( 50 ) UNIQUE NOT NULL,
		password TEXT NOT NULL);`

	_, err = conn.Exec(d.Ctx, usersSQL)
	if err != nil {
		return fmt.Errorf("cant create users table: %s", err.Error())
	}

	ordersSQL := `CREATE TABLE IF NOT EXISTS orders (
		id bigint PRIMARY KEY,
		status VARCHAR (15) DEFAULT 'NEW',
		accrual real,
		created_ts bigint,
		userid integer REFERENCES users (id));`

	_, err = conn.Exec(d.Ctx, ordersSQL)
	if err != nil {
		return fmt.Errorf("cant create orders table: %s", err.Error())
	}

	withdrawalsSQL := `CREATE TABLE IF NOT EXISTS withdrawals (
		id serial PRIMARY KEY,
		amount real NOT NULL,
		processed_at bigint,
		orderid integer,
		userid integer REFERENCES users (id));`

	_, err = conn.Exec(d.Ctx, withdrawalsSQL)
	if err != nil {
		return fmt.Errorf("cant create orders withdrawals: %s", err.Error())
	}

	return nil
}
