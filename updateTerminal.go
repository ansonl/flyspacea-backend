package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	//Worker management
	"context"
	"golang.org/x/sync/semaphore"
	"regexp"
	"runtime"

	//GC
	"runtime/debug"
)

//Get all terminals' FB page info
func getAllTerminalsInfo(terminalArray []Terminal) {
	//Spawn goroutine to download and process each image
	ctx := context.TODO()
	var maxWorkers int
	var sem *semaphore.Weighted
	maxWorkers = runtime.GOMAXPROCS(0)
	//maxWorkers = 1
	sem = semaphore.NewWeighted(int64(maxWorkers))

	for i, _ := range terminalArray {
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Printf("Failed to acquire semaphore: %v\n", err)
			break
		}

		go func(t *Terminal) {
			defer sem.Release(1)
			log.Printf("Getting info for terminal %v\n", (*t).Title)
			if err := t.getAndSetTerminalInfo(); err != nil {
				displayErrorForTerminal((*t), err.Error())
			} else {
				//fmt.Println((*t).PageInfoEdge)
				incrementValidTerminals() //No error for first api call to terminal = valid FB id
			}
		}(&terminalArray[i])
	}

	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		log.Printf("Failed to acquire semaphore: %v\n", err)
	}
}

//Get t *Terminal FB page info
func (t *Terminal) getAndSetTerminalInfo() (err error) {
	//Create request url and parameters
	apiUrl := GRAPH_API_URL
	resource := fmt.Sprintf("%v/%v/", GRAPH_API_VERSION, (*t).Id)
	data := url.Values{}
	//Multiple url.Values.Add will Encode to k=v&k=v. Facebook only processes last key.
	data.Add(GRAPH_FIELDS_KEY, fmt.Sprintf("%v,%v,%v", GRAPH_FIELD_PHONE_KEY, GRAPH_FIELD_EMAILS_KEY, GRAPH_FIELD_GENERAL_INFO_KEY))
	data.Add(GRAPH_ACCESS_TOKEN_KEY, GRAPH_ACCESS_TOKEN)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := fmt.Sprintf("%v?%v", u, data.Encode())

	//log.Println(urlStr)

	//Create request
	var req *http.Request
	var client *http.Client
	client = &http.Client{
		Timeout: time.Second * 20}
	if req, err = http.NewRequest("GET", urlStr, nil); err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	//Make request
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}

	//Read response body into []byte
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		//"Error reading page photos edge resp."
		return
	}

	//Unmarshall into struct
	var pageInfoEdge PageInfoEdge
	if err = json.Unmarshal(body, &pageInfoEdge); err != nil {
		//"Error unmarshaling page info edge."
		return
	}

	//Trim phone number and email
	//TODO: Add regex for phone number :(
	phoneSpaceIndex := strings.Index(pageInfoEdge.Phone, " ")
	if phoneSpaceIndex > 10 { //Trim only if space found after 10th digit. In case managers input a space in the beginning of the phone number.
		pageInfoEdge.Phone = pageInfoEdge.Phone[:phoneSpaceIndex]
	}
	if len(pageInfoEdge.Emails) > 0 {
		spaceIndex := strings.Index(pageInfoEdge.Emails[0], " ")
		if spaceIndex > 10 { //Trim only if space found after 10th char for minlen("@us.af.mil"). In case managers input a space in the beginning of the phone number.
			pageInfoEdge.Emails[0] = pageInfoEdge.Emails[0][:spaceIndex]
		}

		if len(pageInfoEdge.Emails[0]) > 255 {
			pageInfoEdge.Emails[0] = pageInfoEdge.Emails[0][:255]
		}
	}
	if len(pageInfoEdge.Phone) > 50 {
		pageInfoEdge.Phone = pageInfoEdge.Phone[:50]
	}
	if len(pageInfoEdge.GeneralInfo) > 2048 {
		pageInfoEdge.GeneralInfo = pageInfoEdge.GeneralInfo[:2048]
	}

	//Check for error
	if pageInfoEdge.Error.Code != 0 {
		err = fmt.Errorf("Code %v\nMessage %v", pageInfoEdge.Error.Code, pageInfoEdge.Error.Message)
		return
	}

	(*t).PageInfoEdge = pageInfoEdge

	return
}

