//Methods for consolidating (deduplicating) information found in Slides
package main

//Find duplicates by checking if intersecting rect shares >DUPLICATE_AREA_THRESHOLD of area of the smaller of the two rects.
//Passed in pointer to Destination slice is reassigned to new slice.
//Run time O(n^2). n = number of Destinations
func deleteDuplicatesFromDestinationArray(destsArrayPointer *[]Destination) {
	dests := *destsArrayPointer
	originalLength := len(dests)

	for i := 0; i < len(dests); i++ {
		destA := dests[i]
		smallerArea := destA.BBox.Dx() * destA.BBox.Dy()
		for j := i + 1; j < len(dests); j++ {
			destB := dests[j]
			destBArea := destA.BBox.Dx() * destA.BBox.Dy()
			if destBArea < smallerArea {
				smallerArea = destBArea
			}

			//Compare intersection image.Rectangle area to the smaller of destA and destB area
			//OR if same destination terminal (predicted by fuzzy match) and on 50% same horizontal line due to fuzzy match individual words from same location
			var horizontalDuplicate bool
			if destA.TerminalTitle == destB.TerminalTitle {
				horizontalDuplicate = sameHorizontalLine(destA.BBox, destB.BBox)
			}

			intersection := destA.BBox.Intersect(destB.BBox)
			if float64(intersection.Dx())*float64(intersection.Dy()) > float64(smallerArea)*DUPLICATE_AREA_THRESHOLD || horizontalDuplicate {

				//If destA spelling distance > destB spelling distance, replace destA location in array with destB.
				if destA.SpellingDistance > destB.SpellingDistance {
					dests[i] = dests[j]
				}

				//Delete destB location. Decrement j so that same index now with different element is checked on next loop
				copy(dests[j:], dests[j+1:])
				dests[len(dests)-1] = Destination{}
				dests = dests[:len(dests)-1]
				j--
			}
		}
	}

	//If duplicates were found, alloc new Destination slice and reassign passed in slice pointer
	if len(dests) != originalLength {
		//Create new slice to copy elements over. Original slice will have updated length but old elements in memory (displayed when printing).
		tmp := make([]Destination, len(dests))
		for i := 0; i < len(dests); i++ {
			tmp[i] = dests[i]
		}
		*destsArrayPointer = tmp
	}
}

//Find duplicates by checking if intersecting rect shares >DUPLICATE_AREA_THRESHOLD of area of the smaller of the two rects.
//Passed in pointer to RollCall slice is reassigned to new slice.
//Same function as deleteDuplicatesFromDestinationArray
//Run time O(n^2). n = number of RollCalls
func deleteDuplicatesFromRCArray(arrayPointer *[]RollCall) {
	dests := *arrayPointer
	originalLength := len(dests)

	for i := 0; i < len(dests); i++ {
		destA := dests[i]
		smallerArea := destA.BBox.Dx() * destA.BBox.Dy()
		for j := i + 1; j < len(dests); j++ {
			destB := dests[j]
			destBArea := destA.BBox.Dx() * destA.BBox.Dy()
			if destBArea < smallerArea {
				smallerArea = destBArea
			}

			//Compare intersection image.Rectangle area to the smaller of destA and destB area
			intersection := destA.BBox.Intersect(destB.BBox)
			if float64(intersection.Dx())*float64(intersection.Dy()) > float64(smallerArea)*DUPLICATE_AREA_THRESHOLD {

				//Delete destB location. Decrement j so that same index now with different element is checked on next loop
				copy(dests[j:], dests[j+1:])
				dests[len(dests)-1] = RollCall{}
				dests = dests[:len(dests)-1]
				j--
			}
		}
	}

	//If duplicates were found, alloc new Destination slice and reassign passed in slice pointer
	if len(dests) != originalLength {
		//Create new slice to copy elements over. Original slice will have updated length but old elements in memory (displayed when printing).
		tmp := make([]RollCall, len(dests))
		for i := 0; i < len(dests); i++ {
			tmp[i] = dests[i]
		}
		*arrayPointer = tmp
	}
}

