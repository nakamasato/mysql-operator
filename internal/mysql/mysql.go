package mysql

import (
	"database/sql"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type MySQLConfig struct {
	AdminUser     string
	AdminPassword string
	Host          string
}

type MySQLClient interface {
	Exec(query string, args ...interface{}) error
	Ping() error
	Close()
}

type mysqlClient struct {
	db *sql.DB
}

type fakeMysqlCLient struct {
}

func NewFakeMySQLClient(cfg MySQLConfig) MySQLClient {
	return &fakeMysqlCLient{}
}

func (mc fakeMysqlCLient) Exec(query string, args ...interface{}) error {
	return nil
}

func (mc fakeMysqlCLient) Ping() error {
	return nil
}

func (mc fakeMysqlCLient) Close() {
}

type MySQLClientFactory func(cfg MySQLConfig) MySQLClient

func NewMySQLClient(config MySQLConfig) MySQLClient {
	db, _ := sql.Open("mysql", config.AdminUser+":"+config.AdminPassword+"@tcp("+config.Host+":3306)/")
	// TODO error handling
	return &mysqlClient{db: db}
}

func (mc mysqlClient) Exec(query string, args ...interface{}) error {
	var log = logf.Log.WithName("mysql")
	_, err := mc.db.Exec(query)
	if err != nil {
		log.Error(err, "Failed to execute query", query)
		return err
	}
	return nil
}

func (mc mysqlClient) Ping() error {
	return mc.db.Ping()
}

func (mc mysqlClient) Close() {
	mc.db.Close()
}
