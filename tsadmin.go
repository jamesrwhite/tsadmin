// tsadmin
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jamesrwhite/tsadmin/config"
	"github.com/jamesrwhite/tsadmin/database"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
)

var ticker = time.NewTicker(time.Second * 1)
var statuses map[string]*database.DatabaseStatus

func main() {
	// Create an instance of our app
	app := negroni.New()

	// Create a new router
	router := httprouter.New()

	// Fetch the initial statuses of the databases with 2 seconds of data
	statuses, _ = monitor()
	time.Sleep(time.Second * 1)
	statuses, _ = monitor()

	// Then refresh the statuses once a second
	go func() {
		for range ticker.C {
			statuses, _ = monitor()
		}
	}()

	// Add our routes
	router.GET("/status.json", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		// JSON please
		w.Header().Set("Content-Type", "application/json")

		// Define our response array
		response := []*database.DatabaseStatus{}

		// Reformat the statuses map as a simple array for JSON
		for _, status := range statuses {
			response = append(response, status)
		}

		// Encode the response
		jsonResponse, _ := json.Marshal(response)

		fmt.Fprint(w, string(jsonResponse))
	})

	// Set the router to use
	app.UseHandler(router)

	// Start our app on $PORT
	app.Run(":" + os.Getenv("PORT"))
}

func monitor() (map[string]*database.DatabaseStatus, error) {
	// Load the config on each request in case it gets updated
	tsConfig, err := config.Load("config/config.json")

	if err != nil {
		log.Fatal(err)
	}

	// Define our response map
	updatedStatuses := make(map[string]*database.DatabaseStatus)

	// Loop each database and fetch the status
	for _, dbConfig := range tsConfig.Databases {
		// Get the database status, here we pass the last known status
		// so we can determine metrics like queries per second
		status, err := database.Status(dbConfig, statuses[dbConfig.String()])

		if err != nil {
			log.Fatal(err)
		}

		updatedStatuses[dbConfig.String()] = status
	}

	return updatedStatuses, nil
}
