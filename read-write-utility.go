package main

import (
	"encoding/json"
	"io/ioutil"
	//"log"
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

func readKeywordsToArrayFromFiles(filenames ...string) (keywordsArray []LocationKeywords, err error) {
	for _, filename := range filenames {
		var locationsRaw []byte
		locationsRaw, err = ioutil.ReadFile(filename)
		if err != nil {
			return
		}

		var tmp []LocationKeywords
		if err = json.Unmarshal(locationsRaw, &tmp); err != nil {
			return
		}

		//log.Printf("%v locations in file %v", len(tmp), filenames)

		keywordsArray = append(keywordsArray, tmp...)
	}

	return
}
