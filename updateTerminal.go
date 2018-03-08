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
	"time"

	//Worker management
	"context"
	"golang.org/x/sync/semaphore"
	"regexp"
	"runtime"
)

func updateAllTerminals(terminalMap map[string]Terminal) {
	for _, v := range terminalMap {
		if err := updateTerminal(v); err != nil {
			displayErrorForTerminal(v, err.Error())
		}
	}

	log.Printf("Update ended.\n")

	displayStatistics()
}

func updateTerminal(targetTerminal Terminal) (err error) {
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
	}

	//Request Photos edge from Graph API
	var photosEdge PhotosEdge
	if photosEdge, err = getPhotosEdge(albumId); err != nil {
		return
	}

	//Look at the photo nodes returned by the photos edge
	var limit int
	limit = 5
	if len(photosEdge.Data) < limit {
		limit = len(photosEdge.Data)
	}

	//Spawn goroutine to download and process each image
	ctx := context.TODO()
	var maxWorkers int
	var sem *semaphore.Weighted
	maxWorkers = runtime.GOMAXPROCS(0)
	maxWorkers = 1
	sem = semaphore.NewWeighted(int64(maxWorkers))

	displayMessageForTerminal(targetTerminal, fmt.Sprintf("Starting update with %v workers.", maxWorkers))

	for photoIndex := 0; photoIndex < limit; photoIndex++ {
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Printf("Failed to acquire semaphore: %v\n", err)
			break
		}

		go func(edgePhoto PhotosEdgePhoto, t Terminal) {
			defer sem.Release(1)
			if err := processPhotoNode(edgePhoto, t); err != nil {
				displayErrorForTerminal(t, err.Error())
			}
		}(photosEdge.Data[photoIndex], targetTerminal)
	}

	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		log.Printf("Failed to acquire semaphore: %v\n", err)
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
	client = &http.Client{}
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
	client = &http.Client{}
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
func processPhotoNode(edgePhoto PhotosEdgePhoto, targetTerminal Terminal) (err error) {
	incrementPhotosFound()

	//Check if photo created within X timeframe (made recently?)
	var photoCreatedTime time.Time
	//http://stackoverflow.com/questions/24401901/time-parse-why-does-golang-parses-the-time-incorrectly
	layout := "2006-01-02T15:04:05-0700"
	if photoCreatedTime, err = time.Parse(layout, edgePhoto.CreatedTime); err != nil {
		return
	}

	//If image is too old, ignore
	if time.Since(photoCreatedTime) > time.Hour*96 {
		displayMessageForTerminal(targetTerminal, edgePhoto.Id+" over 96 hours old.")
		return
	}

	//displayMessageForTerminal(targetTerminal, fmt.Sprintf("Downloading recent photo %v",photoIndex+1))

	var saveTypes []SaveImageType
	saveTypes = []SaveImageType{SAVE_IMAGE_TRAINING, SAVE_IMAGE_TRAINING_PROCESSED_BLACK, SAVE_IMAGE_TRAINING_PROCESSED_WHITE}

	tmpSlide := Slide{saveTypes[0], "", "", targetTerminal, edgePhoto.Id, "", ""}

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

		/*
		//Manual slide control
		newSlide.Extension = "jpeg"
		newSlide.FBNodeId = "1600297943372942"
		*/

		//create processed image in imagemagick IF slide created is not original slide
		if currentSaveType != SAVE_IMAGE_TRAINING {
			if err = runImageMagickColorProcess(SAVE_IMAGE_TRAINING, newSlide); err != nil {
				return
			}
		}

		if err = doOCRForSlide(&newSlide); err != nil {
			return
		}

		slides = append(slides, newSlide)
	}

	var slideDate time.Time
	if slideDate, err = findDateOfPhotoNodeSlides(slides); err != nil {
		return
	}

	//Check if date was found by comparing to "blank" time.Time
	if slideDate.Equal(time.Time{}) == false {
		incrementPhotosFoundDateHeader()
	}

	//Display found date
	displayMessageForTerminal(slides[0].Terminal, fmt.Sprintf("%v \u001b[1m\u001b[31m%v\u001b[0m", slides[0].FBNodeId, slideDate.Format("02 Jan 2006 -0700")))

	//Get dest bbox and crop
	//TODO?
	var destLabelBBox image.Rectangle
	if destLabelBBox, err = findDestinationLabelBoundsOfPhotoNodeSlides(slides); err != nil {
		return
	}

	displayMessageForTerminal(slides[0].Terminal, fmt.Sprintf(" dest bbox %v %v %v %v", destLabelBBox.Min.X, destLabelBBox.Min.Y, destLabelBBox.Max.X, destLabelBBox.Max.Y))

	//Find potential destinations from all slides
	var destinations []Destination
	if destinations, err = findDestinationsFromSlides(slides); err != nil {
		return
	}

	fmt.Println(destinations)

	for _, d := range destinations {
		/*
			yTolerance := 5
			if d.BBox.Min.Y > destLabelBBox.Max.Y || destLabelBBox.Max.Y - d.BBox.Min.Y < yTolerance {
				log.Println(d)
			}
		*/

		log.Println(d)
	}

	incrementPhotosProcessed()
	/*
		//Debugging print slides array
		log.Printf("len(saveTypes) %v", len(saveTypes))
		log.Printf("len(slides) %v", len(slides))
		for _, s := range slides {
			log.Printf("slide type %v", s.saveType)
		}
	*/

	/*

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
	client = &http.Client{}
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
	client = &http.Client{}
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
