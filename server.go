package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

//Marshal SAResponse into JSON, return error string if Marshal error.
func (resp SAResponse) createJSONOutput() string {
	output, err := json.Marshal(resp)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}
	return string(output)
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

func runServer(wg *sync.WaitGroup, config *tls.Config) {

	//Refresh specific terminal
	//http.HandleFunc("/refreshTerminal", refreshTerminalHandler)

	//Get flights for parameter filters
	http.HandleFunc("/flights", flightsHandler)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}

	log.Println("Server ended on port " + os.Getenv("PORT"))

	wg.Done()
}
