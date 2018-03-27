package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"fmt"
	"time"
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
	var locationKeywordsArray []Terminal
	if locationKeywordsArray, err = readTerminalArrayFromFiles(TERMINAL_FILE, LOCATION_KEYWORDS_FILE); err != nil { //same files reread when building fuzzy models in ocr-fuzzy.go. Maybe pass data in future.
		return
	}

	//Insert locations into table
	var rowsAffected int64
	fmt.Println("Inserting %v locations into %v table...", len(locationKeywordsArray), LOCATIONS_TABLE)

	//spawn go routine to continuously read and run functions in the channel
	//Stop reading from channel when read function returns false boolean
	//https://golang.org/doc/codewalk/functions/
	type printFunc func() bool
	var printChannel chan printFunc
	printChannel = make(chan printFunc)
	go func() {
		for true {
			tmp := <-printChannel
			if !tmp() {
				break
			}
		}
	}()

	defer func() {
		log.Printf("\r\u001b[1A\u001b[0KInserted %v locations into %v table.\n", rowsAffected, LOCATIONS_TABLE)
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

		printChannel <- func() bool {
			fmt.Printf("\r\u001b[1A\u001b[0KInserted %v/%v locations into %v table%v\n", rowsAffected, len(locationKeywordsArray), LOCATIONS_TABLE, strings.Repeat(".", i % 10))
			return true
		}
	}
	

	printChannel <- func() bool {
		return false
	}
	return
}

//Insert []Flight into table.
func insertFlightsIntoTable(table string, flights []Flight) (err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}

	//Insert flights into table
	var rowsAffected int64
	for _, flight := range flights {
		var result sql.Result
		if result, err = db.Exec(fmt.Sprintf(`
			INSERT INTO %v (Origin, Destination, RollCall, SeatCount, SeatType, Cancelled, PhotoSource) 
	    	VALUES ($1, $2, $3, $4, $5, $6, $7);
	 		`, table), flight.Origin, flight.Destination, flight.RollCall.Format("2006-01-02 15:04:05"), flight.SeatCount, flight.SeatType, false, flight.PhotoSource); err != nil {
			return
		} 

		var affected int64
		if affected, err = result.RowsAffected(); err != nil {
			return
		}
		rowsAffected += affected
	}
	
	fmt.Printf("INSERT []Flights to %v len %v\n%v rows affected\n", table, len(flights), rowsAffected)

	return
}

//Pass in time in local TZ
func deleteFlightsFromTableForDayForOrigin(table string, targetDay time.Time, origin string) (err error) {

	dateEqual := func(date1, date2 time.Time) bool {
	    y1, m1, d1 := date1.Date()
	    y2, m2, d2 := date2.Date()
	    return y1 == y2 && m1 == m2 && d1 == d2
	}

	targetDay = targetDay.UTC()

	var start, end time.Time
	currentTime := time.Now()
	if dateEqual(targetDay, currentTime) {
		start = currentTime
		end = currentTime.Round(time.Hour*24)
	} else if targetDay.After(currentTime) {
		year, month, day := targetDay.Date()
		start = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		end = start.Add(time.Hour * 24)
	}

	err = deleteFlightsFromTableBetweenTimesForOrigin(table, start, end, origin)

	return
}

//DELETE Inclusive of start, exclusive of end. Can input 0000 of start date and 0000 end date and get all times up but no including 0000 of end date.
func deleteFlightsFromTableBetweenTimesForOrigin(table string, start time.Time, end time.Time, origin string) (err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}

	var result sql.Result
	if result, err = db.Exec(fmt.Sprintf(`
		DELETE FROM %v 
 		WHERE Origin=$1 AND RollCall >= $2 AND RollCall < $3;
 		`, table), origin, start.Format("2006-01-02"), end.Format("2006-01-02")); err != nil {
		return
	} 
	var affected int64
	if affected, err = result.RowsAffected(); err != nil {
		return
	}

	fmt.Printf("Delete flights between times for origin %v %v %v %v\n%v rows affected\n", table, start, end, origin, affected)

	return
}
