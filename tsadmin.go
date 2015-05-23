// tsadmin
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jamesrwhite/tsadmin/config"
	"github.com/jamesrwhite/tsadmin/database"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
)

type Response struct {
	Metadata  map[string]string `json:"metadata"`
	Metrics   map[string]string `json:"metrics"`
	Variables map[string]string `json:"variables"`
}

func main() {
	// Create an instance of our app
	app := negroni.Classic()

	// Create a new router
	router := httprouter.New()

	// Add our routes
	router.GET("/status.json", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Load the config on each request in case it gets updated
		tsConfig, err := config.Load("config/config.json")

		if err != nil {
			log.Fatal(err)
		}

		// Define our response map
		response := []*Response{}

		for _, dbConfig := range tsConfig.Databases {
			// Get a database instance
			db, err := database.New(dbConfig)

			// Check that we could connect correctly
			if err != nil {
				log.Fatal(err)
			}

			// Get the database status
			status, err := database.Status(db)

			if err != nil {
				log.Fatal(err)
			}

			response = append(response, &Response{
				Metadata: map[string]string{
					"name": dbConfig.Name,
					"host": dbConfig.Host,
					"port": fmt.Sprintf("%v", dbConfig.Port),
				},
				Metrics:   status.Metrics,
				Variables: status.Variables,
			})

			db.Close()
		}

		responseJson, _ := json.Marshal(response)

		// JSON please
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, string(responseJson))
	})

	// Set the router to use
	app.UseHandler(router)

	// Start our app on $PORT
	app.Run(":" + os.Getenv("PORT"))
}
