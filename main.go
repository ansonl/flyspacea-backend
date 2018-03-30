package main

import (
	"log"
	//"time"
)

func main() {
	//fmt.Printf("\n\u001b[1mboldtext\u001b[0m\r\u001b[2Fprevline\n\n\n")

	var err error
	if err = setupDatabase(); err != nil {
		log.Println(err)
	}

	selectFlightsFromTableWithOriginDestTimeDuration

	if err = createFuzzyModels(); err != nil {
		log.Fatal(err)
	}

	terminalArray, err := readTerminalArrayFromFiles(TERMINAL_SINGLE_FILE)
	if err != nil {
		log.Fatal(err)
	}
	terminalMap := readTerminalArrayToMap(terminalArray)

	log.Printf("\u001b[1m\u001b[35m%v\u001b[0m\n", "Starting Update")
	updateAllTerminals(terminalMap)

	/*
	   v := Terminal{}
	   v.Title="Travis Pax Term"
	   v.Id="travispassengerterminal"
	   updateTerminal(v)
	*/

	//fmt.Printf("%v\n",readTerminalFileToArray("terminals.json"))

}
