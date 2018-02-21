package main

import (
	"log"
	//"fmt"
)

func main() {
	//fmt.Printf("\n\u001b[1mboldtext\u001b[0m\r\u001b[2Fprevline\n\n\n")


	createFuzzyModelsForKeywords([]string {KEYWORD_DESTINATION}, &fuzzyModelForKeyword)

	terminalArray, err := readTerminalFileToArray("terminals.json")
	if err != nil {
		log.Fatal(err)
	}
	terminalMap := readTerminalArrayToMap(terminalArray)

	
	updateAllTerminals(terminalMap)

	/*
	   v := Terminal{}
	   v.Title="Travis Pax Term"
	   v.Id="travispassengerterminal"
	   updateTerminal(v)
	*/

	//fmt.Printf("%v\n",readTerminalFileToArray("terminals.json"))

}


