package main

import (
	"encoding/json"
	"io/ioutil"
	//"log"
)

func readTerminalArrayToMap(terminalArray []Terminal) (terminalMap map[string]Terminal) {
	//set key to v.Title and set v.Index
	terminalMap = make(map[string]Terminal)
	for _, v := range terminalArray {
		terminalMap[v.Title] = v
	}

	return
}

func readTerminalArrayFromFiles(filenames ...string) (keywordsArray []Terminal, err error) {
	for _, filename := range filenames {
		var locationsRaw []byte
		locationsRaw, err = ioutil.ReadFile(filename)
		if err != nil {
			return
		}

		var tmp []Terminal
		if err = json.Unmarshal(locationsRaw, &tmp); err != nil {
			return
		}

		//log.Printf("%v locations in file %v", len(tmp), filenames)

		keywordsArray = append(keywordsArray, tmp...)
	}

	return
}
