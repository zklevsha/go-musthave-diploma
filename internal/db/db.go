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

func (d *DBConnector) Register(login string, password string) error {
	err := d.checkInit()
	if err != nil {
		return err
	}

	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	// Check if user don`t exists
	var counter int
	sql := `select count(id) from users where login=$1;`
	row := conn.QueryRow(d.Ctx, sql, login)
	err = row.Scan(&counter)
	if err != nil {
		return fmt.Errorf("failed to query users table: %s", err.Error())
	}
	if counter != 0 {
		return structs.ErrUserAlreadyExists
	}

	// adding new user
	sql = `INSERT INTO users (login, password)
		   VALUES($1, $2);`
	_, err = conn.Exec(d.Ctx, sql, login, password)
	return err
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
			return false, structs.ErrOrderIdAlreadyUsed
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
		accural int,
		created_ts bigint,
		userid integer REFERENCES users (id));`

	_, err = conn.Exec(d.Ctx, ordersSQL)
	if err != nil {
		return fmt.Errorf("cant create orders table: %s", err.Error())
	}

	return nil
}
