package main

import (
	"log"
	"sync"
	"time"
)

func main() {
	//fmt.Printf("\n\u001b[1mboldtext\u001b[0m\r\u001b[2Fprevline\n\n\n")

	var err error
	if err = setupDatabase(); err != nil {
		log.Println(err)
	}

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

	if err = createFuzzyModels(); err != nil {
		log.Fatal(err)
	}

	terminalArray, err := readTerminalArrayFromFiles(TERMINAL_SINGLE_FILE)
	if err != nil {
		log.Fatal(err)
	}
	terminalMap := readTerminalArrayToMap(terminalArray)

	log.Printf("\u001b[1m\u001b[35m%v\u001b[0m\n", "Starting Update")
	go updateAllTerminals(terminalMap)

	//start server and wait
	var wg sync.WaitGroup
	wg.Add(1)
	go runServer(&wg, nil)
	wg.Wait()

	/*
	   v := Terminal{}
	   v.Title="Travis Pax Term"
	   v.Id="travispassengerterminal"
	   updateTerminal(v)
	*/

	//fmt.Printf("%v\n",readTerminalFileToArray("terminals.json"))

}
