package db

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxFn can be run inside a transaction
type TxFn func(pgx.Tx) error

var pool *pgxpool.Pool

// InitializeConnectionPool must be called before connecting to db
func InitializeConnectionPool(url string) {
	p, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("ERROR db.connect: ", err)
	}
	pool = p
	log.Println("INFO maximum db pool connections", p.Config().MaxConns)
}

// WithTransaction executes a function within a transaction
func WithTransaction(fn TxFn) error {
	var err error
	var tx pgx.Tx
	if pool == nil {
		return errors.New("connection pool is nil. see db.InitializeConnectionPool(connectionURL)")
	}
	if tx, err = pool.Begin(context.Background()); err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(context.Background())
			panic(p)
		} else if err != nil {
			tx.Rollback(context.Background())
		} else {
			err = tx.Commit(context.Background())
			if err != nil {
				log.Println("ERROR WithTransaction Commit", err)
			}
		}
	}()
	err = fn(tx)
	return err
}

// NextID gets a new id from the pg sequence generator
func NextID() int64 {
	if pool == nil {
		log.Fatal("Connection pool is nil. Initialize it with db.InitializeConnectionPool(connectionURL)")
	}
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Panic("ERROR Can't get connection for nextID", err)
	}
	defer conn.Release()
	var id int64
	err = conn.QueryRow(context.Background(), "select id_generator()").Scan(&id)
	if err != nil {
		log.Panic("ERROR Can't get nextID", err)
	}
	return id
}
