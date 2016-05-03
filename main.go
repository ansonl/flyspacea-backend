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

type Photo struct {
	Source string `json:"source"`
}

type PhotosPhoto struct {
	CreatedTime string `json:"created_time"`
    Name string `json:"name"`
	Id  string `json:"id"`
}

type Photos struct {
	Data   []PhotosPhoto `json:"data"`
}

type TextBlock struct {
	Text string `json:"text"`
}

type OCRHavenResp struct {
	TextBlockArray []TextBlock `json:"text_block"`
}

type TesseractOCRResp struct {
	TextArray []string `json:"ocr"`
}

type Departure struct {
    RollCall time.Time `json:"rollCall"`
    Origin string `json:"origin"`
    Destination string `json:"destination"`
    SeatCount int `json:"seatCount"`
    SeatType string `json:"seatType"`
    Canceled bool `json:"canceled"`
}

type Terminal struct {
    Title string `json:"title"`
    Id string `json:"id"`
}

var db *(sql.DB)

func sendTesseractOCRRequest(photoUrl string) TesseractOCRResp {
	apiUrl := "https://tesseract-harbor.herokuapp.com"
    resource := "/process"
    data := url.Values{}
    data.Add("url", photoUrl)

    u, _ := url.ParseRequestURI(apiUrl)
    u.Path = resource
    urlStr := fmt.Sprintf("%v", u)

    client := &http.Client{}
    r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode())) // <-- URL-encoded payload
    r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

    fmt.Println("Tesseract OCR request sent.")

    tesseractResp, _ := client.Do(r)
    //fmt.Println(tesseractResp.Status)
    tesseractOutputRaw, ocrErr := ioutil.ReadAll(tesseractResp.Body)
    if (ocrErr != nil) {
    	fmt.Println("Error reading OCR request.")
    }

    //fmt.Println(string(tesseractOutputRaw))

    tesseractOCRResp := TesseractOCRResp{}
    ocrUnmarshalErr := json.Unmarshal(tesseractOutputRaw, &tesseractOCRResp)
    if (ocrUnmarshalErr != nil) {
    	fmt.Println("Error unmarshaling OCR request.")
    }
    fmt.Println("OCR request received.")
    
    /*
    //Print Tesseract OCR response in struct
    testOCRTesseractRespString, _ := json.Marshal(tesseractOCRResp)
	fmt.Println(string(testOCRTesseractRespString))
    */
	
	return tesseractOCRResp
}

func sendHavenOCRRequest(photoUrl string) OCRHavenResp {
	apiUrl := "https://api.havenondemand.com"
    resource := "/1/api/sync/ocrdocument/v1"
    data := url.Values{}
    data.Set("apikey", "6f5569d3-8fc0-4c74-80b4-efe9fd4c90a0")
    data.Add("url", photoUrl)
    data.Add("mode", "document_scan")

    u, _ := url.ParseRequestURI(apiUrl)
    u.Path = resource
    urlStr := fmt.Sprintf("%v", u)

    client := &http.Client{}
    r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode())) // <-- URL-encoded payload
    r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

    fmt.Println("Haven OCR request sent.")

    havenResp, _ := client.Do(r)
    //fmt.Println(havenResp.Status)
    ocrOutputRaw, ocrErr := ioutil.ReadAll(havenResp.Body)
    if (ocrErr != nil) {
    	fmt.Println("Error reading OCR request.")
    }

    /*
    //Test ocr response unmarshalling
    ocrOutputRaw := []byte(`
    	{
		    "text_block": [
		        {
		            "text": "DEPARTURES FROM: ANDERSEN AFB, GUAM\n(UAM)\nMONDAY 18 APRIL 2016\n;C•:j&quot;l -i-, m;lt 0 l31,.;.t;;;l;;i&gt;:1\n0800 ELMENDORF AFB, AK - MCCONELL AFB\nKS\n&apos; 10F\n15 l;;&apos; : 5* £ ; 11=: U rt}? ; ; , ,.1 ? BE £ g;k;F;\n1750 HICKAM AFB, 1-11 10T\nSeat Releases: T=Tentative; F=Finn\nAll flights subject to change without notice",
		            "left": 0,
		            "top": 0,
		            "width": 720,
		            "height": 538
		        }
		    ]
		}
    `)
    */

    ocrHavenResp := OCRHavenResp{}
    ocrUnmarshalErr := json.Unmarshal(ocrOutputRaw, &ocrHavenResp)
    if (ocrUnmarshalErr != nil) {
    	fmt.Println("Error unmarshaling OCR request.")
    }
    fmt.Println("OCR request received.")
    
    /*
    //Print Haven OCR response in struct
    testOCRHavenRespString, _ := json.Marshal(ocrHavenResp)
	fmt.Println(string(testOCRHavenRespString))
	*/

	return ocrHavenResp
}

