package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"fmt"
	"strings"
)

var db *(sql.DB)

func setupDatabase() (err error) {
	//Create global db handle
	if db, err = sql.Open("postgres", os.Getenv("DATABASE_URL")); err != nil {
		return
	}

	//Create needed tables
	if err = setupRequiredTables(); err != nil {
		return
	}
	return
}

func setupRequiredTables() (err error){
	var flightsTableAlreadyExist bool
	if flightsTableAlreadyExist, err = setupTable(FLIGHTS_72HR_TABLE, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %v (
		Origin VARCHAR(50),
		Destination VARCHAR(50),
		RollCall TIMESTAMP,
		SeatCount INT,
		SeatType VARCHAR(3), 
		Cancelled BOOLEAN,
		PhotoSource VARCHAR(2048),
		CONSTRAINT flights_pk PRIMARY KEY (Origin, Destination, RollCall));
		`, FLIGHTS_72HR_TABLE)); err != nil {
		return
	}
	if (flightsTableAlreadyExist) {
		log.Println(FLIGHTS_72HR_TABLE + " table already exists.")
	} else {
		log.Println(FLIGHTS_72HR_TABLE + " table created.")
	}

	var locationsAlreadyExist bool
	if locationsAlreadyExist, err = setupTable(LOCATIONS_TABLE, fmt.Sprintf(`
		CREATE TABLE %v (
		Title VARCHAR(50),
		URL VARCHAR(2048),
		CONSTRAINT locations_pk PRIMARY KEY (Title));
		`, LOCATIONS_TABLE)); err != nil {
		return
	} 
	if (locationsAlreadyExist) {
		log.Println(LOCATIONS_TABLE + " table already exists.")
	} else {
		log.Println(LOCATIONS_TABLE + " table created.")
		if err = populateLocationsTable(); err != nil {
			return
		}
	}
	return
}

func setupTable(tableName string, query string) (tableAlreadyExist bool, err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return 
	}

	//Check if table exists
	if err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1);", tableName).Scan(&tableAlreadyExist); err != nil {
		return
	}

	//If table does not exist, run query to create table
	if !tableAlreadyExist {
		if _, err = db.Exec(query); err != nil {
			return
		}
	}

	return
}

func populateLocationsTable() (err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}

	//Read in location keyword file
	var locationKeywordsArray []LocationKeywords
	if locationKeywordsArray, err = readLocationKeywordsFileToArray(TERMINAL_KEYWORDS_FILE); err != nil {
		return
	}

	//Insert locations into table
	var rowsAffected int64
	fmt.Printf("Inserting %v locations into %v table...", len(locationKeywordsArray), LOCATIONS_TABLE)

	//spawn go routine to continuously read and run functions in the channel
	var printChannel chan func()
	printChannel = make(chan func())
	go func() {
		for true {
			tmp := <-printChannel
			tmp()
		}
	}()

	defer func() {
		fmt.Printf("\rInserted %v locations into %v table.\n", rowsAffected, LOCATIONS_TABLE)
		}()
	for i, lk := range locationKeywordsArray {
		var result sql.Result
		if result, err = db.Exec(fmt.Sprintf(`
			INSERT INTO %v (Title, URL) 
	    	VALUES ($1, $2);
	    	`, LOCATIONS_TABLE), lk.Title, nil); err != nil {
			return
		}

		var affected int64
		if affected, err = result.RowsAffected(); err != nil {
			return
		}
		rowsAffected += affected

		printChannel <- func() {
			fmt.Printf("\r\u001b[0KInserted %v locations into %v table%v", rowsAffected, LOCATIONS_TABLE, strings.Repeat(".", i % 4))
		}

	}
	return
}