//Update flights for all terminals in terminalMap map[string]Terminal
//Calls fuzzy model creation and release
func updateAllTerminalsFlights(terminalMap map[string]Terminal) {
	resetStatistics()

	var startTime, endTime time.Time
	startTime = time.Now()

	//Set live stats info
	setLiveTotalTerminals(len(terminalMap))

	//Create fuzzy models for lookup
	if err := createFuzzyModels(); err != nil {
		log.Fatal(err)
	}

	for _, v := range terminalMap {
		if err := updateTerminalFlights(v); err != nil {
			displayErrorForTerminal(v, err.Error())
		}

		incrementLiveTerminalsUpdated()
	}

	//Tear down fuzzy models to release memory
	destroyFuzzyModels()

	endTime = time.Now()

	log.Printf("Terminal Flights Update ended.\nStart time: %v\n End time: %v\nElapsed time: %v\n", startTime, endTime, endTime.Sub(startTime))

	displayStatistics()

	debug.FreeOSMemory()
}

//Update targetTerminal flights
func updateTerminalFlights(targetTerminal Terminal) (err error) {
	var terminalId string
	terminalId = targetTerminal.Id

	//Make sure Terminal struct is ready for use
	if len(terminalId) == 0 {
		log.Fatal("Terminal %v missing Id.\n", targetTerminal.Title)
		return
	}

	//Request Albums edge from Graph API
	//Try to find 72 hour album id. If no 72 hour album found, use terminal id.
	var albumId string
	if albumId, err = find72HrAlbumId(targetTerminal); err != nil {
		return
	}
	if len(albumId) == 0 {
		albumId = targetTerminal.Id
		displayErrorForTerminal(targetTerminal, "72 hour album not found.")
	} else {
		displayMessageForTerminal(targetTerminal, "72 hour album found.")
		incrementTerminalsWith72HRAlbum()
	}

	//Request Photos edge from Graph API
	var photosEdge PhotosEdge
	if photosEdge, err = getPhotosEdge(albumId); err != nil {
		return
	}

	//Look at the photo nodes returned by the photos edge
	var limit int
	limit = 4
	if len(photosEdge.Data) < limit {
		limit = len(photosEdge.Data)
	}

	//Spawn goroutine to download and process each image
	ctx := context.TODO()
	var maxWorkers int
	var sem *semaphore.Weighted
	maxWorkers = runtime.GOMAXPROCS(0)
	//maxWorkers = 1
	sem = semaphore.NewWeighted(int64(maxWorkers))

	displayMessageForTerminal(targetTerminal, fmt.Sprintf("Starting update with %v workers.", maxWorkers))

	var errorCount, flightsFound int

	for photoIndex := 0; photoIndex < limit; photoIndex++ {
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Printf("Failed to acquire semaphore: %v\n", err)
			break
		}

		go func(edgePhoto PhotosEdgePhoto, t Terminal) {
			defer sem.Release(1)
			var flightsFoundInPhoto int
			var err error
			if flightsFoundInPhoto, err = processPhotoNode(edgePhoto, t); err != nil {
				displayErrorForTerminal(t, err.Error())
				errorCount++
			}

			flightsFound += flightsFoundInPhoto
		}(photosEdge.Data[photoIndex], targetTerminal)
	}

	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		log.Printf("Failed to acquire semaphore: %v\n", err)
	}

	if errorCount == 0 {
		incrementNoErrorTerminals()
	}

	if flightsFound > 0 {
		incrementFoundFlightsTerminals()
	}

	return
}

//Find 72Hr Album Id for Terminal
func find72HrAlbumId(t Terminal) (albumId string, err error) {
	//Create request url and parameters
	apiUrl := GRAPH_API_URL
	resource := fmt.Sprintf("%v/%v/%v", GRAPH_API_VERSION, t.Id, GRAPH_EDGE_ALBUMS)
	data := url.Values{}
	data.Add(GRAPH_ACCESS_TOKEN_KEY, GRAPH_ACCESS_TOKEN)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := fmt.Sprintf("%v?%v", u, data.Encode())

	//log.Println(urlStr)

	//Create request
	var req *http.Request
	var client *http.Client
	client = &http.Client{
		Timeout: time.Second * 20}
	if req, err = http.NewRequest("GET", urlStr, nil); err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	//Make request
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}

	//Read response body into []byte
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		//"Error reading page photos edge resp."
		return
	}

	//Unmarshall into struct
	var pageAlbumsEdge AlbumsEdge
	if err = json.Unmarshal(body, &pageAlbumsEdge); err != nil {
		//"Error unmarshaling page photos edge."
		return
	}
	//fmt.Printf("%v\n", string(body))

	//Check for error
	if pageAlbumsEdge.Error.Code != 0 {
		err = fmt.Errorf("Code %v\nMessage %v", pageAlbumsEdge.Error.Code, pageAlbumsEdge.Error.Message)
		return
	}

	var Hr72Regex *regexp.Regexp
	//Match Date Month Year. Capture date and year
	if Hr72Regex, err = regexp.Compile("(?i)72.*hour"); err != nil {
		return
	}
	for _, album := range pageAlbumsEdge.Data {
		if regexResult := Hr72Regex.FindStringSubmatch(album.Name); regexResult != nil {
			//Found match
			albumId = album.Id
			return
		}
	}
	return
}

