package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	//"time"
	"io/ioutil"
	"regexp"
	//"strings"
	"net/url"
	"bytes"
	"strconv"
)

type Photo struct {
	Source string `json:"source"`
}

type PhotosPhoto struct {
	CreatedTime string `json:"created_time"`
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

func main() {
	resp, err := http.Get("https://graph.facebook.com/v2.5/travispassengerterminal/photos?type=uploaded&access_token=522755171230853%7ChxS9OzJ4I0CqmmrESRpNHfx77vs")
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	//fmt.Println(string(body))

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

    var limit int
    limit = 3
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

    	photoNode := fmt.Sprintf("https://graph.facebook.com/v2.5/%v?fields=source&access_token=522755171230853%%7ChxS9OzJ4I0CqmmrESRpNHfx77vs", v.Id)
    	//fmt.Println(photoNode)
    	resp, err := http.Get(photoNode)
    	if err != nil {
			// handle error
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

    	photo := Photo{}
    	photoNodeErr := json.Unmarshal(body, &photo)
    	if (photoNodeErr != nil) {
    		fmt.Println("Error unmarshaling photoNode")
    	}
    	fmt.Println("Photo node retrieved.")

    	/*
    	testPhotoNodeString, _ := json.Marshal(photo)
    	fmt.Println(string(testPhotoNodeString))
    	*/
    	
    	apiUrl := "https://api.havenondemand.com"
	    resource := "/1/api/sync/ocrdocument/v1"
	    data := url.Values{}
	    data.Set("apikey", "6f5569d3-8fc0-4c74-80b4-efe9fd4c90a0")
	    data.Add("url", photo.Source)
	    data.Add("mode", "document_scan")

	    u, _ := url.ParseRequestURI(apiUrl)
	    u.Path = resource
	    urlStr := fmt.Sprintf("%v", u)

	    client := &http.Client{}
	    r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode())) // <-- URL-encoded payload
	    r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	    r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	    fmt.Println("OCR request sent.")

	    havenResp, _ := client.Do(r)
	    fmt.Println(havenResp.Status)
	    ocrOutputRaw, err := ioutil.ReadAll(havenResp.Body)
	    

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
	    ocrErr := json.Unmarshal(ocrOutputRaw, &ocrHavenResp)
	    if (ocrErr != nil) {
	    	fmt.Println("Error unmarshaling OCR request.")
	    }
	    fmt.Println("OCR request received.")
	    
	    testOCRHavenRespString, _ := json.Marshal(ocrHavenResp)
    	fmt.Println(string(testOCRHavenRespString))
    	

    	if (len(ocrHavenResp.TextBlockArray) == 0) {
    		fmt.Println("No text recognized.")
    		continue
    	}

    	//handle case to replace `1\n` but not `2016\n`
    	reNewLine := regexp.MustCompile(`(([^0-9][0-9])?\n)`)
    	ocrHavenResp.TextBlockArray[0].Text = reNewLine.ReplaceAllString(ocrHavenResp.TextBlockArray[0].Text, "")

    	//ocrHavenResp.TextBlockArray[0].Text = strings.Replace(ocrHavenResp.TextBlockArray[0].Text, "\n", "", -1)

    	rePhotoDate := regexp.MustCompile(`(?i:[0-9]{2}[ ]*(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)[ ]*[0-9Il]{2,4})`)
	    fmt.Printf("Photo date %v\n", rePhotoDate.FindString(ocrHavenResp.TextBlockArray[0].Text))

	    //fmt.Println(ocrHavenResp.TextBlockArray[0].Text)

	    rePhotoDateToListing := regexp.MustCompile(`(?im:[0-9]{2}[ ]*(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)[ ]*[0-9]{2,4}.*?)[0-9]{4}`)    
	   	returnStringIndex := rePhotoDateToListing.FindStringIndex(ocrHavenResp.TextBlockArray[0].Text)
	   	if (len(returnStringIndex) > 1) {
	   		departureStartIndex := returnStringIndex[1] - 4
	   		reDeparture := regexp.MustCompile(`(([0-9]{4})[ ]*([A-z ,&;]*)?.*?(-[A-z ,&;]*)*[ ]*(?:([0-9]{1,3}[A-z])|TBD))`)
	   		fmt.Printf("%q\n",reDeparture.FindAllStringSubmatch(ocrHavenResp.TextBlockArray[0].Text[departureStartIndex:], -1))
	   	}
	    //fmt.Println(ocrHavenResp.TextBlockArray[0].Text[departureStartIndex:])

	    

    }
}
