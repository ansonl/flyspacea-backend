package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
	//"sync"

	//worker management
	"context"
	"golang.org/x/sync/semaphore"
	"regexp"
	"runtime"
)

func readTerminalFileToArray(terminalFilename string) (terminalArray []Terminal, err error) {
	terminalsRaw, err := ioutil.ReadFile(terminalFilename)
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

func updateAllTerminals(terminalMap map[string]Terminal) {
	ctx := context.TODO()

	var maxWorkers int
	var sem *semaphore.Weighted
	maxWorkers = runtime.GOMAXPROCS(0)
	maxWorkers = 1
	sem = semaphore.NewWeighted(int64(maxWorkers))

	log.Printf("Starting update with %v workers.\n", maxWorkers)

	for _, v := range terminalMap {
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Printf("Failed to acquire semaphore: %v\n", err)
			break
		}

		go func(t Terminal) {
			defer sem.Release(1)
			if err := updateTerminal(t); err != nil {
				displayErrorForTerminal(t, err.Error())
			}
		}(v)
	}

	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		log.Printf("Failed to acquire semaphore: %v\n", err)
	}

	log.Printf("Update ended.\n")

	displayStatistics()
}

func updateTerminal(targetTerminal Terminal) (err error) {
	var terminalId string
	terminalId = targetTerminal.Id

	if len(terminalId) == 0 {
		log.Fatal("Terminal %v missing Id.\n", targetTerminal.Title)
		return
	} else {
		displayMessageForTerminal(targetTerminal, "Requesting page photos edge.")
	}

	//Try to find 72 hour album id. If no 72 hour album found, use terminal id.
	var albumId string
	if albumId, err = find72HrAlbumId(targetTerminal); err != nil {
		return
	}
	if len(albumId) == 0 {
		albumId = targetTerminal.Id
		displayMessageForTerminal(targetTerminal, "72 hour album not found.")
	} else {
		displayMessageForTerminal(targetTerminal, "72 hour album found.")
	}

	//Create request url and parameters
	apiUrl := GRAPH_API_URL
	resource := fmt.Sprintf("%v/%v/%v", GRAPH_API_VERSION, albumId, GRAPH_EDGE_PHOTOS)
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
	var pagePhotosEdge PhotosEdge
	if err = json.Unmarshal(body, &pagePhotosEdge); err != nil {
		//"Error unmarshaling page photos edge."
		return
	}

	//displayMessageForTerminal(targetTerminal, "Read page photos edge.")

	//fmt.Printf("%v\n", string(body))

	//Check for error
	if pagePhotosEdge.Error.Code != 0 {
		err = fmt.Errorf("Code %v\nMessage %v", pagePhotosEdge.Error.Code, pagePhotosEdge.Error.Message)
		return
	}

	//Look at the photo nodes returned by the photos edge
	var limit int
	limit = 4
	if len(pagePhotosEdge.Data) < limit {
		limit = len(pagePhotosEdge.Data)
	}

	for photoIndex := 0; photoIndex < limit; photoIndex++ {
		incrementPhotosFound()
		photo := pagePhotosEdge.Data[photoIndex]

		//Check if photo created within X timeframe (made recently?)
		var photoCreatedTime time.Time
		//http://stackoverflow.com/questions/24401901/time-parse-why-does-golang-parses-the-time-incorrectly
		layout := "2006-01-02T15:04:05-0700"
		if photoCreatedTime, err = time.Parse(layout, photo.CreatedTime); err != nil {
			return
		}

		//If image is too old, ignore
		if time.Since(photoCreatedTime) > time.Hour*72 {
			continue
		}

		//displayMessageForTerminal(targetTerminal, fmt.Sprintf("Downloading recent photo %v",photoIndex+1))

		var saveTypes []SaveImageType
		saveTypes = []SaveImageType{SAVE_IMAGE_TRAINING, SAVE_IMAGE_TRAINING_PROCESSED_BLACK, SAVE_IMAGE_TRAINING_PROCESSED_WHITE}

		tmpSlide := Slide{saveTypes[0], "", "", targetTerminal, photo.Id, "", ""}

		//Download, save
		if err = downloadAndSaveSlide(&tmpSlide); err != nil {
			return
		}

		var slides []Slide
		slides = make([]Slide, 0)
		for _, currentSaveType := range saveTypes {
			var newSlide Slide
			newSlide.SaveType = currentSaveType
			newSlide.Extension = tmpSlide.Extension
			newSlide.Terminal = targetTerminal
			newSlide.FBNodeId = photo.Id

			/*
				//Manual slide control
				newSlide.Extension = "jpeg"
				newSlide.FBNodeId = "1579091732160230"
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

		displayMessageForTerminal(slides[0].Terminal, fmt.Sprintf("%v \u001b[1m\u001b[31m%v\u001b[0m", slides[0].FBNodeId, slideDate.Format("02 Jan 2006 -0700")))

		if slideDate.Equal(time.Time{}) == false {
			incrementPhotosFoundDateHeader()
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
			//Find closest spelling for KEYWORD_DESTINATION
			var closestDestinationSpelling string
			var closestDestinationSlide Slide
			if closestDestinationSpelling, closestDestinationSlide, err = findKeywordClosestSpellingInPhotoInSaveImageTypes(KEYWORD_DESTINATION, slides); err != nil {
				return
			}

			if len(closestDestinationSpelling) == 0 {
				displayMessageForTerminal(targetTerminal, fmt.Sprintf("No close dest spelling founds"));
			} else {
				displayMessageForTerminal(targetTerminal, fmt.Sprintf("Closest dest spelling %v in saveType %v", closestDestinationSpelling, closestDestinationSlide.saveType));
			}

			//Find KEYWORD_DESTINATION bounds in hOCR
			bbox, err := getDestinationBounds(closestDestinationSlide.hHOCRText, closestDestinationSpelling)
			if err != nil {
				log.Println(err)
			}

			displayMessageForTerminal(targetTerminal, fmt.Sprintf("%v bbox %v %v %v %v", photoIndex, bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y))
		*/
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

//Download Photo Node from Photos Edge. Return error.
func downloadAndSaveSlide(sReference *Slide) (err error) {
	//Create request url and parameters
	apiUrl := GRAPH_API_URL
	resource := fmt.Sprintf("%v/%v", GRAPH_API_VERSION, (*sReference).FBNodeId)
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
	var photo PhotoNode
	if err = json.Unmarshal(body, &photo); err != nil {
		return
	}

	//Check for error
	if photo.Error.Code != 0 {
		err = fmt.Errorf("Code %v\nMessage %v", photo.Error.Code, photo.Error.Message)
		return
	}

	if err = downloadAndSavePhotoNode(photo, sReference); err != nil {
		return
	}

	return
}

//Download Photo Node. Return error.
func downloadAndSavePhotoNode(photoNode PhotoNode, sReference *Slide) (err error) {

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
