package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"strings"
)

var serverStartTime time.Time

//Marshal SAResponse into JSON, return error string if Marshal error.
func (resp SAResponse) createJSONOutput() string {
	output, err := json.Marshal(resp)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}
	return string(output)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	//bypass same origin policy
	w.Header().Set("Access-Control-Allow-Origin", "*")

	http.Redirect(w, r, "https://github.com/ansonl/shipmate", http.StatusFound)

	log.Println("About requested")
}

func uptimeHandler(w http.ResponseWriter, r *http.Request) {
	//bypass same origin policy
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//Get approx table rows
	//https://wiki.postgresql.org/wiki/Count_estimate
	getTableRows := func(table string) (rows int, err error) {
		if err = checkDatabaseHandleValid(db); err != nil {
			return
		}

		var flightRows *sql.Rows
		if flightRows, err = db.Query(fmt.Sprintf(`
			SELECT reltuples FROM pg_class WHERE relname = '%v';
			`, table)); err != nil {
			return
		}

		for flightRows.Next() {
			if err = flightRows.Scan(&rows); err != nil {
				return
			}
		}
		return
	}

	diff := time.Since(serverStartTime)
	var err error
	var flightRows int
	if flightRows, err = getTableRows(FLIGHTS_72HR_TABLE); err != nil {
		fmt.Fprintf(w, "Error: %v\n", err.Error())
	}

	fmt.Fprintf(w, `Server uptime:	%v
Process mode:	%v
Flights stored (approx):	%v

%v

Current:
%v`, diff.String(), *processMode, flightRows, statisticsString(), liveStatisticsString())
}

func locationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var err error

	//Parse HTTP Form
	if err = r.ParseForm(); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("Parse form error: %v", err.Error())}.createJSONOutput())
		return
	}

	//Initialize variables needed to find flights
	var startTime time.Time
	var duration time.Duration

	//Parse REST_START_TIME_KEY
	var startTimeText string
	if startTimeText = r.Form.Get(REST_START_TIME_KEY); len(startTimeText) > 0 {
		if startTime, err = time.Parse(time.RFC3339, startTimeText); err != nil {
			fmt.Fprintf(w, SAResponse{
				Status: 1,
				Error:  fmt.Sprintf("%v parameter error: %v", REST_START_TIME_KEY, err.Error())}.createJSONOutput())
			return
		}
	}

	//Parse REST_DURATION_DAYS_KEY
	var durationDaysText string
	if durationDaysText = r.Form.Get(REST_DURATION_DAYS_KEY); len(durationDaysText) > 0 {
		var durationDays int
		if durationDays, err = strconv.Atoi(durationDaysText); err != nil {
			fmt.Fprintf(w, SAResponse{
				Status: 1,
				Error:  fmt.Sprintf("%v parameter error: %v", REST_DURATION_DAYS_KEY, err.Error())}.createJSONOutput())
			return
		}
		duration = time.Hour * 24 * time.Duration(durationDays) //Convert duration to days
	}
	
	

	var locationsArr []string
	if locationsArr, err = selectActiveLocationsFromTable(
		FLIGHTS_72HR_TABLE,
		startTime,
		duration); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("Get locations error: %v", err.Error())}.createJSONOutput())
		return
	}

	fmt.Fprintf(w, SAResponse{
		Status:    0,
		Locations: locationsArr}.createJSONOutput())
}

/*
func allLocationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var err error

	var locationsArr []Terminal
	if locationsArr, err = selectAllLocationsFromTable(
		FLIGHTS_72HR_TABLE); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("Get locations error: %v", err.Error())}.createJSONOutput())
		return
	}
	//TODO: decide what to do with all locations data and whether to switch SAResponse.locations from []string to []Terminal

	fmt.Fprintf(w, SAResponse{
		Status:    0,
		Locations: locationsArr}.createJSONOutput())
		
}
*/

func oldestRollCallHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var err error
	var oldest time.Time
	if oldest, err = selectOldestRollCallDateFromTable(FLIGHTS_72HR_TABLE); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("Get oldest rollcall error: %v", err.Error())}.createJSONOutput())
		return
	}

	fmt.Fprintf(w, SAResponse{
		Status:    0,
		Data: oldest.Format("2006-01-02T15:04:05Z")}.createJSONOutput())
}

func flightsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var err error

	//Parse HTTP Form
	if err = r.ParseForm(); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("Parse form error: %v", err.Error())}.createJSONOutput())
		return
	}

	//Initialize variables needed to find flights
	var origin string
	var destination string
	var startTime time.Time
	var duration time.Duration

	origin = r.Form.Get(REST_ORIGIN_KEY)
	destination = r.Form.Get(REST_DESTINATION_KEY)

	//Parse REST_START_TIME_KEY
	var startTimeText string
	if startTimeText = r.Form.Get(REST_START_TIME_KEY); len(startTimeText) == 0 {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("Missing %v parameter.", REST_START_TIME_KEY)}.createJSONOutput())
		return
	}
	if startTime, err = time.Parse(time.RFC3339, startTimeText); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("%v parameter error: %v", REST_START_TIME_KEY, err.Error())}.createJSONOutput())
		return
	}

	//Parse REST_DURATION_DAYS_KEY
	var durationDaysText string
	if durationDaysText = r.Form.Get(REST_DURATION_DAYS_KEY); len(durationDaysText) == 0 {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("Missing %v parameter.", REST_DURATION_DAYS_KEY)}.createJSONOutput())
		return
	}
	var durationDays int
	if durationDays, err = strconv.Atoi(durationDaysText); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("%v parameter error: %v", REST_DURATION_DAYS_KEY, err.Error())}.createJSONOutput())
		return
	}
	duration = time.Hour * 24 * time.Duration(durationDays) //Convert duration to days

	var foundFlights []Flight
	if foundFlights, err = selectFlightsFromTableWithOriginDestTimeDuration(
		FLIGHTS_72HR_TABLE,
		origin,
		destination,
		startTime,
		duration); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 2,
			Error:  fmt.Sprintf("Query error: %v", err.Error())}.createJSONOutput())
		return
	}

	fmt.Fprintf(w, SAResponse{
		Status:  0,
		Flights: foundFlights}.createJSONOutput())
}

func submitPhotoReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var err error

	//Parse HTTP Form
	if err = r.ParseForm(); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("Parse form error: %v", err.Error())}.createJSONOutput())
		return
	}

	//Initialize variables needed to find flights
	var location string
	var photoSource string
	var comment string

	location = r.Form.Get(REST_LOCATION_KEY)
	photoSource = r.Form.Get(REST_PHOTOSOURCE_KEY)
	comment = r.Form.Get(REST_COMMENT_KEY)

	//Get X-forwarded-for IP from header or request remoteaddr field.
	reportIP := r.RemoteAddr
	forwardedIPs := strings.Split(r.Header.Get("X-FORWARDED-FOR"), ", ")
	if len(forwardedIPs) > 0 {
		reportIP = forwardedIPs[len(forwardedIPs)-1]
	}

	submittedPhotoReport := PhotoReport{
		Location: location,
		PhotoSource: photoSource,
		Comment: comment,
		SubmitDate: time.Now(),
		IPAddress: reportIP}

	if err = insertPhotoReportIntoTable(PHOTOS_REPORTS_TABLE, submittedPhotoReport); err != nil {
		fmt.Fprintf(w, SAResponse{
			Status: 1,
			Error:  fmt.Sprintf("Insert photo report error: %v", err.Error())}.createJSONOutput())
		return
	}

	fmt.Fprintf(w, SAResponse{
		Status:  0}.createJSONOutput())
}

func runServer(wg *sync.WaitGroup, config *tls.Config) {

	serverStartTime = time.Now()

	//Refresh specific terminal
	//http.HandleFunc("/refreshTerminal", refreshTerminalHandler)
	
	http.HandleFunc("/uptime", uptimeHandler)

	//Get active locations within a time range
	http.HandleFunc("/locations", locationsHandler)

	//Get active locations within a time range
	//http.HandleFunc("/allLocations", allLocationsHandler)

	//Get oldest rollcall within the last year
	http.HandleFunc("/oldestRC", oldestRollCallHandler)

	//Get flights for parameter filters
	http.HandleFunc("/flights", flightsHandler)

	//Log photo report from user
	http.HandleFunc("/submitPhotoReport", submitPhotoReportHandler)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}

	log.Println("Server ended on port " + os.Getenv("PORT"))

	wg.Done()
}
