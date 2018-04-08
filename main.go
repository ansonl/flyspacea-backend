package main

import (
	"log"
	"sync"
	//"time"
)

func main() {
	//fmt.Printf("\n\u001b[1mboldtext\u001b[0m\r\u001b[2Fprevline\n\n\n")

	/*
	deleteFlightsFromTableForDayForOriginTerminal("", time.Now(), Terminal{
		Timezone: "Beijing"})
	return
	*/

	var err error

	//Load terminals
	var terminalArray []Terminal
	if terminalArray, err = readTerminalArrayFromFiles(TERMINAL_FILE); err != nil {
		log.Fatal(err)
	}
	//Read in location keyword file
	getAllTerminalsInfo(terminalArray)
	//log.Println(terminalArray)
	terminalMap := readTerminalArrayToMap(terminalArray)
	log.Printf("Loaded %v Terminals.\n", len(terminalArray))

	//Setup storage database
	if err = createDatabase(); err != nil {
		log.Println(err)
	}
	log.Println("Created storage database.")

	if err = populateLocationsTable(nil); err != nil {
		log.Println(err)
		return
	}
	log.Println("Populated storage reference database.")

	

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

	//Start HTTP server
	var wg sync.WaitGroup
	wg.Add(1)
	go runServer(&wg, nil)

	if err = createFuzzyModels(); err != nil {
		log.Fatal(err)
	}

	log.Printf("\u001b[1m\u001b[35m%v\u001b[0m\n", "Starting Update")
	go updateAllTerminalsFlights(terminalMap)

	//Wait for server to end
	wg.Wait()

	/*
	   v := Terminal{}
	   v.Title="Travis Pax Term"
	   v.Id="travispassengerterminal"
	   updateTerminal(v)
	*/

	//fmt.Printf("%v\n",readTerminalFileToArray("terminals.json"))

}