func parseOCRResponseByNewline(splitResponse []string) []Departure {
    var departures []Departure
    departures = make([]Departure, 0)

    for i := 0; i < len(splitResponse); i++ {
        //fmt.Printf("%v %v\n", i, splitResponse[i])
        //find photo date
        //handle both 01 April 2016 AND April 28th, 2016 AND April 01, 2016 formats
        rePhotoDate := regexp.MustCompile(`(?:(?i:[0-9]{2}[ ]*(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)[ ]*[0-9Il]{2,4})|(?i:(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)[ ]*[0-9]{2}(?:st|nd|rd|th|)?[ ]*,?[ ]+[0-9Il]{2,4}))`)
        photoDate := rePhotoDate.FindString(splitResponse[i])

        if (len(photoDate) > 0) {
            fmt.Printf("Photo date %v\n", photoDate)
            for i = i+1;i < len(splitResponse); i++ {
                //fmt.Printf("%v %v\n", i, splitResponse[i])

                //Format1 = 1405 OSAN AB, KOR 10T
                reDepartureFormat1 := regexp.MustCompile(`(?:([0-9]{4})[ ]+([A-z0-9 .\/,&;-]+?)[ ]+(?:(?:([0-9]{1,3})([TF]))|(TBD)))`)
                foundDeparture := reDepartureFormat1.FindStringSubmatch(splitResponse[i])
                
                if (len(foundDeparture) == 0) {
                    //Format2 = JB ANDREWS, MD 0920 10T
                    reDepartureFormat2 := regexp.MustCompile(`(?:([A-z0-9 .\/,&;-]+?)(?:[ ]+)([0-9]{4})(?:)[ ]+)(?:(?:([0-9]{1,3})([TF]))|(TBD))`)
                    foundDeparture = reDepartureFormat2.FindStringSubmatch(splitResponse[i])
                    if (len(foundDeparture) > 3) {
                        tmp := foundDeparture[1]
                        foundDeparture[1] = foundDeparture[2]
                        foundDeparture[2] = tmp

                        //The ([0-9]{1,3})([TF]) counts as captrue groups #3 and #4 so (TBD) is #5, if TDB, move TDB match to group #3 so we can handle it further on
                        if (foundDeparture[5] == "TBD") {
                            foundDeparture[3] = "TBD"
                        }
                    }

                }

                //fmt.Printf("REGEX FIND %q", foundDeparture)
                //Array should be in order ["1405 OSAN AB, KOR 10T" "1405" "OSAN AB, KOR" "10" "T" ""]



                if (len(foundDeparture) > 3) {
                    
                    var tmpDeparture Departure
                    tmpDeparture = Departure{}

                    layout := "02 January 2006 1504"

                    //Replace I and l strings in year of captured date with 1 string to successfully parse date
                    photoDate = photoDate[:len(photoDate)-4] + strings.Replace(photoDate[len(photoDate)-4:], "I", "1", -1)
                    photoDate = photoDate[:len(photoDate)-4] + strings.Replace(photoDate[len(photoDate)-4:], "l", "1", -1)

                    addedTimeString := photoDate + " " + foundDeparture[1]

                    tmpRollCall, rollCallErr := time.Parse(layout, addedTimeString)
                    if rollCallErr != nil {
                        fmt.Println(rollCallErr)
                        return nil
                    }

                    tmpDeparture.RollCall = tmpRollCall

                    tmpDeparture.Origin = "???"
                    tmpDeparture.Destination = foundDeparture[2]
                    if (foundDeparture[3] == "TBD") {
                        tmpDeparture.SeatType = "TBD"
                    } else if (len(foundDeparture) > 4) {
                        seatCount, atioErr := strconv.Atoi(foundDeparture[3])
                        if atioErr != nil {
                            fmt.Println(atioErr)
                        }

                        tmpDeparture.SeatCount = seatCount

                        tmpDeparture.SeatType = foundDeparture[4]
                    } else {
                        fmt.Printf("Problem with length of match %q\n", foundDeparture)
                    }
                    
                    departures = append(departures, tmpDeparture)
                }
            }
            break
        }
    }

    return departures
}

func selectRowsFromOrigin4Weeks(targetTable string, targetOrigin string) *(sql.Rows) {
    //we construct the SELECT query in Go because SQL does not support ordinal marker for table names
    query := fmt.Sprintf(`SELECT Origin, Destination, RollCall, SeatCount, SeatType, Cancelled` +" FROM %v "+
 `WHERE Origin=$1 AND RollCall <= current_timestamp + INTERVAL '14 day' AND RollCall > current_timestamp - INTERVAL '14 day';`, targetTable)
    rows, err := db.Query(query, targetOrigin)
    if err != nil {
        log.Println(err)
    } else {
        return rows
    }
    return nil
}

