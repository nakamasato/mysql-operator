package mysql

import (
	"database/sql"
	"errors"
)

type MySQLClients map[string]*sql.DB

var ErrMySQLClientNotFound = errors.New("MySQL client not found")

func (m MySQLClients) GetClient(name string) (*sql.DB, error) {
	mysqlClient, ok := m[name]
	if ok {
		return mysqlClient, nil
	} else {
		return nil, ErrMySQLClientNotFound
	}
}

// Cloase a MySQL client
func (m MySQLClients) Close(name string) error {
	mysqlClient, ok := m[name]
	if !ok {
		return ErrMySQLClientNotFound
	}
	if err := mysqlClient.Close(); err != nil {
		return err
	}
	delete(m, name)
	return nil
}

// Close all MySQL clients.
// Return error immediately when error occurs for a client.
func (m MySQLClients) CloseAll() error {
	for name := range m {
		err := m.Close(name)
		if err != nil {
			return err
		}
	}
	return nil
}