//Request PhotosEdge from Graph API
func getPhotosEdge(id string) (photosEdge PhotosEdge, err error) {
	//Create request url and parameters
	apiUrl := GRAPH_API_URL
	resource := fmt.Sprintf("%v/%v/%v", GRAPH_API_VERSION, id, GRAPH_EDGE_PHOTOS)
	data := url.Values{}
	data.Add(GRAPH_TYPE_KEY, GRAPH_TYPE_UPLOADED_KEY)
	data.Add(GRAPH_ACCESS_TOKEN_KEY, GRAPH_ACCESS_TOKEN)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := fmt.Sprintf("%v?%v", u, data.Encode())

	//log.Println(urlStr)

	//Create request
	var req *http.Request
	var client *http.Client
	client = &http.Client{
		Timeout: time.Second * 20}
	if req, err = http.NewRequest("GET", urlStr, nil); err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	//Make request
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}

	//Read response body into []byte
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		//"Error reading page photos edge resp."
		return
	}

	//Unmarshall into struct
	if err = json.Unmarshal(body, &photosEdge); err != nil {
		//"Error unmarshaling page photos edge."
		return
	}

	//displayMessageForTerminal(targetTerminal, "Read page photos edge.")

	//fmt.Printf("%v\n", string(body))

	//Check for error
	if photosEdge.Error.Code != 0 {
		err = fmt.Errorf("Code %v\nMessage %v", photosEdge.Error.Code, photosEdge.Error.Message)
		return
	}

	return
}

