package main

import (
	"errors"
	"fmt"
	"image"
	//"log"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	//"unicode"
)

//Find date of 72 hour slide in header by looking for month name
//Returned time.Time is set to TZ of slides[0]
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
	//Get current month of image creation date on Facebook instead of current month of time.Now()
	//currentMonth = time.Now().Month()
	currentMonth = slides[0].FBCreatedTime.Month()
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
				if foundDate, err = findDateFromPlainText(s.PlainText, closestMonthSpelling, estimatedMonth, s.Terminal.Timezone); err != nil {
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
				if err = runImageMagickDateCropVerticalProcess(s, bbox.Min.Y-5, (bbox.Max.Y-bbox.Min.Y)+10); err != nil {
					return
				}
				//Run OCR on cropped image for current slide using bbox
				copySlide := s
				copySlide.Suffix = IMAGE_SUFFIX_CROPPED
				if err = doOCRForSlide(&copySlide, OCR_WHITELIST_NORMAL); err != nil {
					return
				}

				//Try to find date from cropped image
				if foundDate, err = findDateFromPlainText(copySlide.PlainText, closestMonthSpelling, estimatedMonth, s.Terminal.Timezone); err != nil {
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
			if foundDate, err = findDateFromPlainText(s.PlainText, v, estimatedMonth, s.Terminal.Timezone); err != nil {
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

	//If found date is > 144hr away from current time, assume we got the wrong date since we are only looking at recent photos
	if math.Abs(float64(time.Since(slideDate))) > float64(time.Hour*144) {
		slideDate = time.Time{}
	}

	//fmt.Printf("%v date %v bbox %v %v %v %v\n", closestMonthSlide.Terminal.Title, closestMonthSpelling, bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y)

	return
}

//Return time.Time of detected date for slide.
func findDateFromPlainText(plainText string, closestMonthSpelling string, estimatedMonth time.Month, slideTZ *time.Location) (date time.Time, err error) {
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
	date = time.Date(capturedYear, estimatedMonth, capturedDay, 0, 0, 0, 0, slideTZ)
	return
}

//Return bounds of KEYWORD_XXX in slide.
func findLabelBoundsOfPhotoNodeSlides(slides []Slide, label string) (bbox image.Rectangle, err error) {
	//Find closest spelling for label
	var closestDestinationSpelling string
	var closestDestinationSlide Slide
	if closestDestinationSpelling, closestDestinationSlide, err = findKeywordClosestSpellingInPhotoInSaveImageTypes(label, slides); err != nil {
		return
	}

	//fmt.Println("closest spelling ", closestDestinationSpelling)

	//Find KEYWORD_DESTINATION bounds in hOCR
	bboxes, err := getTextBounds(closestDestinationSlide.HOCRText, closestDestinationSpelling)
	if err != nil {
		return
	}

	if len(bboxes) == 0 {
		err = errors.New("No bbox found for findLabelBoundsOfPhotoNodeSlides")
		return
	} else if len(bboxes) > 1 {
		fmt.Println("multiple bbox", bboxes)
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
		var found map[string]TerminalKeywordsResult //map[spelling]{Title, Distance}
		found = make(map[string]TerminalKeywordsResult)
		found = findTerminalKeywordsInPlainText(s.PlainText)

		//fmt.Println("found keywords", found)

		//Get text bounds from hOCR for each potential spelling found.
		for spelling, result := range found {
			var bboxes []image.Rectangle
			if bboxes, err = getTextBounds(s.HOCRText, spelling); err != nil {
				return
			}

			//fmt.Println("found bbox for ", spelling, bboxes, "\nmin y ", limitMinY)

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
func findRollCallTimesFromSlides(slides []Slide, estimatedDay time.Time, limitMinY int) (foundRCs []RollCall, foundNoBBoxRCs []RollCall, err error) {
	for _, s := range slides {
		var found24HR []time.Time
		//fmt.Println("date slide type ", s.SaveType)
		if found24HR, err = find24HRFromPlainText(s.PlainText, estimatedDay, s.Terminal.Timezone); err != nil {
			return
		}

		//Get text bounds from hOCR for each 24HR time text found.
		for _, result := range found24HR {
			var bboxes []image.Rectangle
			if bboxes, err = getTextBounds(s.HOCRText, result.Format("1504")); err != nil {
				return
			}

			//fmt.Println("hocr result for ", result.Format("1504"), bboxes)

			//Not found in hOCR. Probably because time had spaces in it. Keep time as possible time to display to user?
			if len(bboxes) == 0 {
				foundNoBBoxRCs = append(foundNoBBoxRCs, RollCall{
					Time: result})
			}

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
func find24HRFromPlainText(plainText string, estimatedDay time.Time, slideTZ *time.Location) (found24HR []time.Time, err error) {
	//lowercase input string
	var input = strings.ToLower(plainText)
	//fmt.Println(input)

	/*
	//Find whitespace between digits with Regexp\
	//These will not be found in hOCR
	var whitespaceBetweenDigitsRegex *regexp.Regexp
	if whitespaceBetweenDigitsRegex, err = regexp.Compile("\\d( )\\d(?:( )\\d)?(?:( )\\d)?"); err != nil {
		return
	}
	input = whitespaceBetweenDigitsRegex.ReplaceAllStringFunc(input, func(match string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, match)
	})
	*/

	//Find 24hr time with Regexp
	var HR24Regex *regexp.Regexp
	//Match 24 hr time format
	//original https://stackoverflow.com/a/1494700
	if HR24Regex, err = regexp.Compile("\\b(?:[01]\\d|2[0-3])(?:[0-5]\\d)\\b"); err != nil {
		return
	}

	var regexResult []string
	if regexResult = HR24Regex.FindAllString(input, -1); regexResult == nil {
		//No match, proceed to next processed slide
		//fmt.Println("no RC regex result")
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
			slideTZ))
	}
	return
}

//Search slides in Slide slice for SeatsAvailable text.
//seatsLabelBBox is KEYWORD_SEATS bbox for cropping
//Return a SeatsAvailable slice with found and deduplicated SeatsAvailable.
func findSeatsAvailableFromSlides(slides []Slide, seatsLabelBBox image.Rectangle) (foundSAs []SeatsAvailable, err error) {
	for _, s := range slides {

		//Try to find seats on cropped image
		//Crop current slide to only show the image column downwards from where seats label was found on
		if err = runImageMagickDateCropHorizontalProcess(s, image.Point{X: seatsLabelBBox.Min.X - SEATS_CROP_HORIZONTAL_BUFFER, Y: seatsLabelBBox.Max.X + SEATS_CROP_HORIZONTAL_BUFFER}, seatsLabelBBox.Max.Y); err != nil {
			return
		}

		//Run OCR on cropped image for current slide
		cropSlide := s
		cropSlide.Suffix = IMAGE_SUFFIX_CROPPED
		if err = doOCRForSlide(&cropSlide, OCR_WHITELIST_SA); err != nil {
			return
		}

		//Regex search for seats counts
		if foundSAs, err = findSeatsFromPlainText(cropSlide.PlainText); err != nil {
			return
		}

		//fmt.Println("look SA in slide", s.SaveType, cropSlide.HOCRText)

		//Get text bounds from hOCR for each seat text found.
		for _, result := range foundSAs {
			var bboxes []image.Rectangle

			var searchString string
			if result.Number == 0 {
				searchString = fmt.Sprintf("%v", result.Letter)
			} else {
				searchString = fmt.Sprintf("%v%v", strconv.Itoa(result.Number), result.Letter)
			}

			//Find text bounds of found SA text
			if bboxes, err = getTextBounds(cropSlide.HOCRText, searchString); err != nil {
				return
			}

			for _, bbox := range bboxes {
				newSA := result
				bbox.Min.Y += seatsLabelBBox.Max.Y
				bbox.Max.Y += seatsLabelBBox.Max.Y
				newSA.BBox = bbox

				foundSAs = append(foundSAs, newSA)
			}
		}
	}

	deleteDuplicatesFromSAArray(&foundSAs)
	//fmt.Println(foundSAs)

	return
}

//Search plain text for seat count text. Return slice of found SeatsAvailable without BBox.
func findSeatsFromPlainText(plainText string) (foundSAs []SeatsAvailable, err error) {

	var SeatsCountRegex *regexp.Regexp
	if SeatsCountRegex, err = regexp.Compile("\\b(?:(?:(\\d{1,3})(f|t|sp)?)|(sp|tbd))\\b"); err != nil {
		return
	}

	//lowercase input string
	var input = strings.ToLower(plainText)
	//fmt.Println(input)
	var regexResult [][]string
	if regexResult = SeatsCountRegex.FindAllStringSubmatch(input, -1); regexResult == nil {
		//No match, proceed to next processed slide
		return
	}

	//fmt.Println("SA regex results len ", len(regexResult))

	for _, result := range regexResult {
			
			/*
			fmt.Println("found SA regex len ", len(result))
			for n, r := range result {
				fmt.Println(n, len(r), r)
			}
			*/
			

		var capturedSeatCount int
		var capturedSeatLetter string //F/T/SP|TBD->SP

		//Check appropriate result indices for seat info
		if len(result[1]) > 0 { //Check if number captured
			if capturedSeatCount, err = strconv.Atoi(result[1]); err != nil {
				return
			}

			if len(result[2]) > 0 { //Check for letter code
				capturedSeatLetter = result[2]
			}
		} else if len(result[3]) > 0 { //No number, check for letter code in third capture group
			capturedSeatLetter = result[3]
		}

		sa := SeatsAvailable{
			Number: capturedSeatCount,
			Letter: capturedSeatLetter}

		foundSAs = append(foundSAs, sa)
	}
	return
}

//For each RollCall - link vertically closest Destination. Create 'anchors' for grouping. Best effort match no threshold.
//Runtime: O(n*mlogm + n*m) n=len(rcs) m=len(destsArray)
func linkRollCallsToNearestDestinations(rcs []RollCall, destsArray []Destination) {

	//Struct for holding Destination and (vertical) distance from target for comparison
	type DestDist struct {
		Destination *Destination
		Distance    int
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

			distances[&rcs[rcIndex]] = append(distances[&rcs[rcIndex]], DestDist{
				Destination: &destsArray[dIndex],
				Distance:    vertDist})
		}

		//Sort distances slice
		sort.Slice(distances[&rcs[rcIndex]], func(i, j int) bool {
			return distances[&rcs[rcIndex]][i].Distance < distances[&rcs[rcIndex]][j].Distance
		})
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
					break
				} else {
					//look at next closest dest
					distances[rcP] = distances[rcP][1:]
					//continue //same behavior due to this being last operation in loop
				}
			} else { //No existing linked RollCall
				//Set current RC as linked RollCall
				(*currentDestP).LinkedRollCall = rcP
				break
			}
		}
	}

	//Runtime: O(n*m) n=len(rcs) m=len(destsArray)
	for rcIndex, _ := range rcs {
		findNearestDestForRollCall(&rcs[rcIndex])
	}

	/*
		//Naive implementation, incorrect for conflicts
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

//Group Destinations into Grouping{}. Pass Destinations into function as array of individual Grouping{}.
//Runtime: O(n*n)
func combineDestinationGroupsToAnchorDestinations(groupsP *[]Grouping) {
	groups := *groupsP

	for growIndex := 0; growIndex < len(groups); growIndex++ {

		/*
			//fmt.Println("groups")
			for _, g := range groups {
				fmt.Println(g)
			}
		*/

		//fmt.Println("grow group ", growIndex, groups[growIndex])

		var deletedPreviousGroupCount int

		//Keep growing group to include nearest groups while it is not including an anchor. Only grow groups that do not contain anchor.
		for groups[growIndex].LinkedRollCall == nil {
			//Look for nearest group
			var nearestGroupP *Grouping
			var nearestGroupIndex int //Needed to know where to delete
			var nearestGroupDistance int
			for j := 0; j < len(groups); j++ {
				//If same group as growing group, skip. Don't combine group with itself.
				if &groups[growIndex] == &groups[j] {
					continue
				}

				//Check if group is the nearest group found so far
				groupDist := getVerticalDistance(groups[growIndex].BBox, groups[j].BBox)

				//fmt.Println("check group ", groups[j], "dist", groupDist, "nearestDist", nearestGroupDistance)

				if nearestGroupP == nil || groupDist < nearestGroupDistance {
					nearestGroupP = &groups[j]
					nearestGroupIndex = j
					nearestGroupDistance = groupDist
				}
			}

			//Only one group was passed into function
			if nearestGroupP == nil {
				break
			}

			//Copy nearest group Destinations to grow group Destinations
			groups[growIndex].Destinations = append(groups[growIndex].Destinations, (*nearestGroupP).Destinations...)
			//Copy RollCall to indicate anchor (if applicable)
			groups[growIndex].LinkedRollCall = (*nearestGroupP).LinkedRollCall
			//Update grow group BBox
			groups[growIndex].updateBBox()

			//Delete nearest group
			copy(groups[nearestGroupIndex:], groups[nearestGroupIndex+1:])
			groups[len(groups)-1] = Grouping{}
			groups = groups[:len(groups)-1]

			//Decrement for loop counter because we removed an element.
			//If the removed group was at a lower index than current grow group, decrement loop counter by 1 so that next loop will start at same index that will point to the previously "next" element.
			//If removed group at higher index, don't decrement because for loop counter will get the next element no matter what.
			if nearestGroupIndex <= growIndex {
				deletedPreviousGroupCount++
			}

			growIndex -= deletedPreviousGroupCount

			//fmt.Println("grow index ", growIndex, "len groups ", len(groups), "delete count", deletedPreviousGroupCount)
		}

	}

	//Create new slice to copy elements over. Original slice will have updated length but old elements in memory (displayed when printing).
	tmp := make([]Grouping, len(groups))
	for i := 0; i < len(groups); i++ {
		tmp[i] = groups[i]
	}
	*groupsP = tmp
}

//For each RollCall - link to SeatsAvailable sharing > threshold vertical pixels. Similar to linkRollCallsToNearestDestinations but reduced for simplicity. There should be one to one relationship for RollCall and SeatsAvailable
func linkRollCallsToNearestSeatsAvailable(rcs []RollCall, saArray []SeatsAvailable) {
	/*
		for _, r := range rcs {
			fmt.Println(r)
		}
		for _, s := range saArray {
			fmt.Println(s)
		}
	*/

	//Link each SeatsAvailable. RollCall -> SeatsAvailable.
	//Runtime: O(n*m)
	for saIndex, _ := range saArray {
		for rcIndex, _ := range rcs {
			vertDist := getVerticalDistance(rcs[rcIndex].BBox, saArray[saIndex].BBox)

			//If intersecting vertically enough, link
			if vertDist < ROLLCALLS_SEATS_LINK_VERTICAL_THRESHOLD {
				rcs[rcIndex].LinkedSeatsAvailable = &saArray[saIndex]
				break
			}
		}
	}
}

//For each Destination - link to SeatsAvailable sharing > threshold vertical pixels. Similar to linkRollCallsToNearestDestinations but reduced for simplicity. There MAY be a one to one relationship for RollCall and SeatsAvailable. Not guaranteed since there may be only one seat label for multiple destinations (ex: in a grouping).
func linkDestinationsToNearestSeatsAvailable(dests []Destination, saArray []SeatsAvailable) {
	/*
		for _, r := range dests {
			fmt.Println(r)
		}
		for _, s := range saArray {
			fmt.Println(s)
		}
	*/

	//Link each SeatsAvailable. RollCall -> SeatsAvailable.
	//Runtime: O(n*m)
	for saIndex, _ := range saArray {
		for dIndex, _ := range dests {
			vertDist := getVerticalDistance(dests[dIndex].BBox, saArray[saIndex].BBox)

			//If intersecting vertically enough, link
			if vertDist < ROLLCALLS_SEATS_LINK_VERTICAL_THRESHOLD {
				dests[dIndex].LinkedSeatsAvailable = &saArray[saIndex]
				break
			}
		}
	}
}
