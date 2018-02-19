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
	"strings"
	"time"
	"errors"
	//"sync"
)

type PhotosEdge struct {
	Data []PhotosEdgePhoto `json:"data"`
}

type PhotosEdgePhoto struct {
	CreatedTime string `json:"created_time"`
	Name        string `json:"name"`
	Id          string `json:"id"`
}

type PhotoNode struct {
	Images []PhotoNodeImage `json:"images"`
}

type PhotoNodeImage struct {
	Source string `json:"source"`
}

type TextBlock struct {
	Text string `json:"text"`
}

type Departure struct {
	RollCall    time.Time `json:"rollCall"`
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
	SeatCount   int       `json:"seatCount"`
	SeatType    string    `json:"seatType"`
	Canceled    bool      `json:"canceled"`
	PhotoSource string    `json:"photoSource"`
}

type Terminal struct {
	Title string `json:"title"`
	Id    string `json:"id"`
}

func readTerminalFileToMap(terminalFilename string) map[string]Terminal {
	terminalsRaw, readErr := ioutil.ReadFile(terminalFilename)
	if readErr != nil {
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
	if len(terminalId) == 0 {
		fmt.Printf("Terminal %v missing Id.\n", targetTerminal.Title)
		return
	} else {
		fmt.Printf("Starting %v %v\n", targetTerminal.Title, targetTerminal.Id)
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
	client := &http.Client{}
	r, _ := http.NewRequest("GET", urlStr, nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	//Make request
	resp, err := client.Do(r)
	if err != nil {
		log.Println(err)
	}

	//Read response body into []byte
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading page photos edge resp.")
	}

	//Unmarshall into struct
	pagePhotosEdge := PhotosEdge{}
	unmarshalErr := json.Unmarshal(body, &pagePhotosEdge)
	if unmarshalErr != nil {
		fmt.Println("Error unmarshaling page photos edge.")
	}
	//fmt.Println("Page photos edge retrieved.")

	/*
	   res2B, _ := json.Marshal(pagePhotos)
	   fmt.Println(string(res2B))
	*/

	//fmt.Printf("%v\n", string(body))

	//Look at the photo nodes returned by the photos edge
	var limit int
	limit = 5
	if len(pagePhotosEdge.Data) < limit {
		limit = len(pagePhotosEdge.Data)
	}

	var recentPhotos int
	for i := 0; i < limit; i++ {
		photo := pagePhotosEdge.Data[i]

		//Check if photo created within X timeframe (made recently?)
		var photoCreatedTime time.Time
		/* http://stackoverflow.com/questions/24401901/time-parse-why-does-golang-parses-the-time-incorrectly */
		layout := "2006-01-02T15:04:05-0700"
		photoCreatedTime, err := time.Parse(layout, photo.CreatedTime)
		if err != nil {
			log.Println(err)
		}
		if time.Since(photoCreatedTime) > time.Hour*72 {
			continue
		}

		recentPhotos += 1
		fmt.Printf("%v recent photos for %v\n", recentPhotos, targetTerminal.Title)

		//Download, save, create processed image in imagemagick
		downloadAndSavePhotosEdgePhoto(photo, targetTerminal, i)

		closestDestinationSpelling := findKeywordClosestSpellingInPhotoInSaveImageTypes(KEYWORD_DESTINATION, targetTerminal, i, []SaveImageType {SAVE_IMAGE_TRAINING, SAVE_IMAGE_TRAINING_PROCESSED})

		if len(closestDestinationSpelling) == 0 {
			log.Printf("No close dest spelling found %v\n", closestDestinationSpelling);
		} else {
			log.Printf("Closest dest spelling %v\n", closestDestinationSpelling);
		}

		

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
	fmt.Println()
}

func photoPath(saveType SaveImageType, prefix string, targetTerminal Terminal, photoNumber int) string {
	var photoDirectory string
	var photoFilename string
	switch (saveType) {
		case SAVE_IMAGE_TRAINING:
			photoDirectory = IMAGE_TRAINING_DIRECTORY
			break;
		case SAVE_IMAGE_TRAINING_PROCESSED:
			photoDirectory = IMAGE_TRAINING_PROCESSED_DIRECTORY
			break;
		default:
			log.Println("Unknown save type")
	}

	//prefix_terminal_title_n.png
	photoFilename = fmt.Sprintf("%v%v_%v.png", prefix, strings.Replace(strings.ToLower(targetTerminal.Title), " ", "_", -1), photoNumber)

	return fmt.Sprintf("%v/%v", photoDirectory, photoFilename)
}

func downloadAndSavePhotoNode(photoNode PhotoNode, targetTerminal Terminal, photoNumber int) error {
	if len(photoNode.Images) == 0 {
		return errors.New(fmt.Sprintf("PhotoNode %v %v has no images.", targetTerminal.Title, photoNumber))
	}

	if exist, _ := exists(IMAGE_TRAINING_DIRECTORY); exist == false {
		return errors.New("Directory does not exist")
	}

	//Create file handle
	out, err := os.Create(photoPath(SAVE_IMAGE_TRAINING, "", targetTerminal, photoNumber))
	defer out.Close()
	if err != nil {
		return err
	}

	//Create request
	client := &http.Client{}
	r, _ := http.NewRequest("GET", photoNode.Images[0].Source, nil)

	//Make request
	resp, err := client.Do(r)
	if err != nil {
		return err
	}

	//Read response body into []byte
	defer resp.Body.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

func downloadAndSavePhotosEdgePhoto(photosEdgePhoto PhotosEdgePhoto, targetTerminal Terminal, photoNumber int) {
	//Create request url and parameters
	apiUrl := GRAPH_API_URL
	resource := fmt.Sprintf("%v/%v", GRAPH_API_VERSION, photosEdgePhoto.Id)
	data := url.Values{}
	data.Add(GRAPH_FIELDS_KEY, GRAPH_FIELD_IMAGES_KEY)
	data.Add(GRAPH_ACCESS_TOKEN_KEY, GRAPH_ACCESS_TOKEN)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := fmt.Sprintf("%v?%v", u, data.Encode())

	//Create request
	client := &http.Client{}
	r, _ := http.NewRequest("GET", urlStr, nil)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	//Make request
	resp, err := client.Do(r)
	if err != nil {
		log.Println(err)
	}

	//Read response body into []byte
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	//Unmarshall into struct
	photo := PhotoNode{}
	unmarshalErr := json.Unmarshal(body, &photo)
	if unmarshalErr != nil {
		log.Println("Error unmarshaling photoNode")
		return
	}

	if err = downloadAndSavePhotoNode(photo, targetTerminal, photoNumber); err != nil {
		log.Println(err)
		return
	} else {
		runImageMagickConvert(targetTerminal, photoNumber)
	}



	/*


	hocr, err := getHOCRText(SAVE_IMAGE_TRAINING, targetTerminal, photoNumber)
	if err != nil {
		log.Println(err)
	}

	bbox, err := getDestinationBounds(hocr)
	if err != nil {
		log.Println(err)
	}

	log.Printf("bbox %v %v %v %v\n", bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y)
*/
	
}

func updateAllTerminals(terminalMap map[string]Terminal) {
	for _, v := range terminalMap {
		updateTerminal(v)
	}
}

func main() {
	createFuzzyModelsForKeywords([]string {KEYWORD_DESTINATION}, &fuzzyModelForKeyword)

	terminalMap := readTerminalFileToMap("terminals.json")
	updateAllTerminals(terminalMap)

	/*
	   v := Terminal{}
	   v.Title="Travis Pax Term"
	   v.Id="travispassengerterminal"
	   updateTerminal(v)
	*/

	return

	//fmt.Printf("%v\n",readTerminalFileToArray("terminals.json"))

}