//Download, save, OCR a photo from Photos Edge
func processPhotoNode(edgePhoto PhotosEdgePhoto, targetTerminal Terminal) (flightsFound int, err error) {

	//Check if photo created within X timeframe (made recently?)
	var photoCreatedTime time.Time
	//http://stackoverflow.com/questions/24401901/time-parse-why-does-golang-parses-the-time-incorrectly
	layout := "2006-01-02T15:04:05-0700"
	if photoCreatedTime, err = time.Parse(layout, edgePhoto.CreatedTime); err != nil {
		return
	}

	//If image is too old, ignore
	if time.Since(photoCreatedTime) > time.Hour*24 {
		displayMessageForTerminal(targetTerminal, edgePhoto.Id+" over 24 hours old.")
		return
	}

	incrementPhotosFound()

	//displayMessageForTerminal(targetTerminal, fmt.Sprintf("Downloading recent photo %v",photoIndex+1))

	var saveTypes []SaveImageType
	saveTypes = []SaveImageType{SAVE_IMAGE_TRAINING, SAVE_IMAGE_TRAINING_PROCESSED_BLACK, SAVE_IMAGE_TRAINING_PROCESSED_WHITE}

	tmpSlide := Slide{
		SaveType:      saveTypes[0],
		Terminal:      targetTerminal,
		FBNodeId:      edgePhoto.Id,
		FBCreatedTime: time.Time{}}

	//Request Photo node for slide
	var photoNode PhotoNode
	if photoNode, err = getPhotoNodeForSlide(tmpSlide); err != nil {
		return
	}

	//Download and Save Image for Photo node
	if err = downloadAndSaveImageForPhotoNode(photoNode, &tmpSlide); err != nil {
		return
	}

	var slides []Slide
	slides = make([]Slide, 0)
	for _, currentSaveType := range saveTypes {
		var newSlide Slide
		newSlide.SaveType = currentSaveType
		newSlide.Extension = tmpSlide.Extension
		newSlide.Terminal = targetTerminal
		newSlide.FBNodeId = edgePhoto.Id
		newSlide.FBCreatedTime = photoCreatedTime

		//Manual slide control
		//newSlide.Extension = "jpeg"
		//newSlide.FBNodeId = "1630035233732546"
		//newSlide.FBNodeId = "1600297960039607"
		//newSlide.FBNodeId = "1600298003372936"

		//create processed image in imagemagick IF slide created is not original slide
		if currentSaveType != SAVE_IMAGE_TRAINING {
			if err = runImageMagickColorProcess(SAVE_IMAGE_TRAINING, newSlide); err != nil {
				return
			}
		}

		if err = doOCRForSlide(&newSlide, OCR_WHITELIST_NORMAL); err != nil {
			return
		}

		slides = append(slides, newSlide)
	}

	//Find date displayed in photo. Pick best date from slides.
	var slideDate time.Time
	if slideDate, err = findDateOfPhotoNodeSlides(slides); err != nil {
		return
	}

	//Display found date
	displayMessageForTerminal(slides[0].Terminal, fmt.Sprintf("%v found date for photo node \u001b[1m\u001b[31m%v\u001b[0m", slides[0].FBNodeId, slideDate.Format("02 Jan 2006 -0700")))

	//Get dest bbox
	//TODO: Find bounds of destination column and crop to get better OCR results.
	var destLabelBBox image.Rectangle
	if destLabelBBox, err = findLabelBoundsOfPhotoNodeSlides(slides, KEYWORD_DESTINATION); err != nil {
		return
	}

	//Find potential rollcall times from all slides
	var rollCalls []RollCall
	var rollCallsNoBBox []RollCall

	//Check if Destination label is over 0.5 (threshold constant) of total image height. If too low on image, Destination Label may be incorrect.
	var destLabelValid bool
	if destLabelValid, err = slides[0].isYCoordinateInHeightPercentage(destLabelBBox.Min.Y, DESTINATION_TEXT_VERTICAL_THRESHOLD); err != nil {
		return
	}
	if !destLabelValid { //
		destLabelBBox.Min.Y = 0
	}

	if rollCalls, rollCallsNoBBox, err = findRollCallTimesFromSlides(slides, slideDate, destLabelBBox.Min.Y); err != nil {
		return
	}

	//Print any not found in hOCR rollcalls
	if len(rollCallsNoBBox) > 0 {
		fmt.Println("rollcalls w/o bbox", rollCallsNoBBox)
	}

	//Get seats bbox
	var seatsLabelBBox image.Rectangle
	if seatsLabelBBox, err = findLabelBoundsOfPhotoNodeSlides(slides, KEYWORD_SEATS); err != nil {
		return
	}

	//Find potential seats
	var seatsAvailable []SeatsAvailable
	if seatsAvailable, err = findSeatsAvailableFromSlides(slides, seatsLabelBBox); err != nil {
		return
	}

	//Find potential destinations from all slides
	var destinations []Destination
	if destinations, err = findDestinationsFromSlides(slides, destLabelBBox.Min.Y); err != nil {
		return
	}

	deleteTerminalFromDestArray(&destinations, slides[0].Terminal)

	/*
		//Print destination object. Shows spelling found and distance.
		fmt.Println("found dests in all slides")
		for _, d := range destinations {
			log.Println(d)
		}
		//return
	*/

	//Find vertically closest Destination for every RollCall
	linkRollCallsToNearestDestinations(rollCalls, destinations)

	/*
		fmt.Println("After nearest rc to dest link")
		for _, d := range destinations {
			log.Println(d)
			if d.LinkedRollCall != nil {
				log.Println(*(d.LinkedRollCall))
			}
		}
	*/

	//Link SeatsAvailable with RollCalls on same line
	linkRollCallsToNearestSeatsAvailable(rollCalls, seatsAvailable)

	/*
		fmt.Println("After link seats")
		for _, rc := range rollCalls {
			log.Println(rc)
			if rc.LinkedSeatsAvailable != nil {
				log.Println(*(rc.LinkedSeatsAvailable))
			}
		}
	*/

	//Find vertically closest Destination for every SeatsAvailable
	linkDestinationsToNearestSeatsAvailable(destinations, seatsAvailable)

	//Create array of individual Grouping for each Destination to pass into combine Destinations to Groupings stage
	var destinationGroupings []Grouping
	for _, d := range destinations {
		destinationGroupings = append(destinationGroupings, Grouping{
			Destinations:   []Destination{d},
			LinkedRollCall: d.LinkedRollCall})
		destinationGroupings[len(destinationGroupings)-1].updateBBox()
	}

	//Link GroupingA with nearest GroupingB so all Destinations in GroupingB are in GroupingA. Repeat until GroupingA contains a Destination that (horizontally) intersects a RollCall.
	combineDestinationGroupsToAnchorDestinations(&destinationGroupings)

	//Print out found flights
	fmt.Println("After combine dests:")
	for _, dg := range destinationGroupings {
		if dg.LinkedRollCall != nil {
			fmt.Printf("%v - ", (*dg.LinkedRollCall).Time.Format("\u001b[1m\u001b[35m‣ 02JAN2006 1504 MST -0700\u001b[0m"))

			if (*dg.LinkedRollCall).LinkedSeatsAvailable != nil {
				fmt.Printf("%v%v\n", (*(*dg.LinkedRollCall).LinkedSeatsAvailable).Number, (*(*dg.LinkedRollCall).LinkedSeatsAvailable).Letter)
			} else {
				fmt.Println("No seat text found.")
			}
		} else {
			fmt.Println("No time for grouping.")
		}

		for _, d := range dg.Destinations {
			fmt.Println(d.TerminalTitle)
		}
	}

	//Create list of flights,
	//Initially Destination is a flight.
	//Structure: flight -> Destination -> LinkedRollCall * -> LinkedSeatsAvailable *
	//Convert Destinations to Flight struct and add
	var finalFlights []Flight
	for dgIndex, _ := range destinationGroupings {
		//Link RollCall to all Destinations in Grouping (if RollCall linked)
		//Add to each Destination final flights list
		for dIndex, _ := range destinationGroupings[dgIndex].Destinations {
			if destinationGroupings[dgIndex].LinkedRollCall != nil {
				destinationGroupings[dgIndex].Destinations[dIndex].LinkedRollCall = destinationGroupings[dgIndex].LinkedRollCall

				//Link Destination to Grouping.LinkedRollCall.LinkedSeatsAvailable if no Destination linked seats already
				if destinationGroupings[dgIndex].Destinations[dIndex].LinkedSeatsAvailable == nil {
					destinationGroupings[dgIndex].Destinations[dIndex].LinkedSeatsAvailable = (*destinationGroupings[dgIndex].LinkedRollCall).LinkedSeatsAvailable
				}
			}

			/*
				//Do not add any destinations without roll calls
				if destinationGroupings[dgIndex].Destinations[dIndex].LinkedRollCall == nil {
					continue
				}
			*/

			//Set Flight.UnknownRollCallDate if applicable
			//Check if date was NOT found by comparing to "blank" time.Time
			var unknownRCDate bool
			if slideDate.Equal(time.Time{}) {
				unknownRCDate = true
			} else {
				incrementPhotosFoundDateHeader()
			}

			//Create Flight struct to add to slice
			tmpFlight := Flight{
				Origin:      slides[0].Terminal.Title,
				Destination: destinationGroupings[dgIndex].Destinations[dIndex].TerminalTitle,

				UnknownRollCallDate: unknownRCDate,
				PhotoSource:         slides[0].FBNodeId,
				SourceDate:          slides[0].FBCreatedTime}

			if destinationGroupings[dgIndex].Destinations[dIndex].LinkedRollCall != nil {
				tmpFlight.RollCall = (*destinationGroupings[dgIndex].Destinations[dIndex].LinkedRollCall).Time
			}

			if destinationGroupings[dgIndex].Destinations[dIndex].LinkedSeatsAvailable != nil {
				tmpFlight.SeatCount = (*destinationGroupings[dgIndex].Destinations[dIndex].LinkedSeatsAvailable).Number
				tmpFlight.SeatType = (*destinationGroupings[dgIndex].Destinations[dIndex].LinkedSeatsAvailable).Letter
			}

			finalFlights = append(finalFlights, tmpFlight)
		}
	}

	/*
		//Print flights list for photo
		displayMessageForTerminal(slides[0].Terminal, fmt.Sprintf("%v Flights list for photo node %v", slides[0].FBNodeId))
		for _, ff := range finalFlights {
			fmt.Println(ff)
		}
	*/

	if err = deleteFlightsFromTableForDayForOriginTerminal(FLIGHTS_72HR_TABLE, slideDate, slides[0].Terminal); err != nil {
		return
	}

	if err = insertFlightsIntoTable(FLIGHTS_72HR_TABLE, finalFlights); err != nil {
		return
	}

	flightsFound = len(finalFlights)
	incrementPhotosProcessed()
	/*
		//Debugging print slides array
		log.Printf("len(saveTypes) %v", len(saveTypes))
		log.Printf("len(slides) %v", len(slides))
		for _, s := range slides {
			log.Printf("slide type %v", s.saveType)
		}
	*/

	return
}

