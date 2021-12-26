package mysql

import (
	"context"
	"database/sql"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type MySQLConfig struct {
	AdminUser     string
	AdminPassword string
	Host          string
}

type MySQLClient interface {
	Exec(query string) error
	Ping() error
	Close()
}

type mysqlClient struct {
	db *sql.DB
}

type fakeMysqlClient struct {
}

func NewFakeMySQLClient(cfg MySQLConfig) (MySQLClient, error) {
	return &fakeMysqlClient{}, nil
}

func (mc fakeMysqlClient) Exec(query string) error {
	return nil
}

func (mc fakeMysqlClient) Ping() error {
	return nil
}

func (mc fakeMysqlClient) Close() {
}

type MySQLClientFactory func(cfg MySQLConfig) (MySQLClient, error)

func NewMySQLClient(config MySQLConfig) (MySQLClient, error) {
	db, err := sql.Open("mysql", config.AdminUser+":"+config.AdminPassword+"@tcp("+config.Host+":3306)/")
	// TODO error handling
	return &mysqlClient{db: db}, err
}

func (mc mysqlClient) Exec(query string) error {
	var log = logf.Log.WithName("mysql")
	_, err := mc.db.Exec(query)
	if err != nil {
		log.Error(err, "Failed to execute query", query)
		return err
	}
	return nil
}

func (mc mysqlClient) Ping() error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		err := mc.db.Ping()
		done <- err
	}()
	select {
	case e := <-done:
		return e
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (mc mysqlClient) Close() {
	mc.db.Close()
}