//Find duplicates by checking if intersecting rect shares >DUPLICATE_AREA_THRESHOLD of area of the smaller of the two rects.
//Passed in pointer to SeatsAvailable slice is reassigned to new slice.
//Same function as deleteDuplicatesFromDestinationArray
//Run time O(n^2). n = number of SeatsAvailable
func deleteDuplicatesFromSAArray(arrayPointer *[]SeatsAvailable) {
	dests := *arrayPointer
	originalLength := len(dests)

	for i := 0; i < len(dests); i++ {
		destA := dests[i]
		smallerArea := destA.BBox.Dx() * destA.BBox.Dy()
		for j := i + 1; j < len(dests); j++ {
			destB := dests[j]
			destBArea := destA.BBox.Dx() * destA.BBox.Dy()
			if destBArea < smallerArea {
				smallerArea = destBArea
			}

			//Compare intersection image.Rectangle area to the smaller of destA and destB area
			intersection := destA.BBox.Intersect(destB.BBox)
			if float64(intersection.Dx())*float64(intersection.Dy()) > float64(smallerArea)*DUPLICATE_AREA_THRESHOLD {

				//If destA has no number found and destB found a number replace
				if destA.Number == 0 && destB.Number != 0 || destB.Number > destA.Number {
					dests[i] = dests[j]
				}

				//Delete destB location. Decrement j so that same index now with different element is checked on next loop
				copy(dests[j:], dests[j+1:])
				dests[len(dests)-1] = SeatsAvailable{}
				dests = dests[:len(dests)-1]
				j--
			}
		}
	}

	//If duplicates were found, alloc new Destination slice and reassign passed in slice pointer
	if len(dests) != originalLength {
		//Create new slice to copy elements over. Original slice will have updated length but old elements in memory (displayed when printing).
		tmp := make([]SeatsAvailable, len(dests))
		for i := 0; i < len(dests); i++ {
			tmp[i] = dests[i]
		}
		*arrayPointer = tmp
	}
}

//Delete terminal self matches from destination array
func deleteTerminalFromDestArray(arrayPointer *[]Destination, targetTerminal Terminal) {
	dests := *arrayPointer
	originalLength := len(dests)

	//Find and remove any matching dests
	for i := 0; i < len(dests); i++ {
		if dests[i].TerminalTitle == targetTerminal.Title {
			copy(dests[i:], dests[i+1:])
			dests[len(dests)-1] = Destination{}
			dests = dests[:len(dests)-1]
			i--
		}
	}
	//If duplicates were found, alloc new Destination slice and reassign passed in slice pointer
	if len(dests) != originalLength {
		//Create new slice to copy elements over. Original slice will have updated length but old elements in memory (displayed when printing).
		tmp := make([]Destination, len(dests))
		for i := 0; i < len(dests); i++ {
			tmp[i] = dests[i]
		}
		*arrayPointer = tmp
	}
}

//Delete destination matches where keyword (minY) is too low on slide
func deleteLowDestsFromDestArray(arrayPointer *[]Destination, sReference Slide) (err error) {
	dests := *arrayPointer
	originalLength := len(dests)

	//Find and remove any matching dests
	for i := 0; i < len(dests); i++ {
		var destKeywordLow bool
		if destKeywordLow, err = sReference.isYCoordinateWithinHeightPercentage(dests[i].BBox.Min.Y, DESTINATION_KEYWORD_VERTICAL_THRESHOLD); err != nil {
			return
		}
		if !destKeywordLow {
			copy(dests[i:], dests[i+1:])
			dests[len(dests)-1] = Destination{}
			dests = dests[:len(dests)-1]
			i--
		}
	}
	//If duplicates were found, alloc new Destination slice and reassign passed in slice pointer
	if len(dests) != originalLength {
		//Create new slice to copy elements over. Original slice will have updated length but old elements in memory (displayed when printing).
		tmp := make([]Destination, len(dests))
		for i := 0; i < len(dests); i++ {
			tmp[i] = dests[i]
		}
		*arrayPointer = tmp
	}
	return
}
