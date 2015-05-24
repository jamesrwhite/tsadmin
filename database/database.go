// tsadmin/database
package database

import (
	"database/sql"
	"fmt"
	"strconv"

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
	Metrics   DatabaseMetrics
	Variables DatabaseVariables
}

type DatabaseMetrics struct {
	Connections int `json:"connections"`
	Uptime      int `json:"uptime"`
}

type DatabaseVariables struct {
	MaxConnections int `json:"max_connections"`
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
		Metrics:   DatabaseMetrics{},
		Variables: DatabaseVariables{},
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

		// Current connections
		if key == "THREADS_CONNECTED" {
			connections, _ := strconv.Atoi(value)
			status.Metrics.Connections = connections
		} else if key == "UPTIME" {
			uptime, _ := strconv.Atoi(value)
			status.Metrics.Uptime = uptime
		}
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

		// Max allowed connections
		if Variable_name == "max_connections" {
			maxConnections, _ := strconv.Atoi(Value)
			status.Variables.MaxConnections = maxConnections
		}
	}

	// Check for any remaining errors
	err = rows.Err()

	if err != nil {
		return status, err
	}

	return status, nil
}