func scanFlightRowsToMap(targetRows *(sql.Rows)) map[string]Departure {
    var existingSavedFlights map[string]Departure
    existingSavedFlights = make(map[string]Departure)
    
    for targetRows.Next() {
        var tmpDeparture Departure
        tmpDeparture = Departure{}

        if err := targetRows.Scan(&tmpDeparture.Origin, &tmpDeparture.Destination, &tmpDeparture.RollCall, &tmpDeparture.SeatCount, &tmpDeparture.SeatType, &tmpDeparture.Canceled); err != nil {
            log.Println(err)
        }

        key := fmt.Sprintf("%v*%v*%v", tmpDeparture.Origin, tmpDeparture.Destination, tmpDeparture.RollCall.Format(time.RFC3339))
        //fmt.Println(key)

        existingSavedFlights[key] = tmpDeparture
    }
    targetRows.Close()
    return existingSavedFlights
}

func updateDatabaseWithUpdatedFlights(updatedFlightArray []Departure) {
    fmt.Println("COMPARE DB")

    targetRows := selectRowsFromOrigin4Weeks("flights", "???")
    existingSavedFlights := scanFlightRowsToMap(targetRows)

    //find new flights that are not previously saved and update old flights
    for _, v := range updatedFlightArray {
        key := fmt.Sprintf("%v*%v*%v", v.Origin, v.Destination, v.RollCall.Format(time.RFC3339))
        //fmt.Println(key)

        //Check if updated flight already exists in database
        if existingDeparture, ok := existingSavedFlights[key]; ok {
            //If updated flight seat count/type differs from database version, UPDATE existing row in database
            if (v.SeatCount != existingDeparture.SeatCount || v.SeatType != existingDeparture.SeatType) {
                var result sql.Result
                var err error
                if result, err = db.Exec(`UPDATE flights 
                    SET SeatCount = $1, SeatType = $2 
                    WHERE Origin = $3 AND Aestination = $4 AND RollCall = $5;`, v.SeatCount, v.SeatType, v.Origin, v.Destination, v.RollCall); err != nil {
                    log.Println(err)
                } else {
                    rowsAffected, _ := result.RowsAffected()
                    fmt.Printf("UPDATE %v rows affected\n", rowsAffected)
                }
            }
        } else { //If updated flight does not exist in database, INSERT new row for flight into database
            //insert
            var result sql.Result
            var err error
            if result, err = db.Exec(`INSERT INTO flights (Origin, Destination, RollCall, SeatCount, SeatType, Cancelled) 
                VALUES ($1, $2, $3, $4, $5, $6);`, v.Origin, v.Destination, v.RollCall, v.SeatCount, v.SeatType, v.Canceled); err != nil {
                log.Println(err)
            } else {
                rowsAffected, _ := result.RowsAffected()
                fmt.Printf("INSERT %v rows affected\n", rowsAffected)
            }
        }
    }
}