//Request Photo node for Slide (info from Photo edge).
func getPhotoNodeForSlide(sReference Slide) (photoNode PhotoNode, err error) {
	//Create request url and parameters
	apiUrl := GRAPH_API_URL
	resource := fmt.Sprintf("%v/%v", GRAPH_API_VERSION, sReference.FBNodeId)
	data := url.Values{}
	data.Add(GRAPH_FIELDS_KEY, GRAPH_FIELD_IMAGES_KEY)
	data.Add(GRAPH_ACCESS_TOKEN_KEY, GRAPH_ACCESS_TOKEN)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := fmt.Sprintf("%v?%v", u, data.Encode())

	//Create request
	var req *http.Request
	var client *http.Client
	client = &http.Client{
		Timeout: time.Second * 20}
	if req, err = http.NewRequest("GET", urlStr, nil); err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	//Make request
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}

	//Read response body into []byte
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	//Unmarshall into struct
	if err = json.Unmarshal(body, &photoNode); err != nil {
		return
	}

	//Check for error
	if photoNode.Error.Code != 0 {
		err = fmt.Errorf("Code %v\nMessage %v", photoNode.Error.Code, photoNode.Error.Message)
		return
	}

	return
}

//Download first image for Photo node to IMAGE_TMP_DIRECTORY and Save in location for Slide.
//Sets Extension for Slide based on http.DetectContentType()
func downloadAndSaveImageForPhotoNode(photoNode PhotoNode, sReference *Slide) (err error) {

	if len(photoNode.Images) == 0 {
		err = errors.New(fmt.Sprintf("PhotoNode %v %v has no images.", (*sReference).Terminal.Title, (*sReference).FBNodeId))
		return
	}

	//Create tmp directory if needed
	var exist bool
	if exist, err = exists(IMAGE_TMP_DIRECTORY); exist == false || err != nil {
		if err != nil {
			return
		}
		if err = os.Mkdir(IMAGE_TMP_DIRECTORY, os.ModePerm); err != nil {
			return
		}
		return
	}

	//Create file handle to tmp photo path
	tmpFilepath := fmt.Sprintf("%v/%v", IMAGE_TMP_DIRECTORY, (*sReference).FBNodeId)
	var tmpFile *os.File
	if tmpFile, err = os.Create(tmpFilepath); err != nil {
		return
	}
	defer tmpFile.Close()

	//Create request
	var req *http.Request
	var client *http.Client
	client = &http.Client{
		Timeout: time.Second * 20}
	if req, err = http.NewRequest("GET", photoNode.Images[0].Source, nil); err != nil {
		return
	}

	//Make request
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}

	//Read response body into os.File
	defer resp.Body.Close()
	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		return
	}

	//Open tmp file for reading and read first 512 bytes for http.DetectContentType()

	var fileHeader []byte
	fileHeader = make([]byte, 512)
	if tmpFile, err = os.Open(tmpFilepath); err != nil {
		return
	}

	defer tmpFile.Close()
	var length int
	if length, err = tmpFile.Read(fileHeader); err != nil && err != io.EOF {
		return
	}
	if length < len(fileHeader) {
		fileHeader = fileHeader[:length]
	}

	var detectedContentType string
	detectedContentType = http.DetectContentType(fileHeader)
	if detectedContentType == "image/png" {
		(*sReference).Extension = "png"
	} else if detectedContentType == "image/jpeg" {
		(*sReference).Extension = "jpeg"
	} else if detectedContentType == "image/gif" {
		(*sReference).Extension = "gif"
	} else {
		err = fmt.Errorf("Unhandled MIME type %v detected.", detectedContentType)
		return
	}

	if err = copyFileContents(tmpFilepath, photoPath(*sReference)); err != nil {
		return
	}
	return
}
