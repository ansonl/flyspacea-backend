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
			//OR if same destination terminal and on 50% same horizontal line due to fuzzy match individual words from same location
			var horizontalDuplicate bool
			smallerHeight := destA.BBox.Dy()
			if destB.BBox.Dy() < smallerHeight {
				smallerHeight = destB.BBox.Dy()
			}

			if (destA.TerminalTitle == destB.TerminalTitle) {
				if (destA.BBox.Min.Y >= destB.BBox.Min.Y && float64(destB.BBox.Max.Y - destA.BBox.Min.Y) > float64(smallerHeight)*DUPLICATE_AREA_THRESHOLD) {
					horizontalDuplicate = true
				}

				if (destB.BBox.Min.Y >= destA.BBox.Min.Y && float64(destA.BBox.Max.Y - destB.BBox.Min.Y) > float64(smallerHeight)*DUPLICATE_AREA_THRESHOLD) {
					horizontalDuplicate = true
				}
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
