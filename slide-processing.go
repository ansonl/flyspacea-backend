package main

import (
	"errors"
	"fmt"
	"image"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//Find date of 72 hour slide in header by looking for month name
func findDateOfPhotoNodeSlides(slides []Slide) (slideDate time.Time, err error) {

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

	//Look through all slides (savetypes) to find closest date to time.Now. Assume any OCR error dates will be at higher date difference than a correct OCR date within ~72 hours.
	compareTargetDate := time.Now()
	for i, v := range monthsSearchArray {
		if closestMonthSpelling, closestMonthSlide, err = findKeywordClosestSpellingInPhotoInSaveImageTypes(v, slides); err != nil {
			return
		}

		//Determine which month we are trying to find in this loop
		var estimatedMonth time.Month
		if i%2 == 0 {
			estimatedMonth = currentMonth
		} else {
			estimatedMonth = nextMonth
		}

		if len(closestMonthSpelling) > 0 {
			//We found a close spelling, move onto finding bounding box

			//Find month bounds in hOCR
			var bboxes []image.Rectangle
			bboxes, err = getTextBounds(closestMonthSlide.HOCRText, closestMonthSpelling)
			if err != nil {
				return
			}

			if len(bboxes) == 0 {
				err = fmt.Errorf("No bboxes found for month??.")
				return
			}
			bbox := bboxes[0]

			//Search for date in all slides using closest match spelling
			for _, s := range slides {
				var foundDate time.Time
				//Try to find date from uncropped image
				if foundDate, err = findDateFromPlainText(s.PlainText, closestMonthSpelling, estimatedMonth); err != nil {
					return
				}

				//Check if date is closest date found so far
				if slideDate.Equal(time.Time{}) || !foundDate.Equal(time.Time{}) {
					slideDate = foundDate
				} else {
					slideDate = closerDate(compareTargetDate, slideDate, foundDate)
				}

				if (slideDate.Equal(time.Time{}) || math.Abs(float64(compareTargetDate.Sub(foundDate))) < math.Abs(float64(compareTargetDate.Sub(slideDate)))) {
					slideDate = foundDate
				}

				//Try to find date on cropped image
				//Crop current slide to only show the image line that month was found on
				if err = runImageMagickDateCropProcess(s, bbox.Min.Y-5, (bbox.Max.Y-bbox.Min.Y)+10); err != nil {
					return
				}
				//Run OCR on cropped image for current slide using bbox
				copySlide := s
				copySlide.Suffix = IMAGE_SUFFIX_CROPPED
				if err = doOCRForSlide(&copySlide); err != nil {
					return
				}

				//Try to find date from cropped image
				if foundDate, err = findDateFromPlainText(copySlide.PlainText, closestMonthSpelling, estimatedMonth); err != nil {
					return
				}

				//Check if date is closest date found so far
				if slideDate.Equal(time.Time{}) || !foundDate.Equal(time.Time{}) {
					slideDate = foundDate
				} else {
					slideDate = closerDate(compareTargetDate, slideDate, foundDate)
				}
			}
		} else {
			//fuzzy match not found
		}
		//Try to find exact match with regex. Fuzzy match may fail or get bad results if month is between other letters.
		for _, s := range slides {
			var foundDate time.Time
			//Try to find date from uncropped image
			if foundDate, err = findDateFromPlainText(s.PlainText, v, estimatedMonth); err != nil {
				return
			}

			//Check if date is closest date found so far
			if slideDate.Equal(time.Time{}) || !foundDate.Equal(time.Time{}) {
				slideDate = foundDate
			} else {
				slideDate = closerDate(compareTargetDate, slideDate, foundDate)
			}

			if (slideDate.Equal(time.Time{}) || math.Abs(float64(compareTargetDate.Sub(foundDate))) < math.Abs(float64(compareTargetDate.Sub(slideDate)))) {
				slideDate = foundDate
			}
		}
	}

	//fmt.Printf("%v date %v bbox %v %v %v %v\n", closestMonthSlide.Terminal.Title, closestMonthSpelling, bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y)

	return
}

func findDateFromPlainText(plainText string, closestMonthSpelling string, estimatedMonth time.Month) (date time.Time, err error) {
	//Lowercase closestMonthSpelling
	closestMonthSpelling = strings.ToLower(closestMonthSpelling)

	//fmt.Println("find date with closestMonthSpelling ", closestMonthSpelling)
	//fmt.Println("plaintext", plainText)

	//Find date with Regexp
	var DMYRegex *regexp.Regexp
	var MDYRegex *regexp.Regexp
	//Match Date Month Year. Capture date and year
	if DMYRegex, err = regexp.Compile(fmt.Sprintf("([0-9]{1,2})[a-z]{0,3}%v([0-9]{2,4})", closestMonthSpelling)); err != nil {
		return
	}
	//Match Month Date Year. Capture date and year
	if MDYRegex, err = regexp.Compile(fmt.Sprintf("%v([0-9]{2})[a-z]{0,3}([0-9]{4})", closestMonthSpelling)); err != nil {
		return
	}

	//Remove common OCR errors from and lowercase input string
	r := strings.NewReplacer(
		".", "",
		",", "",
		" ", "")
	var input = strings.ToLower(r.Replace(plainText))
	var regexResult []string
	if regexResult = DMYRegex.FindStringSubmatch(input); len(regexResult) == 3 {
		/*
			fmt.Println(regexResult, len(regexResult))
			for i, r := range regexResult {
				fmt.Println(i, r)
			}
		*/
	} else if regexResult = MDYRegex.FindStringSubmatch(input); len(regexResult) == 3 {
		/*
			fmt.Println(regexResult, len(regexResult))
			for i, r := range regexResult {
				fmt.Println(i, r)
			}
		*/
	} else {
		//No match, proceed to next processed slide
		//fmt.Println("no regex match")
		//fmt.Println(input)
		noMatchDateHeaderInputs = append(noMatchDateHeaderInputs, input)
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

	//If found date is closer to time.Now than other dates, keep it and check other processed slides for other date matches
	date = time.Date(capturedYear, estimatedMonth, capturedDay, 0, 0, 0, 0, time.UTC)
	return
}

func findDestinationLabelBoundsOfPhotoNodeSlides(slides []Slide) (bbox image.Rectangle, err error) {
	//Find closest spelling for KEYWORD_DESTINATION
	var closestDestinationSpelling string
	var closestDestinationSlide Slide
	if closestDestinationSpelling, closestDestinationSlide, err = findKeywordClosestSpellingInPhotoInSaveImageTypes(KEYWORD_DESTINATION, slides); err != nil {
		return
	}

	/*
		if len(closestDestinationSpelling) == 0 {
			displayMessageForTerminal(closestDestinationSlide.Terminal, fmt.Sprintf("No close dest spelling founds"));
		} else {
			displayMessageForTerminal(closestDestinationSlide.Terminal, fmt.Sprintf("Closest dest spelling %v in saveType %v", closestDestinationSpelling, closestDestinationSlide.SaveType));
		}
	*/

	//Find KEYWORD_DESTINATION bounds in hOCR
	bboxes, err := getTextBounds(closestDestinationSlide.HOCRText, closestDestinationSpelling)
	if err != nil {
		return
	}

	if len(bboxes) == 0 {
		err = errors.New("No bbox found for findDestinationLabelBoundsOfPhotoNodeSlides")
	}

	bbox = bboxes[0]

	return
}

func findDestinationsFromSlides(slides []Slide) (destinations []Destination, err error) {

	var foundDestinations []Destination
	//var foundRollCall []RollCall

	for _, s := range slides {
		var found map[string]TerminalKeywordsResult
		found = make(map[string]TerminalKeywordsResult)
		found = findTerminalKeywordsInPlainText(s.PlainText)

		for spelling, result := range found {
			var bboxes []image.Rectangle
			if bboxes, err = getTextBounds(s.HOCRText, spelling); err != nil {
				return
			}

			for _, bbox := range bboxes {
				foundDestinations = append(foundDestinations, Destination{
					TerminalTitle:    locationKeywordMap[result.Keyword],
					Spelling:         spelling,
					SpellingDistance: result.Distance,
					BBox:             bbox})

				//Find duplicates by checking if intersecting rect shares >50% of area of the smaller of the two rects
				intersectThreshold := 0.5
				for i := 0; i < len(foundDestinations); i++ {
					destA := foundDestinations[i]
					smallerArea := destA.BBox.Dx() * destA.BBox.Dy()
					for j := i + 1; j < len(foundDestinations); j++ {
						destB := foundDestinations[j]
						destBArea := destA.BBox.Dx() * destA.BBox.Dy()
						if destBArea < smallerArea {
							smallerArea = destBArea
						}

						//Compare intersection image.Rectangle area to the smaller of destA and destB area
						intersection := destA.BBox.Intersect(destB.BBox)
						if float64(intersection.Dx())*float64(intersection.Dy()) > float64(smallerArea)*intersectThreshold {

							fmt.Println("duplicate found for ", destA.TerminalTitle)

							//If destA spelling distance > destB spelling distance, replace destA location in array with destB.
							if destA.SpellingDistance > destB.SpellingDistance {
								foundDestinations[i] = foundDestinations[j]
							}

							//Delete destB location. Decrement j so that same index now with different element is checked on next loop
							copy(foundDestinations[j:], foundDestinations[j+1:])
							foundDestinations[len(foundDestinations)-1] = Destination{}
							foundDestinations = foundDestinations[:len(foundDestinations)-1]
							j--
						}
					}
				}

				//Create Grouping with nearest DestinationA to DestinationB

				//Link GroupingA with nearest GroupingB so all Destinations in GroupingB are in GroupingA. Repeat until GroupingA contains a Destination that (horizontally) intersects a RollCall.

				//match to time

			}
		}
	}

	destinations = foundDestinations
	return
}
