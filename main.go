package main

import (
	"flag"
	"log"
	"sync"
	"time"
)

/*
 * curl http://localhost:8080/debug/pprof/heap > base.heap
 * go tool pprof -base base.heap ../spacea
 * pdf
 */
//import _ "net/http/pprof"

var processMode = flag.String("procMode", "all", "Process Mode for server. all/web/worker")

func main() {
	//fmt.Printf("\n\u001b[1mboldtext\u001b[0m\r\u001b[2Fprevline\n\n\n")

	/*
		deleteFlightsFromTableForDayForOriginTerminal("", time.Now(), Terminal{
			Timezone: "Beijing"})
		return
	*/

	var wg sync.WaitGroup

	//Start web server
	startWebMode := func() {
		//Start HTTP server
		wg.Add(1)
		go runServer(&wg, nil)
	}

	//Start flight image processing and database setup
	startWorkerMode := func() {
		var err error

		if err = createImageDirectories(IMAGE_TMP_DIRECTORY, IMAGE_TRAINING_DIRECTORY, IMAGE_TRAINING_PROCESSED_DIRECTORY_BLACK, IMAGE_TRAINING_PROCESSED_DIRECTORY_WHITE); err != nil {
			log.Println(err)
		}

		//Setup storage database
		if err = createDatabaseTables(); err != nil {
			log.Println(err)
		}

		//Load terminals
		var terminalsToUpdateFile string
		terminalsToUpdateFile = TERMINAL_FILE
		if DEBUG_SINGLE_FILE {
			terminalsToUpdateFile = TERMINAL_SINGLE_FILE
		}

		var terminalArray []Terminal
		if terminalArray, err = readTerminalArrayFromFiles(terminalsToUpdateFile); err != nil {
			log.Fatal(err)
		}

		//Read in location keyword file
		getAllTerminalsInfo(terminalArray)
		//log.Println(terminalArray)
		terminalMap := readTerminalArrayToMap(terminalArray)
		log.Printf("Loaded %v Terminals.\n", len(terminalArray))


		//Population locations table with locations from file
		if err = populateLocationsTable(terminalArray); err != nil {
			log.Println(err)
			return
		}

		log.Printf("\u001b[1m\u001b[35m%v\u001b[0m\n", "Starting Update")

		//Update terminal flights every hour
		updateAllTerminalsFlights(terminalMap)
		for _ = range time.Tick(time.Minute * 30) {
			updateAllTerminalsFlights(terminalMap)
		}
		//go updateAllTerminalsFlights(terminalMap)
	}

	//Parse cmd parameters and launch appropriate mode
	flag.Parse()
	if *processMode == "web" {
		connectDatabase()
		startWebMode()
	} else if *processMode == "worker" {
		connectDatabase()
		startWorkerMode()
	} else if *processMode == "all" {
		connectDatabase()
		startWebMode()
		startWorkerMode()
	} else {
		log.Println("procMode " + *processMode + " invalid.")
		flag.PrintDefaults()
		return
	}

	//Wait for server to end
	wg.Wait()

	/*
	   v := Terminal{}
	   v.Title="Travis Pax Term"
	   v.Id="travispassengerterminal"
	   updateTerminal(v)
	*/

	//fmt.Printf("%v\n",readTerminalFileToArray("terminals.json"))

	/*
		//Test table selection
		var startDate time.Time
		if startDate, err = time.Parse("2006-01-02", "2018-03-23"); err != nil {
			log.Panic(err)
		}
		var flightsSelected []Flight
		if flightsSelected, err = selectFlightsFromTableWithOriginDestTimeDuration(FLIGHTS_72HR_TABLE, "", "", startDate, time.Hour*24); err != nil {
			log.Panic(err)
		} else {
			for _, f := range flightsSelected {
				log.Println(f)
			}
		}
		return
	*/

	/*
		//Fuzzy model creation release test
		//Create fuzzy models for lookup
		if err := createFuzzyModels(); err != nil {
			log.Fatal(err)
		}
		log.Println("created fuzzy model")

		time.Sleep(time.Second*10)

		//Tear down fuzzy models to release memory
		destroyFuzzyModels()
		log.Println("destroyed fuzzy model")
	*/

}
