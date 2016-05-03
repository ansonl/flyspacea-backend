package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"io/ioutil"
	"regexp"
	"strings"
	"net/url"
	"bytes"
	"strconv"
    "database/sql"
    _ "github.com/lib/pq"
    "os"
    "log"
)

type Terminal struct {
    Title string `json:"title"`
    Url string `json:"url"`
    Id string `json:"id"`
}

func readTerminalFileToMap(terminalFilename string) map[string]Terminal {
    terminalsRaw, readErr := ioutil.ReadFile(terminalFilename)
    if (readErr != nil) {
        log.Println(readErr)
    }
    var terminalArray []Terminal
    terminalErr := json.Unmarshal(terminalsRaw, &terminalArray)
    if terminalErr != nil {
        log.Println(terminalErr)
    }

    //set key to title
    var terminalMap map[string]Terminal
    terminalMap = make(map[string]Terminal)
    for _, v := range terminalArray {
        terminalMap[v.Title] = v
    }

    return terminalMap
}

func updateMapWIthIdForTerminal(targetTerminal Terminal) {
	
}

func main() {
    terminalMap := readTerminalFileToMap("terminals.json")

    for _, v := range terminalMap {
        updateMapWIthIdForTerminal(v)
        break
    }

}
