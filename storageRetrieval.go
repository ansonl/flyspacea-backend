package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
	"time"
)

var db *(sql.DB)

func connectDatabase() (err error) {
	//Create global db handle
	if db, err = sql.Open("postgres", os.Getenv("DATABASE_URL")); err != nil {
		return
	}
	return
}

func createDatabaseTables() (err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}

	//Create needed tables
	if err = createRequiredTables(); err != nil {
		return
	}
	return
}

func createRequiredTables() (err error) {
	var locationsAlreadyExist bool
	if locationsAlreadyExist, err = setupTable(LOCATIONS_TABLE, fmt.Sprintf(`
		CREATE TABLE %v (
		Title VARCHAR(100),
		Phone VARCHAR(50),
		Email VARCHAR(255),
		GeneralInfo VARCHAR(2048),
		FBId VARCHAR(255),
		URL VARCHAR(2048),
		CONSTRAINT locations_pk PRIMARY KEY (Title));
		`, LOCATIONS_TABLE)); err != nil {
		return
	}
	if locationsAlreadyExist {
		//log.Println(LOCATIONS_TABLE + " table already exists.")
	} else {
		log.Println(LOCATIONS_TABLE + " table created.")
	}

	var flightsTableAlreadyExist bool
	if flightsTableAlreadyExist, err = setupTable(FLIGHTS_72HR_TABLE, fmt.Sprintf(`
			CREATE TABLE %v (
			Origin VARCHAR(100),
			Destination VARCHAR(100),
			RollCall TIMESTAMP NULL,
			UnknownRollCallDate BOOLEAN,
			SeatCount INT,
			SeatType VARCHAR(3), 
			Cancelled BOOLEAN,
			PhotoSource VARCHAR(2048),
			SourceDate TIMESTAMP,
			CONSTRAINT flights_pk PRIMARY KEY (Origin, Destination, RollCall, PhotoSource),
			CONSTRAINT flights_origin_fk FOREIGN KEY (Origin) REFERENCES Locations(Title),
			CONSTRAINT flights_dest_fk FOREIGN KEY (Destination) REFERENCES Locations(Title));
		`, FLIGHTS_72HR_TABLE)); err != nil {
		return
	}
	if flightsTableAlreadyExist {
		//log.Println(FLIGHTS_72HR_TABLE + " table already exists.")
	} else {
		log.Println(FLIGHTS_72HR_TABLE + " table created.")
	}

	/*
		//Delete indexes
		if _, err = db.Exec(fmt.Sprintf(`
			DROP INDEX IF EXISTS %v;
			DROP INDEX IF EXISTS %v;
			DROP INDEX IF EXISTS %v;
			DROP INDEX IF EXISTS %v;
			`,
			FLIGHTS_72HR_TABLE_INDEX_ORIGIN_DEST_RC,
			FLIGHTS_72HR_TABLE_INDEX_ORIGIN_RC,
			FLIGHTS_72HR_TABLE_INDEX_DEST_RC,
			FLIGHTS_72HR_TABLE_INDEX_RC)); err != nil {

		}
	*/

	if _, err = db.Exec(fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %v ON %v (
			Origin ASC,
			Destination ASC,
			RollCall DESC
		);
		CREATE INDEX IF NOT EXISTS %v ON %v (
			Origin ASC,
			RollCall DESC
		);
		CREATE INDEX IF NOT EXISTS %v ON %v (
			Destination ASC,
			RollCall DESC
		);
		CREATE INDEX IF NOT EXISTS %v ON %v (
			RollCall DESC
		);
		`,
		FLIGHTS_72HR_TABLE_INDEX_ORIGIN_DEST_RC, FLIGHTS_72HR_TABLE,
		FLIGHTS_72HR_TABLE_INDEX_ORIGIN_RC, FLIGHTS_72HR_TABLE,
		FLIGHTS_72HR_TABLE_INDEX_DEST_RC, FLIGHTS_72HR_TABLE,
		FLIGHTS_72HR_TABLE_INDEX_RC, FLIGHTS_72HR_TABLE)); err != nil {
		return
	} else {
		//log.Println(FLIGHTS_72HR_TABLE + " indexes created.")
	}

	return
}

//Create table if table does not exist
func setupTable(tableName string, query string) (tableAlreadyExist bool, err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}

	//Check if table exists
	if tableAlreadyExist, err = doesTableExist(tableName); err != nil {
		return
	}

	//fmt.Println(tableName, " exist", tableAlreadyExist)

	//If table does not exist, run query to create table
	if !tableAlreadyExist {
		if _, err = db.Exec(query); err != nil {
			fmt.Println(query)
			return
		}
	}

	return
}

//Check if specified tableName exists as table in database
func doesTableExist(tableName string) (tableExist bool, err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}

	//Check if table exists
	if err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1);", tableName).Scan(&tableExist); err != nil {
		return
	}
	return
}

//INSERT locations into locations table
func populateLocationsTable(terminalArray []Terminal) (err error) {
	var locationKeywordsArray []Terminal
	if locationKeywordsArray, err = readTerminalArrayFromFiles(LOCATION_KEYWORDS_FILE); err != nil { //same files reread when building fuzzy models in ocr-fuzzy.go. Maybe pass data in future.
		return
	}

	//If debug single terminal, still load all terminals and locations into database.
	if DEBUG_SINGLE_FILE {
		if locationKeywordsArray, err = readTerminalArrayFromFiles(LOCATION_KEYWORDS_FILE, TERMINAL_FILE); err != nil {
			return
		}
	}
	//Testing
	//locationKeywordsArray = nil

	for _, v := range terminalArray {
		locationKeywordsArray = append(locationKeywordsArray, v)
	}

	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}
	//Insert locations into table
	var rowsAffected int64
	fmt.Printf("Inserting %v locations into %v table...", len(locationKeywordsArray), LOCATIONS_TABLE)

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

		var insertPhone sql.NullString
		if len(lk.Phone) > 0 {
			insertPhone.String = lk.Phone
			insertPhone.Valid = true
		}

		var insertEmail sql.NullString
		if len(lk.Emails) > 0 && len(lk.Emails[0]) > 0 {
			insertEmail.String = lk.Emails[0]
			insertEmail.Valid = true
		}

		if result, err = db.Exec(fmt.Sprintf(`
			INSERT INTO %v (Title, Phone, Email, GeneralInfo, FBId, URL) 
	    	VALUES ($1, $2, $3, $4, $5, $6) 
	    	ON CONFLICT (Title) DO UPDATE SET
	    	Phone = EXCLUDED.Phone,
	    	Email = EXCLUDED.Email,
	    	GeneralInfo = EXCLUDED.GeneralInfo,
	    	FBId = EXCLUDED.FBId;
	    	`, LOCATIONS_TABLE), lk.Title, insertPhone, insertEmail, lk.GeneralInfo, lk.Id, nil); err != nil {
			return
		}

		var affected int64
		if affected, err = result.RowsAffected(); err != nil {
			return
		}
		rowsAffected += affected

		printChannel <- func() bool {
			fmt.Printf("\r\u001b[1A\u001b[0KInserted %v/%v locations into %v table%v\n", rowsAffected, len(locationKeywordsArray), LOCATIONS_TABLE, strings.Repeat(".", i%10))
			return true
		}
	}

	printChannel <- func() bool {
		return false
	}
	return
}

/* 
 * SELECT DISTINCT locations from table from all origins and destinations. Return a distinct list of locations from all origins and destinations stored. 
 *
 */
func selectAllLocationsFromTable(table string) (distinctLocations []string, err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}

	var locationRows *sql.Rows

	if locationRows, err = db.Query(fmt.Sprintf(`
		(SELECT DISTINCT origin AS location FROM %v 
			UNION 
			SELECT DISTINCT destination AS location FROM hr72_flights) 
			ORDER BY location ASC;
		`, table)); err != nil {
		return
	}

	var countOfRows = 0
	for locationRows.Next() {
		var tmp string

		if err = locationRows.Scan(&tmp); err != nil {
			return
		}

		distinctLocations = append(distinctLocations, tmp)
		countOfRows++
	}
	locationRows.Close()

	fmt.Printf("SELECT DISTINCT location list\n%v rows selected.\n", countOfRows)
	return
}

/*
//SELECT flights from table.
 * Parameters
 * table SQL DB table name
 * origin Location title (optional)
 * dest Location title (optional)
 * start Time to search from (inclusive)
 * duration Duration to add to start Time (exclusive) 1Day=time.Hour*24
*/
func selectFlightsFromTableWithOriginDestTimeDuration(table string, origin string, dest string, start time.Time, duration time.Duration) (flights []Flight, err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}

	var flightRows *sql.Rows
	//Determine which query to use
	if len(origin) > 0 && len(dest) == 0 { //Search by only Origin
		if flightRows, err = db.Query(fmt.Sprintf(`
			SELECT Origin, Destination, RollCall, UnknownRollCallDate, SeatCount, SeatType, Cancelled, PhotoSource, SourceDate
			FROM %v
			WHERE Origin=$1 AND RollCall >= $2 AND RollCall < $3
			ORDER BY RollCall, Origin, Destination, SeatCount, SeatType, SourceDate;
 		`, table), origin, start, start.Add(duration)); err != nil {
			return
		}
	} else if len(origin) == 0 && len(dest) > 0 { //Search by only Destination
		if flightRows, err = db.Query(fmt.Sprintf(`
			SELECT Origin, Destination, RollCall, UnknownRollCallDate, SeatCount, SeatType, Cancelled, PhotoSource, SourceDate
			FROM %v
			WHERE Destination=$1 AND RollCall >= $2 AND RollCall < $3
			ORDER BY RollCall, Origin, Destination, SeatCount, SeatType, SourceDate;
 		`, table), dest, start, start.Add(duration)); err != nil {
			return
		}
	} else if len(origin) > 0 && len(dest) > 0 { //Search by Origin and Destination
		if flightRows, err = db.Query(fmt.Sprintf(`
			SELECT Origin, Destination, RollCall, UnknownRollCallDate, SeatCount, SeatType, Cancelled, PhotoSource, SourceDate
			FROM %v
			WHERE Origin=$1 AND Destination=$2 AND RollCall >= $3 AND RollCall < $4
			ORDER BY RollCall, Origin, Destination, SeatCount, SeatType, SourceDate;
 		`, table), origin, dest, start, start.Add(duration)); err != nil {
			return
		}
	} else { //Search all in time duration
		if flightRows, err = db.Query(fmt.Sprintf(`
			SELECT Origin, Destination, RollCall, UnknownRollCallDate, SeatCount, SeatType, Cancelled, PhotoSource, SourceDate
			FROM %v
			WHERE RollCall >= $1 AND RollCall < $2
			ORDER BY RollCall, Origin, Destination, SeatCount, SeatType, SourceDate;
 		`, table), start, start.Add(duration).Format("2006-01-02")); err != nil {
			return
		}
	}

	var countOfRows = 0
	for flightRows.Next() {
		var flight Flight

		if err = flightRows.Scan(&flight.Origin, &flight.Destination, &flight.RollCall, &flight.UnknownRollCallDate, &flight.SeatCount, &flight.SeatType, &flight.Cancelled, &flight.PhotoSource, &flight.SourceDate); err != nil {
			return
		}

		flights = append(flights, flight)
		countOfRows++
	}
	flightRows.Close()

	fmt.Printf("SELECT flights %v between origin %v dest %v times %v %v\n%v rows selected.\n", table, origin, dest, start, start.Add(duration), countOfRows)

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
			INSERT INTO %v (Origin, Destination, RollCall, UnknownRollCallDate, SeatCount, SeatType, Cancelled, PhotoSource, SourceDate) 
	    	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	 		`, table), flight.Origin, flight.Destination, flight.RollCall.In(time.UTC), flight.UnknownRollCallDate, flight.SeatCount, flight.SeatType, false, flight.PhotoSource, flight.SourceDate.In(time.UTC)); err != nil {
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

//Delete flights for time.Time with origin from a originTerminal. Accounts for current time to avoid deleting past flights.
//Pass in time in origin Terminal TZ
//Expect targetDay to be 00:00:00 time.
func deleteFlightsFromTableForDayForOriginTerminal(table string, targetDay time.Time, originTerminal Terminal) (err error) {

	dateEqual := func(date1, date2 time.Time) bool {
		y1, m1, d1 := date1.Date()
		y2, m2, d2 := date2.Date()
		return y1 == y2 && m1 == m2 && d1 == d2
	}

	//Truncate function rounds to UTC 00:00:00, does not account for TZ
	//Truncate 24hr equivalent snippet that works correctly with all TZ
	truncateDay := func(input time.Time) (output time.Time) {
		year, month, day := input.Date()
		output = time.Date(year, month, day, 0, 0, 0, 0, input.Location())
		return
	}

	//Determine whether DELETE timeframe is in the future. We do not want to delete past flights because Terminals' Facebook pages may post updated slides throughout the day that omit fulfilled flights. Safe assumption is to only delete future flights on current and future days.
	//If DELETE targetDay is the same 24hr date as current time, we only delete flights between time.Now().In(TERMINAL_LOCAL_TZ) and end of day (2359).
	//If DELETE targetDay is in future of currentTime, delete all origin flights for targetDay. Expect targetDay to be 00:00:00 time so it will resolve to a future date.
	//if DELETE targetDay is in past of currentTime, do not delete anything.
	var start, end time.Time
	currentTime := time.Now().In(originTerminal.Timezone)
	if dateEqual(targetDay, currentTime) { //DELETE in current date
		start = currentTime
		end = truncateDay(currentTime).Add(time.Hour * 24)
	} else if targetDay.After(currentTime) { //DELETE in future
		start = truncateDay(targetDay)
		end = start.Add(time.Hour * 24)
	} else { //DELETE in past.
		//Do not delete anything.
		log.Printf("Not deleting past flights for %v for date %v\n", originTerminal.Title, targetDay)
		return
	}

	err = deleteFlightsFromTableBetweenTimesForOrigin(table, start.UTC(), end.UTC(), originTerminal.Title)

	return
}

//DELETE Inclusive of start, exclusive of end. Can input 0000 of start date and 0000 end date and get all times up but no including 0000 of end date.
//Pass in start and end time in UTC TZ
func deleteFlightsFromTableBetweenTimesForOrigin(table string, start time.Time, end time.Time, origin string) (err error) {
	if err = checkDatabaseHandleValid(db); err != nil {
		return
	}

	var result sql.Result
	if result, err = db.Exec(fmt.Sprintf(`
		DELETE FROM %v 
 		WHERE Origin=$1 AND RollCall >= $2 AND RollCall < $3;
 		`, table), origin, start, end); err != nil {
		return
	}
	var affected int64
	if affected, err = result.RowsAffected(); err != nil {
		return
	}

	fmt.Printf("Delete flights between times for origin %v %v %v %v\n%v rows affected\n", table, start, end, origin, affected)

	return
}
