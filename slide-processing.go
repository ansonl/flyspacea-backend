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
	"log"
	"sort"
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

//Return time.Time of detected date for slide.
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

//Return bounds of KEYWORD_DESTINATION in slide.
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

//Search slides in Slide slice for Destinations.
//limitMinY is minimum Y coordinate needed to RollCall (destination keyword bbox)
//Return a Destination slice with found and deduplicated Destinations.
func findDestinationsFromSlides(slides []Slide, limitMinY int) (foundDestinations []Destination, err error) {
	//Find location keyword spellings in image pointed to by each slide.
	for _, s := range slides {
		var found map[string]TerminalKeywordsResult
		found = make(map[string]TerminalKeywordsResult)
		found = findTerminalKeywordsInPlainText(s.PlainText)

		//Get text bounds from hOCR for each potential spelling found.
		for spelling, result := range found {
			var bboxes []image.Rectangle
			if bboxes, err = getTextBounds(s.HOCRText, spelling); err != nil {
				return
			}

			//Skip result if bounding box too high. MinY too small.
			//Append new Destination to foundDestinations for each bounding box found.
			for _, bbox := range bboxes {
				if bbox.Min.Y < limitMinY {
					continue
				}

				foundDestinations = append(foundDestinations, Destination{
					TerminalTitle:    locationKeywordMap[result.Keyword],
					Spelling:         spelling,
					SpellingDistance: result.Distance,
					SharedInfo:       SharedInfo{BBox: bbox}})
			}
		}
	}

	deleteDuplicatesFromDestinationArray(&foundDestinations)

	return
}

//Search slides in Slide slice for RollCalls.
//estimatedDay is passed in to set day for returned RollCall
//limitMinY is minimum Y coordinate needed to RollCall (destination keyword bbox)
//Return a RollCall slice with found and deduplicated RollCalls.
func findRollCallTimesFromSlides(slides []Slide, estimatedDay time.Time, limitMinY int) (foundRCs []RollCall, err error) {
	for _, s := range slides {
		var found24HR []time.Time
		//fmt.Println("date slide type ", s.SaveType)
		if found24HR, err = find24HRFromPlainText(s.PlainText, estimatedDay); err != nil {
			return
		}

		//Get text bounds from hOCR for each 24HR time text found.
		for _, result := range found24HR {
			var bboxes []image.Rectangle
			if bboxes, err = getTextBounds(s.HOCRText, result.Format("1504")); err != nil {
				return
			}

			//fmt.Println("hocr result for ", result.Format("1504"), bboxes)

			//Add RollCalls to foundRCs slice
			//Skip result if bounding box too high. MinY too small.
			//Append new RollCall to foundRCs for each bounding box found.
			for _, bbox := range bboxes {
				if bbox.Min.Y < limitMinY {
					continue
				}

				foundRCs = append(foundRCs, RollCall{
					Time:       result,
					SharedInfo: SharedInfo{BBox: bbox}})
			}
		}
	}

	deleteDuplicatesFromRCArray(&foundRCs)

	return
}

//Search plain text for 24HR time. Return slice of time.Time of found 24HR time on estimatedDay.
func find24HRFromPlainText(plainText string, estimatedDay time.Time) (found24HR []time.Time, err error) {
	//Find 24hr time with Regexp
	var HR24Regex *regexp.Regexp
	//Match 24 hr time format
	//original https://stackoverflow.com/a/1494700
	if HR24Regex, err = regexp.Compile("\\b(?:[01]\\d|2[0-3])(?:[0-5]\\d)\\b"); err != nil {
		return
	}

	//lowercase input string
	var input = strings.ToLower(plainText)
	var regexResult []string
	if regexResult = HR24Regex.FindAllString(input, -1); regexResult == nil {
		//No match, proceed to next processed slide
		return
	}

	for _, result := range regexResult {
		//fmt.Println("found regex", result)

		var capturedHour int
		var capturedMinute int
		if capturedHour, err = strconv.Atoi(result[:2]); err != nil {
			return
		}
		if capturedMinute, err = strconv.Atoi(result[2:]); err != nil {
			return
		}

		found24HR = append(found24HR, time.Date(
			estimatedDay.Year(),
			estimatedDay.Month(),
			estimatedDay.Day(),
			capturedHour,
			capturedMinute,
			0, 0,
			time.UTC))
	}
	return
}

