// tsadmin/database
package database

import (
	"database/sql"
	"fmt"
	"log"
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
	CurrentConnections          int `json:"current_connections"`
	Connections                 int `json:"connections"`
	ConnectionsPerSecond        int `json:"connections_per_second"`
	AbortedConnections          int `json:"aborted_connections"`
	AbortedConnectionsPerSecond int `json:"aborted_connections_per_second"`
	Queries                     int `json:"queries"`
	QueriesPerSecond            int `json:"queries_per_second"`
	Reads                       int `json:"reads"`
	ReadsPerSecond              int `json:"reads_per_second"`
	Writes                      int `json:"writes"`
	WritesPerSecond             int `json:"writes_per_second"`
	Uptime                      int `json:"uptime"`
}

type DatabaseVariables struct {
	MaxConnections int `json:"max_connections"`
}

func (db *Database) String() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", db.User, db.Password, db.Host, db.Port)
}

func Status(db Database, previous *DatabaseStatus) (*DatabaseStatus, error) {
	status := &DatabaseStatus{
		Metadata: DatabaseMetadata{
			Name: db.Name,
			Host: db.Host,
			Port: db.Port,
		},
		Metrics:   DatabaseMetrics{},
		Variables: DatabaseVariables{},
	}

	// Fetch the metrics
	status, _ = execQuery(db, "metrics", previous, status)

	// Fetch the variables
	status, _ = execQuery(db, "variables", previous, status)

	return status, nil
}

// Execute a query on the given database for looking up metrics/variables
func execQuery(db Database, queryType string, previous *DatabaseStatus, status *DatabaseStatus) (*DatabaseStatus, error) {
	var (
		key   string
		value string
		table string
	)

	// Fetch all the db metrics/variables
	if queryType == "metrics" {
		table = "GLOBAL_STATUS"
	} else if queryType == "variables" {
		table = "GLOBAL_VARIABLES"
	} else {
		log.Fatal("Unknown queryType")
	}

	// Connect to the database
	conn, _ := sql.Open("mysql", db.String())
	defer conn.Close()

	// Fetch all the db metrics
	rows, err := conn.Query(fmt.Sprintf("SELECT VARIABLE_NAME AS 'key', VARIABLE_VALUE AS 'value' FROM %s", table))

	// Handle query errors
	if err != nil {
		return status, err
	}

	defer rows.Close()

	// Loop each metric/variable in the server status
	for rows.Next() {
		err := rows.Scan(&key, &value)

		// Handle row reading errors
		if err != nil {
			return status, err
		}

		// Process the metrics/variables
		if queryType == "metrics" {
			status, _ = processMetrics(previous, status, key, value)
		} else {
			status, _ = processVariables(status, key, value)
		}
	}

	// Check for any remaining errors
	err = rows.Err()

	return status, err
}

// Process metrics returned from the GLOBAL_STATUS table
func processMetrics(previous *DatabaseStatus, status *DatabaseStatus, key string, value string) (*DatabaseStatus, error) {
	var (
		err                error
		currentConnections int
		connections        int
		diff               int
		abortedConnections int
		queries            int
		// reads              int
		writes             int
		uptime             int
	)

	switch key {
	// Current connections
	case "THREADS_CONNECTED":
		currentConnections, err = strconv.Atoi(value)
		status.Metrics.CurrentConnections = currentConnections
	// Connections per second
	case "CONNECTIONS":
		connections, err = strconv.Atoi(value)

		// If we don't have a previous value for the total connections
		// then cps is technically 0 as we don't know it yet
		if previous == nil || previous.Metrics.Connections == 0 {
			status.Metrics.ConnectionsPerSecond = 0
			status.Metrics.Connections = connections
		// Otherwise the value of cps is the diff between the current
		// and previous count of connections
		} else {
			diff = connections - previous.Metrics.Connections

			// qps can never be below 0..
			if diff > 0 {
				status.Metrics.ConnectionsPerSecond = diff
			} else {
				status.Metrics.ConnectionsPerSecond = 0
			}

			status.Metrics.Connections = connections
		}
	// Aborted connections per second
	case "ABORTED_CONNECTS":
		abortedConnections, err = strconv.Atoi(value)

		// If we don't have a previous value for the total aborted connections
		// then acps is technically 0 as we don't know it yet
		if previous == nil || previous.Metrics.AbortedConnections == 0 {
			status.Metrics.AbortedConnectionsPerSecond = 0
			status.Metrics.AbortedConnections = abortedConnections
		// Otherwise the value of acps is the diff between the current
		// and previous count of connections
		} else {
			diff = abortedConnections - previous.Metrics.AbortedConnections

			// qps can never be below 0..
			if diff > 0 {
				status.Metrics.AbortedConnectionsPerSecond = diff
			} else {
				status.Metrics.AbortedConnectionsPerSecond = 0
			}

			status.Metrics.AbortedConnections = abortedConnections
		}
	// Queries per second
	case "QUERIES":
		queries, err = strconv.Atoi(value)

		// If we don't have a previous value for the total queries
		// then qps is technically 0 as we don't know it yet
		if previous == nil || previous.Metrics.Queries == 0 {
			status.Metrics.QueriesPerSecond = 0
			status.Metrics.Queries = queries
		// Otherwise the value of qps is the diff between the current
		// and previous count of queries
		} else {
			diff = queries - previous.Metrics.Queries

			// qps can never be below 0..
			if diff > 0 {
				status.Metrics.QueriesPerSecond = diff
			} else {
				status.Metrics.QueriesPerSecond = 0
			}

			status.Metrics.Queries = queries
		}
	// Writes per second
	case "COM_DELETE", "COM_INSERT", "COM_UPDATE", "COM_REPLACE", "COM_INSERT_SELECT", "COM_REPLACE_SELECT":
		writes, err = strconv.Atoi(value)

		// If we don't have a previous value for the total reads
		// then wps is technically 0 as we don't know it yet
		if previous == nil || previous.Metrics.Writes == 0 {
			status.Metrics.WritesPerSecond = 0
			status.Metrics.Writes = writes
		// Otherwise the value of wps is the diff between the current
		// and previous count of reads
		} else {
			diff = writes - previous.Metrics.Writes

			fmt.Println(fmt.Sprintf("[%s] %s, diff: %v", key, value, diff))

			// wps can never be below 0..
			if diff > 0 {
				status.Metrics.WritesPerSecond += diff
			}

			status.Metrics.Writes += writes
		}
	// Uptime
	case "UPTIME":
		uptime, err = strconv.Atoi(value)
		status.Metrics.Uptime = uptime
	}

	if err != nil {
		return status, err
	} else {
		return status, nil
	}
}

// Process variables returned from the GLOBAL_VARIABLES table
func processVariables(status *DatabaseStatus, key string, value string) (*DatabaseStatus, error) {
	var (
		err            error
		maxConnections int
	)

	// Max allowed connections
	if key == "MAX_CONNECTIONS" {
		maxConnections, err = strconv.Atoi(value)
		status.Variables.MaxConnections = maxConnections
	}

	if err != nil {
		return status, err
	} else {
		return status, nil
	}
}
