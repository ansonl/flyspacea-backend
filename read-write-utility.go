package main

import (
	"encoding/json"
	"io/ioutil"
)

func readTerminalFileToArray(filename string) (terminalArray []Terminal, err error) {
	terminalsRaw, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	if err = json.Unmarshal(terminalsRaw, &terminalArray); err != nil {
		return
	}

	return
}

func readTerminalArrayToMap(terminalArray []Terminal) (terminalMap map[string]Terminal) {
	//set key to v.Title and set v.Index
	terminalMap = make(map[string]Terminal)
	for _, v := range terminalArray {
		terminalMap[v.Title] = v
	}

	return
}

func readTerminalKeywordsFileToArray(filename string) (terminalKeywordsArray []TerminalKeywords, err error) {
	terminalsRaw, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	if err = json.Unmarshal(terminalsRaw, &terminalKeywordsArray); err != nil {
		return
	}

	return
}