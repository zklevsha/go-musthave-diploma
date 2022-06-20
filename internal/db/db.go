package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DBConnector struct {
	DSN        string
	Ctx        context.Context
	Pool       *pgxpool.Pool
	initalized bool
}

// func (d *DBConnector) checkInit() error {
// 	if !d.initalized {
// 		err := fmt.Errorf("DbConnector is not initiliazed (run DBConnector.Init() to initilize)")
// 		return err
// 	}
// 	return nil
// }

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

func (d *DBConnector) CreateTables() error {
	conn, err := d.Pool.Acquire(d.Ctx)
	defer conn.Release()
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}

	usersSQL := `CREATE TABLE IF NOT EXISTS users (
		id serial PRIMARY KEY,
		username VARCHAR ( 50 ) UNIQUE NOT NULL,
		password TEXT NOT NULL);`

	tokensSQL := `CREATE TABLE IF NOT EXISTS tokens (
		id serial PRIMARY KEY,
		userid int REFERENCES users (Id),
		token TEXT NOT NULL,
		created_ts bigint);
	`

	_, err = conn.Exec(d.Ctx, usersSQL)
	if err != nil {
		return fmt.Errorf("cant create users table: %s", err.Error())
	}

	_, err = conn.Exec(d.Ctx, tokensSQL)
	if err != nil {
		return fmt.Errorf("cant create tokens table: %s", err.Error())
	}
	return nil
}
