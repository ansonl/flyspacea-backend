package main

import (
	"fmt"
	"time"
	"regexp"
	"strconv"
	"image"
	"strings"
)

//Find date of 72 hour slide in header by looking for month name
func findDateOfPhotoNodeSlides(slides []Slide) (date time.Time, err error) {
	//Build months name array
	var monthsLong []string  //Long month ex:January
	var monthsShort []string //Short month ex:Jan
	for i := time.January; i <= time.December; i++ {
		monthsLong = append(monthsLong, i.String())
		monthsShort = append(monthsShort, i.String()[0:3])
	}

	//Store current and next month values
	var currentMonth time.Month
	var nextMonth time.Month
	currentMonth = time.Now().Month()
	if nextMonth = currentMonth + 1; nextMonth > time.December {
		nextMonth = time.January
	}

	//Search for the current and next month strings
	var closestMonthSpelling string
	var closestMonthSlide Slide
	monthsSearchArray := []string{monthsLong[currentMonth-1], monthsLong[nextMonth-1], monthsShort[currentMonth-1], monthsLong[nextMonth-1]}
	var estimatedMonth time.Month
	for i, v := range monthsSearchArray {
		if closestMonthSpelling, closestMonthSlide, err = findKeywordClosestSpellingInPhotoInSaveImageTypes(v, slides); err != nil {
			return
		}
		if len(closestMonthSpelling) > 0 {
			//We found a close spelling, move onto finding bounding box
			if i % 2 == 0 {
				estimatedMonth = currentMonth
			} else {
				estimatedMonth = nextMonth
			}
			break
		}
	}
	
	//Find month bounds in hOCR
	var bbox image.Rectangle
	bbox, err = getTextBounds(closestMonthSlide.HOCRText, closestMonthSpelling)
	if err != nil {
		return
	}

	fmt.Printf("%v date %v bbox %v %v %v %v\n", closestMonthSlide.Terminal.Title, closestMonthSpelling, bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y)

	//Crop original and processed images to only show the image line that month was found on
	for _, s := range slides {
		if err = runImageMagickDateCropProcess(s, bbox.Min.Y-5, (bbox.Max.Y-bbox.Min.Y)+10); err != nil {
			return
		}
	}

	//Run OCR on cropped image
	closestMonthSlide.Suffix = IMAGE_SUFFIX_CROPPED;
	if err = doOCRForSlide(&closestMonthSlide); err != nil {
		return
	}

	//Find date
	var DMYRegex *regexp.Regexp;
	var MDYRegex *regexp.Regexp;
	//Match Date Month Year. Capture date and year
	if DMYRegex, err = regexp.Compile(fmt.Sprintf("([0-9]{1,2})[ ]*[a-z]{0,3}[ ,]*%v[ ]*([0-9]{2,4})", closestMonthSpelling)); err != nil {
		return
	}
	//Match Month Date Year. Capture date and year
	if MDYRegex, err = regexp.Compile(fmt.Sprintf("%v[ ]*([0-9]{2})[ ]*[a-z]{0,3}[ ,]*([0-9]{4})*", closestMonthSpelling)); err != nil {
		return
	}

	var regexResult []string
	if regexResult = DMYRegex.FindStringSubmatch(strings.ToLower(closestMonthSlide.PlainText)); len(regexResult) == 3 {
	} else if regexResult = MDYRegex.FindStringSubmatch(closestMonthSlide.PlainText); len(regexResult) == 3 {	
	} else {
		//No match
		fmt.Println("no regex match")
		fmt.Println(closestMonthSlide.PlainText)
		return
	}

	var capturedYear int
	var capturedDay int
	if capturedYear, err = strconv.Atoi(regexResult[2]); err != nil {
		return
	}
	if capturedDay, err = strconv.Atoi(regexResult[1]); err != nil {
		return
	}

	date = time.Date(capturedYear, estimatedMonth, capturedDay, 0, 0, 0, 0, time.UTC)
	
	return
}
