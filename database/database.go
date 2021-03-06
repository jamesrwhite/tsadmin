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
	ConnectionsPerSecond        int `json:"connections_per_second"`
	AbortedConnectionsPerSecond int `json:"aborted_connections_per_second"`
	QueriesPerSecond            int `json:"queries_per_second"`
	ReadsPerSecond              int `json:"reads_per_second"`
	WritesPerSecond             int `json:"writes_per_second"`
	Uptime                      int `json:"uptime"`
	connections                 int
	abortedConnections          int
	queries                     int
	reads                       int
	writes                      int
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
	err := execQuery(db, "metrics", previous, status)

	if err != nil {
		return nil, err
	}

	// Fetch the variables
	err = execQuery(db, "variables", previous, status)

	if err != nil {
		return nil, err
	}

	return status, nil
}

// Execute a query on the given database for looking up metrics/variables
func execQuery(db Database, queryType string, previous *DatabaseStatus, status *DatabaseStatus) error {
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
	conn, err := sql.Open("mysql", db.String())

	if err != nil {
		return err
	}

	defer conn.Close()

	// Put MySQL in 5.6 compatability mode as the location of some of the metrics has chagned in 5.7
	_, err = conn.Query("SET GLOBAL show_compatibility_56 = ON")

	if err != nil {
		return err
	}

	// Fetch all the db metrics
	rows, err := conn.Query(fmt.Sprintf("SELECT VARIABLE_NAME AS 'key', VARIABLE_VALUE AS 'value' FROM %s", table))

	// Handle query errors
	if err != nil {
		return err
	}

	defer rows.Close()

	// Loop each metric/variable in the server status
	for rows.Next() {
		err := rows.Scan(&key, &value)

		// Handle row reading errors
		if err != nil {
			return err
		}

		// Process the metrics/variables
		if queryType == "metrics" {
			err = processMetric(previous, status, key, value)
		} else {
			err = processVariable(status, key, value)
		}

		if err != nil {
			return err
		}
	}

	// Do some final processing of the metrics
	err = postProcessMetrics(previous, status)

	if err != nil {
		return err
	}

	// Check for any remaining errors
	err = rows.Err()

	return err
}

// Process metric returned from the GLOBAL_STATUS table
func processMetric(previous *DatabaseStatus, status *DatabaseStatus, key string, value string) error {
	var (
		err                error
		currentConnections int
		connections        int
		diff               int
		abortedConnections int
		queries            int
		uptime             int
		readWriteValue     int
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
		if previous == nil || previous.Metrics.connections == 0 {
			status.Metrics.ConnectionsPerSecond = 0
			status.Metrics.connections = connections
		// Otherwise the value of cps is the diff between the current
		// and previous count of connections
		} else {
			diff = connections - previous.Metrics.connections

			// cps can never be below 0..
			if diff > 0 {
				status.Metrics.ConnectionsPerSecond = diff
			} else {
				status.Metrics.ConnectionsPerSecond = 0
			}

			status.Metrics.connections = connections
		}
	// Aborted connections per second
	case "ABORTED_CONNECTS":
		abortedConnections, err = strconv.Atoi(value)

		// If we don't have a previous value for the total aborted connections
		// then acps is technically 0 as we don't know it yet
		if previous == nil || previous.Metrics.abortedConnections == 0 {
			status.Metrics.AbortedConnectionsPerSecond = 0
			status.Metrics.abortedConnections = abortedConnections
		// Otherwise the value of acps is the diff between the current
		// and previous count of connections
		} else {
			diff = abortedConnections - previous.Metrics.abortedConnections

			// acps can never be below 0..
			if diff > 0 {
				status.Metrics.AbortedConnectionsPerSecond = diff
			} else {
				status.Metrics.AbortedConnectionsPerSecond = 0
			}

			status.Metrics.abortedConnections = abortedConnections
		}
	// Queries per second
	case "QUERIES":
		queries, err = strconv.Atoi(value)

		// If we don't have a previous value for the total queries
		// then qps is technically 0 as we don't know it yet
		if previous == nil || previous.Metrics.queries == 0 {
			status.Metrics.QueriesPerSecond = 0
			status.Metrics.queries = queries
		// Otherwise the value of qps is the diff between the current
		// and previous count of queries
		} else {
			diff = queries - previous.Metrics.queries

			// qps can never be below 0..
			if diff > 0 {
				status.Metrics.QueriesPerSecond = diff
			} else {
				status.Metrics.QueriesPerSecond = 0
			}

			status.Metrics.queries = queries
		}
	// Read/Writes per second
	case "COM_SELECT", "COM_INSERT_SELECT", "COM_REPLACE_SELECT", "COM_DELETE", "COM_INSERT", "COM_UPDATE", "COM_REPLACE":
		readWriteValue, err = strconv.Atoi(value)

		// Reads
		if key == "COM_SELECT" || key == "COM_INSERT_SELECT" || key == "COM_REPLACE_SELECT" {
			status.Metrics.reads += readWriteValue

			// Reads/Writes
			if key == "COM_INSERT_SELECT" || key == "COM_REPLACE_SELECT" {
				status.Metrics.writes += readWriteValue
			}
		// Writes
		} else {
			status.Metrics.writes += readWriteValue
		}
	// Uptime
	case "UPTIME":
		uptime, err = strconv.Atoi(value)
		status.Metrics.Uptime = uptime
	}

	if err != nil {
		return err
	} else {
		return nil
	}
}

// Process variables returned from the GLOBAL_VARIABLES table
func processVariable(status *DatabaseStatus, key string, value string) error {
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
		return err
	}

	return nil
}

// Post processing of metrics
func postProcessMetrics(previous *DatabaseStatus, status *DatabaseStatus) error {
	var diff int

	// If we don't have a previous value for the total reads
	// then rps is technically 0 as we don't know it yet
	if previous != nil {
		// Calculate the RPS
		diff = status.Metrics.reads - previous.Metrics.reads

		// rps can never be below 0..
		if diff > 0 {
			status.Metrics.ReadsPerSecond = diff
		} else {
			status.Metrics.ReadsPerSecond = 0
		}

		// Calculate the WPS
		diff = status.Metrics.writes - previous.Metrics.writes

		// wps can never be below 0..
		if diff > 0 {
			status.Metrics.WritesPerSecond = diff
		} else {
			status.Metrics.WritesPerSecond = 0
		}
	}

	return nil
}
