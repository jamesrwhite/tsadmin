// tsadmin/database
package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"username"`
	Password string `json:"password"`
}

type DatabaseStatus struct {
	Metrics   map[string]string
	Variables map[string]string
}

func New(db Database) (*sql.DB, error) {
	return sql.Open("mysql", db.String())
}

func (db *Database) String() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", db.User, db.Password, db.Host, db.Port)
}

func Status(db *sql.DB) (*DatabaseStatus, error) {
	var (
		key           string
		value         string
		Variable_name string
		Value         string
	)

	status := &DatabaseStatus{
		Metrics:   make(map[string]string),
		Variables: make(map[string]string),
	}

	// Fetch all the db metrics
	rows, err := db.Query("SELECT VARIABLE_NAME AS 'key', VARIABLE_VALUE AS 'value' FROM GLOBAL_STATUS")

	// Handle query errors
	if err != nil {
		return status, err
	}

	defer rows.Close()

	// Loop each variable in the server status
	for rows.Next() {
		err := rows.Scan(&key, &value)

		// Handle row reading errors
		if err != nil {
			return status, err
		}

		status.Metrics[key] = value
	}

	// Check for any remaining errors
	err = rows.Err()

	if err != nil {
		return status, err
	}

	// Fetch all the db metrics
	rows, err = db.Query("SHOW GLOBAL VARIABLES")

	// Handle query errors
	if err != nil {
		return status, err
	}

	defer rows.Close()

	// Loop each variable in the server status
	for rows.Next() {
		err := rows.Scan(&Variable_name, &Value)

		// Handle row reading errors
		if err != nil {
			return status, err
		}

		status.Variables[Variable_name] = Value
	}

	// Check for any remaining errors
	err = rows.Err()

	if err != nil {
		return status, err
	}

	return status, nil
}
