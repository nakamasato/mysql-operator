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

type MySQLClient struct {
	db *sql.DB
	// log logr.Logger
}

func NewMySQLClient(config MySQLConfig) MySQLClient {
	db, _ := sql.Open("mysql", config.AdminUser+":"+config.AdminPassword+"@tcp("+config.Host+":3306)/")
	// TODO error handling
	return MySQLClient{db: db}
}

func (mc MySQLClient) Exec(query string) error {
	var log = logf.Log.WithName("mysql")
	_, err := mc.db.Exec(query)
	if err != nil {
		log.Error(err, "Failed to execute query", query)
		return err
	}
	return nil
}

func (mc MySQLClient) Ping() error {
	return mc.db.Ping()
}

func (mc MySQLClient) Close() {
	mc.db.Close()
}