//For each RollCall - link vertically closest Destination. Create 'anchors' for grouping. Best effort match no threshold.
//Runtime: O(n*mlogm + n*m) n=len(rcs) m=len(destsArray)
func linkRollCallsToNearestDestinations(rcs []RollCall, destsArray []Destination) {
	getVerticalDistance := func(bbox1 image.Rectangle, bbox2 image.Rectangle) (verticalDistance int) {

		/*
		 *			***
		 *			*2*
		 *			***
		 *		 |
		 * 	***
		 *	*1*
		 *	***
		 */

		if bbox1.Min.Y >= bbox2.Max.Y {
			verticalDistance = bbox1.Min.Y - bbox2.Max.Y
			return
		}

		/*
		 *	***
		 *	*1*
		 *	***
		 *		 |
		 * 			***
		 *			*2*
		 *			***
		 */

		if bbox1.Max.Y <= bbox2.Min.Y {
			verticalDistance = bbox2.Min.Y - bbox1.Max.Y
			return
		}

		//Boxes overlap. Return negative distance.
		if bbox1.Min.Y >= bbox2.Min.Y {

			/*
			 *			***
			 *			*2*
			 *	***` |	* *
			 *	*1*	 |	* *
			 *	***	 |	* *
			 * 			***
			 *
			 */

			/*
			 *			***
			 *			*2*
			 *	***` |	* *
			 *	*1*	 |	***
			 *	* *
			 * 	***
			 *
			 */

			if bbox1.Max.Y <= bbox2.Max.Y {
				verticalDistance = -1 * bbox1.Dy()
				return
			} else {
				verticalDistance = bbox1.Min.Y - bbox2.Max.Y
				return
			}
		} else {

			/*
			 *	***
			 *	*1*
			 *	* *` |	***
			 *	* *	 |	*2*
			 *	* *	 |	***
			 * 	***
			 *
			 */

			/*
			 *	***
			 *	*1*
			 *	* *` |	***
			 *	***	 |	*2*
			 *		 	* *
			 * 			***
			 *
			 */

			if bbox1.Max.Y >= bbox2.Max.Y {
				verticalDistance = -1 * bbox2.Dy()
				return
			} else {
				verticalDistance = bbox2.Min.Y - bbox1.Max.Y
				return
			}
		}
		log.Fatal("No category matched ", bbox1, bbox2)
		return
	}

	type DestDist struct {
		Destination *Destination
		Distance int
	}

	var distances map[*RollCall][]DestDist
	distances = make(map[*RollCall][]DestDist)

	//Compute distances for each RollCall -> Destination.
	//Runtime: O(n*mlogm) n=len(rcs) m=len(destsArray)
	for rcIndex, _ := range rcs {
		for dIndex, _ := range destsArray {
			vertDist := getVerticalDistance(rcs[rcIndex].BBox, destsArray[dIndex].BBox)

			//If vertical distance > ROLLCALLS_DESTINATION_LINK_VERTICAL_THRESHOLD, don't add Dest to distance array. 
			//This is to prevent false 'anchor' when a destination is not found in OCR/fuzzy search and a RollCall is left without the correct 'anchor' Destination because that Destination was not found. Otherwise the RollCall will be matched with a Destination that should be grouped with another Destination (multiple Destination to 1 RollCall)
			if vertDist > ROLLCALLS_DESTINATION_LINK_VERTICAL_THRESHOLD {
				continue
			}

			distances[&rcs[rcIndex]] = append(distances[&rcs[rcIndex]], DestDist {
				Destination: &destsArray[dIndex],
				Distance: vertDist})
		}

		//Sort distances slice
		sort.Slice(distances[&rcs[rcIndex]], func(i, j int) bool {
			return distances[&rcs[rcIndex]][i].Distance < distances[&rcs[rcIndex]][j].Distance
		});
	}

	//Find nearest Destination for RollCall
	var findNearestDestForRollCall func(rcP *RollCall)
	findNearestDestForRollCall = func(rcP *RollCall) {
		//If RollCall has no more distances left in map, quit
		for len(distances[rcP]) > 0 {

			//Look at closest Destination
			currentDestP := distances[rcP][0].Destination
			linkedRollCallP := (*distances[rcP][0].Destination).LinkedRollCall
			if linkedRollCallP != nil {
				
				//If current RC is closer to the lowest Destination in distances[rc] than the linked RC
				if getVerticalDistance((*rcP).BBox, (*currentDestP).BBox) < getVerticalDistance((*linkedRollCallP).BBox, (*currentDestP).BBox) {

					//Remove Destination from linked RollCall's distance slice (should be first in the slice since it was closest for that linked RollCall)
					distances[linkedRollCallP] = distances[linkedRollCallP][1:]

					//Set current RC as linked RollCall
					(*currentDestP).LinkedRollCall = rcP

					//Look for closest Dest for last linked RollCall. Break out of loop. Any further work for this RollCall will occur when contested by other RollCall.
					findNearestDestForRollCall(linkedRollCallP)
					break;
				} else {
					//look at next closest dest
					distances[rcP] = distances[rcP][1:]
					//continue //same behavior due to this being last operation in loop
				}
			} else { //No existing linked RollCall
				//Set current RC as linked RollCall
				(*currentDestP).LinkedRollCall = rcP
				break;
			}
		}
	}

	//Runtime: O(n*m) n=len(rcs) m=len(destsArray)
	for rcIndex, _ := range rcs {
		findNearestDestForRollCall(&rcs[rcIndex])
	}
	
/*
	for rcIndex, rc := range rcs {
		var nearestDestIndex int
		var nearestVertDist int
		nearestDestIndex = -1

		for dIndex, d := range destsArray {
			vertDist := getVerticalDistance(rc.BBox, d.BBox)
			if nearestDestIndex == -1 || vertDist < nearestVertDist {
				nearestDestIndex = dIndex
				nearestVertDist = vertDist
			}
		}

		destsArray[nearestDestIndex].LinkedRollCall = &rcs[rcIndex]
	}
	*/
}
