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
	Metadata  DatabaseMetadata  `json:"metadata"`
	Metrics   DatabaseMetrics   `json:"metrics"`
	Variables DatabaseVariables `json:"variables"`
}

type DatabaseMetadata struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DatabaseMetrics struct {
	Connections int `json:"connections"`
	Uptime      int `json:"uptime"`
	QueriesPerSecond int `json:"queries_per_second"`
	Queries int `json:"queries"`
}

type DatabaseVariables struct {
	MaxConnections int `json:"max_connections"`
}

func (db *Database) String() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", db.User, db.Password, db.Host, db.Port)
}

func Status(db Database) (*DatabaseStatus, error) {
	var (
		key           string
		value         string
		Variable_name string
		Value         string
	)

	status := &DatabaseStatus{
		Metadata:  DatabaseMetadata{
			Name: db.Name,
			Host: db.Host,
			Port: db.Port,
		},
		Metrics:   DatabaseMetrics{},
		Variables: DatabaseVariables{},
	}

	// Connect to the database
	conn, _ := sql.Open("mysql", db.String())
	defer conn.Close()

	// Fetch all the db metrics
	rows, err := conn.Query("SELECT VARIABLE_NAME AS 'key', VARIABLE_VALUE AS 'value' FROM GLOBAL_STATUS")

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

		switch key {
		// Current connections
		case "THREADS_CONNECTED":
			connections, _ := strconv.Atoi(value)
			status.Metrics.Connections = connections
		// Uptime
		case "UPTIME":
			uptime, _ := strconv.Atoi(value)
			status.Metrics.Uptime = uptime
		// Queries per second
		case "QUERIES":
			queries, _ := strconv.Atoi(value)

			// If we don't have a previous value for the total queries
			// then qps is technically 0 as we don't know it yet
			if status.Metrics.Queries == 0 {
				status.Metrics.QueriesPerSecond = 0
				status.Metrics.Queries = queries
			// Otherwise the value of qps is the diff between the current
			// and previous count of queries
			} else {
				diff := queries - status.Metrics.Queries
				
				fmt.Println(diff)

				// qps can never be below 0..
				if diff > 0 {
					status.Metrics.QueriesPerSecond = diff
				} else {
					status.Metrics.QueriesPerSecond = 0
				}

				status.Metrics.Queries = queries
			}
		}
	}

	// Check for any remaining errors
	err = rows.Err()

	if err != nil {
		return status, err
	}

	// Fetch all the db metrics
	rows, err = conn.Query("SHOW GLOBAL VARIABLES")

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
