package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
	"errors"
	//"sync"

	//worker management
	"context"
    "runtime"
    "golang.org/x/sync/semaphore"
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
	for i, v := range terminalArray {
		v.OffsetUp = (len(terminalArray) - 1) - i
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
}

func updateTerminal(targetTerminal Terminal) (err error) {
	terminalId := targetTerminal.Id
	if len(terminalId) == 0 {
		log.Fatal("Terminal %v missing Id.\n", targetTerminal.Title)
		return
	} else {
		displayMessageForTerminal(targetTerminal, "Requesting page photos edge.")
	}

	//Create request url and parameters
	apiUrl := GRAPH_API_URL
	resource := fmt.Sprintf("%v/%v/%v", GRAPH_API_VERSION, terminalId, GRAPH_EDGE_PHOTOS)
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

	displayMessageForTerminal(targetTerminal, "Read page photos edge.")

	//fmt.Printf("%v\n", string(body))

	//Look at the photo nodes returned by the photos edge
	
	var limit int
	limit = 5
	if len(pagePhotosEdge.Data) < limit {
		limit = len(pagePhotosEdge.Data)
	}

	for photoIndex := 0;photoIndex < limit;photoIndex++ {
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
		saveTypes = []SaveImageType {SAVE_IMAGE_TRAINING, SAVE_IMAGE_TRAINING_PROCESSED_BLACK, SAVE_IMAGE_TRAINING_PROCESSED_WHITE}

		tmpSlide := Slide{saveTypes[0], targetTerminal, photo.Id, "", ""}

		//Download, save
		if err = downloadAndSaveSlide(tmpSlide); err != nil {
			return
		}

		var slides []Slide
		slides = make([]Slide, 0)
		for _, currentSaveType := range saveTypes {
			var newSlide Slide
			newSlide.saveType = currentSaveType
			newSlide.terminal = targetTerminal
			newSlide.fbNodeId = photo.Id

			//create processed image in imagemagick IF slide created is not original slide
			if currentSaveType != SAVE_IMAGE_TRAINING {
				if err = runImageMagickConvert(SAVE_IMAGE_TRAINING, newSlide); err != nil {
					return
				}
			}

			if err = doOCRForSlide(&newSlide); err != nil {
				return
			}

			slides = append(slides, newSlide)
		}

		/*
		//Debugging print slides array
		log.Printf("len(saveTypes) %v", len(saveTypes))
		log.Printf("len(slides) %v", len(slides))
		for _, s := range slides {
			log.Printf("slide type %v", s.saveType)
		}
		*/

		//Find date of 72 hour slide
		
		
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
		bbox, err := getDestinationBounds(closestDestinationSlide.hOCRText, closestDestinationSpelling)
		if err != nil {
			log.Println(err)
		}

		displayMessageForTerminal(targetTerminal, fmt.Sprintf("%v bbox %v %v %v %v", photoIndex, bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y))
		*/
		

		/*
		   //Specify the methods to use
		   havenThing := HavenMethod{OCRMethod{"Haven", photo, nil, false}}
		   tesseractThing := TesseractMethod{OCRMethod{"Tesseract", photo, nil, false}}
		   //Forgot about the need to pass pointer to obj and that method receiver is passed by value and not reference so obj method that modifies needs pointer to actually affect underlying obj. http://jordanorelli.com/post/32665860244/how-to-use-interfaces-in-go
		   //methodsToUse := []OCRMethodInterface{&havenThing}
		   methodsToUse := []OCRMethodInterface{&tesseractThing, &havenThing}

		   var updatedFlightArray [][]Departure
		   updatedFlightArray = make([][]Departure, len(methodsToUse))

		   var wg sync.WaitGroup

		   for i, _ := range methodsToUse {
		       wg.Add(1)

		       go func(targetMethod OCRMethodInterface, targetFlightArray []Departure) {
		           defer wg.Done()
		           targetMethod.doOCR()
		           targetFlightArray = targetMethod.getTargetDeparturesArray()

		           for j, _ := range targetFlightArray {
		               targetFlightArray[j].Origin = targetTerminal.Title
		               //fmt.Printf("%q\n",updatedFlightArray[i][j])
		           }
		       }(methodsToUse[i], updatedFlightArray[i])
		   }


		   wg.Wait()
		*/

	}

	return
}

//Download Photo Node from Photos Edge. Return error.
func downloadAndSaveSlide(sReference Slide) (err error) {
	//Create request url and parameters
	apiUrl := GRAPH_API_URL
	resource := fmt.Sprintf("%v/%v", GRAPH_API_VERSION, sReference.fbNodeId)
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

	if err = downloadAndSavePhotoNode(photo, sReference); err != nil {
		return
	}

	return
}

//Download Photo Node. Return error.
func downloadAndSavePhotoNode(photoNode PhotoNode, sReference Slide) (err error) {

	if len(photoNode.Images) == 0 {
		err = errors.New(fmt.Sprintf("PhotoNode %v %v has no images.", sReference.terminal.Title, sReference.fbNodeId))
		return
	}

	if exist, _ := exists(IMAGE_TRAINING_DIRECTORY); exist == false {
		err = errors.New("Directory does not exist")
		return
	}

	//Create file handle to correct photo path
	var out *os.File
	if out, err = os.Create(photoPath(sReference)); err != nil {
		return
	}
	defer out.Close()

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
	if _, err = io.Copy(out, resp.Body); err != nil {
		return
	}

	return
}