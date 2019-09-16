package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/r4wm/mintz5/db"
	"github.com/r4wm/mintz5/kjv"
	"github.com/r4wm/sqlite3_kjv"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Database
	dbPath := "/tmp/kjv.db"

	// Create db if does not exist
	path, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		_, err := sqlite3_kjv.CreateKJVDB(dbPath)

		if err != nil {
			panic(err)
		}

		log.Infof("Created database %v", path)
	}

	// Create database connection
	db, err := db.CreateDatabase(dbPath)
	if err != nil {
		panic(err)
	}

	log.Infof("Database connection OK.")

	// Router
	router := mux.NewRouter().StrictSlash(false)

	app := kjv.App{
		Router:   router,
		Database: db,
	}
	app.SetupRouter()
	log.Infof("Setup router OK.")

	// Serve
	log.Fatal(http.ListenAndServe(":8000", router))
}