func setupDatabaseHandle() {
    //Create global db handle
    var err error //define err because mixing it with the global db var and := operator creates local scoped db
    db, err = sql.Open("postgres", os.Getenv("SPACEA_DATABASE_URL"))
    if err != nil {
        log.Println(err)
    }
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

func updateTerminal(targetTerminal Terminal) {
    terminalId := targetTerminal.Id
    if (len(terminalId) == 0) {
        fmt.Printf("Terminal %v missing Id.\n", targetTerminal.Title)
        return
    }
        

    apiUrl := "https://graph.facebook.com"
    resource := fmt.Sprintf("v2.5/%v/photos", terminalId)
    data := url.Values{}
    data.Add("type", "uploaded")
    data.Add("access_token", "522755171230853%7ChxS9OzJ4I0CqmmrESRpNHfx77vs")

    u, _ := url.ParseRequestURI(apiUrl)
    u.Path = resource
    urlStr := fmt.Sprintf("%v?%v", u, data.Encode())

    client := &http.Client{}
    r, _ := http.NewRequest("GET", urlStr, nil)
    r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

    pagePhotosEdgeResp, _ := client.Do(r)
    body, ocrErr := ioutil.ReadAll(pagePhotosEdgeResp.Body)
    if (ocrErr != nil) {
        fmt.Println("Error reading page photos edge resp.")
    }

    pagePhotos := Photos{}
    pagePhotosErr := json.Unmarshal(body, &pagePhotos)
    if pagePhotosErr != nil {
        fmt.Println("Error unmarshaling page photos edge.")
    }
    fmt.Println("Page photos edge retrieved.")

    /*
    res2B, _ := json.Marshal(pagePhotos)
    fmt.Println(string(res2B))
    */

    //fmt.Printf("%v\n", string(body))

    //Look at the photo nodes returned by the photos edge
    var limit int
    limit = 4
    if len(pagePhotos.Data) < limit {
        limit = len(pagePhotos.Data)
    }

    for i := 0; i < limit; i++ {
        v := pagePhotos.Data[i]

        //var timeOjb time.Time

        /* http://stackoverflow.com/questions/24401901/time-parse-why-does-golang-parses-the-time-incorrectly */
        /*
        layout := "2006-01-02T15:04:05-0700"
        timeOjb, err := time.Parse(layout, v.CreatedTime)
        if err != nil {
            fmt.Println(err)
        }
        fmt.Println(timeOjb)
        */

        fmt.Println("Getting photo node.")

        apiUrl := "https://graph.facebook.com"
        resource := fmt.Sprintf("v2.5/%v", v.Id)
        data := url.Values{}
        data.Add("fields", "source")
        data.Add("access_token", "522755171230853%7ChxS9OzJ4I0CqmmrESRpNHfx77vs")

        u, _ := url.ParseRequestURI(apiUrl)
        u.Path = resource
        urlStr := fmt.Sprintf("%v?%v", u, data.Encode())

        client := &http.Client{}
        r, _ := http.NewRequest("GET", urlStr, nil)
        r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
        r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

        resp, err := client.Do(r)
        if err != nil {
            // handle error
        }

        //Read response body into []byte
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)

        photo := Photo{}
        photoNodeErr := json.Unmarshal(body, &photo)
        if (photoNodeErr != nil) {
            fmt.Println("Error unmarshaling photoNode")
        }
        fmt.Println("Photo node retrieved.")

        
        testPhotoNodeString, _ := json.Marshal(photo)
        fmt.Println(string(testPhotoNodeString))
        
        
        var updatedFlightArray []Departure

        //Haven On Demand OCR request
        ocrHavenResp := sendHavenOCRRequest(photo.Source)
        if (len(ocrHavenResp.TextBlockArray) == 0) {
            fmt.Println("No text recognized by Haven On Demand OCR")
        } else {
            //Haven OCR request contains one string with newlines so we split response into array of strings by newline
            //Split by \n characters
            splitByString := "\n"
            var splitResponse []string
            if (len(splitByString) > 0) {
                splitResponse = strings.Split(ocrHavenResp.TextBlockArray[0].Text, splitByString)
                //fmt.Printf("%v\n", splitResponse)
            }
            if updatedFlightArray = parseOCRResponseByNewline(splitResponse); updatedFlightArray != nil {
                for i, _ := range updatedFlightArray {
                    updatedFlightArray[i].Origin = targetTerminal.Title
                }
                updateDatabaseWithUpdatedFlights(updatedFlightArray)
            } else {
                fmt.Printf("Error parsing OCR Response for %v\n", targetTerminal.Title)
                return
            }
            fmt.Printf("HAVEN score %v\n", len(updatedFlightArray))
            
        }

        //Tesseract OCR request
        ocrTesseractResp := sendTesseractOCRRequest(photo.Source)
        if (len(ocrTesseractResp.TextArray) == 0) {
            fmt.Println("No text recognized by Tesseract OCR.")
        }
        if updatedFlightArray = parseOCRResponseByNewline(ocrTesseractResp.TextArray); updatedFlightArray != nil {
            for i, _ := range updatedFlightArray {
                updatedFlightArray[i].Origin = targetTerminal.Title
            }
            updateDatabaseWithUpdatedFlights(updatedFlightArray)
        } else {
            fmt.Printf("Error parsing OCR Response for %v\n", targetTerminal.Title)
            return
        }
        fmt.Printf("Tesseract score %v\n", len(updatedFlightArray))
        

        /*
        //Print what regex finds
        fmt.Printf("WHAT WE FOUND\n")
        for _, v := range updatedFlightArray {
            tmpDeparture, _ := json.Marshal(v)
            fmt.Println(string(tmpDeparture))
        }
        */
    }
}

func updateAllTerminals(terminalArray []Terminal) {

}

func main() {
    setupDatabaseHandle()

    //terminalMap := readTerminalFileToMap("terminals.json")

    /*
    for _, v := range terminalMap {
        updateTerminal(v)
        break
    }
    */
    v := Terminal{}
    v.Title="Andersen Pax Term"
    v.Id="321748637899280"
    updateTerminal(v)

    
    return

    //fmt.Printf("%v\n",readTerminalFileToArray("terminals.json"))

}
