package main

import (
	"log"
	//"time"
)

func main() {
	//fmt.Printf("\n\u001b[1mboldtext\u001b[0m\r\u001b[2Fprevline\n\n\n")

	createFuzzyModels(&fuzzyModelForKeyword)

	terminalArray, err := readTerminalFileToArray("terminals.json")
	if err != nil {
		log.Fatal(err)
	}
	terminalMap := readTerminalArrayToMap(terminalArray)

	log.Printf("\u001b[1m\u001b[31m%v\u001b[0m\n", "Starting Update")
	updateAllTerminals(terminalMap)

	/*
	   v := Terminal{}
	   v.Title="Travis Pax Term"
	   v.Id="travispassengerterminal"
	   updateTerminal(v)
	*/

	//fmt.Printf("%v\n",readTerminalFileToArray("terminals.json"))

}
