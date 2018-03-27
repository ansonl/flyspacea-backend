package main

import (
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"
	"database/sql"
)

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// isDir returns whether given path is a directory
func isDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		return true, nil
	}
	return false, nil
}

//https://stackoverflow.com/a/21067803
// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

//Return local photo path for SaveImageType
func photoPath(slide Slide) string {
	var photoDirectory string
	var photoFilename string
	switch slide.SaveType {
	case SAVE_IMAGE_TRAINING:
		photoDirectory = IMAGE_TRAINING_DIRECTORY
		break
	case SAVE_IMAGE_TRAINING_PROCESSED_BLACK:
		photoDirectory = IMAGE_TRAINING_PROCESSED_DIRECTORY_BLACK
		break
	case SAVE_IMAGE_TRAINING_PROCESSED_WHITE:
		photoDirectory = IMAGE_TRAINING_PROCESSED_DIRECTORY_WHITE
		break
	default:
		log.Println("Unknown save type ", slide.SaveType)
	}

	var suffix string
	if len(slide.Suffix) > 0 {
		suffix = fmt.Sprintf("_%v", slide.Suffix)
	}

	//prefixterminal_title_n_suffix.png
	strippedTerminalTitle := strings.Replace(strings.ToLower(slide.Terminal.Title), " ", "_", -1)

	photoFilename = fmt.Sprintf("%v_%v%v.%v", strippedTerminalTitle, slide.FBNodeId, suffix, slide.Extension)

	return fmt.Sprintf("%v/%v", photoDirectory, photoFilename)
}

//Returns the closer time.Time to the target time.Time of two time.Time
func closerDate(target time.Time, one time.Time, two time.Time) (closerDate time.Time) {
	if math.Abs(float64(target.Sub(one))) < math.Abs(float64(target.Sub(two))) {
		closerDate = one
	} else {
		closerDate = two
	}
	return
}

//Return int vertical distance between two image.Rectangle. Return negative distance if Rectangles intersect.
func getVerticalDistance(bbox1 image.Rectangle, bbox2 image.Rectangle) (verticalDistance int) {

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

func (sReference Slide) getImageConfig() (im image.Config, err error) {
	//find original dimensions
	originalSavePath := photoPath(sReference)
	var reader *os.File
	if reader, err = os.Open(originalSavePath); err != nil {
		return
	}

	//load image
	defer reader.Close()
	if im, _, err = image.DecodeConfig(reader); err != nil {
		return
	}
	return
}

//Check whether two bounding boxes share > DUPLICATE_AREA_THRESHOLD of the smaller box's height on the same horizontal line.
func sameHorizontalLine(bbox1 image.Rectangle, bbox2 image.Rectangle) (horizontalDuplicate bool) {
	smallerHeight := bbox1.Dy()
	if bbox2.Dy() < smallerHeight {
		smallerHeight = bbox2.Dy()
	}

	if bbox1.Min.Y >= bbox2.Min.Y && float64(bbox2.Max.Y-bbox1.Min.Y) > float64(smallerHeight)*DUPLICATE_AREA_THRESHOLD {
		horizontalDuplicate = true
		return
	}

	if bbox2.Min.Y >= bbox1.Min.Y && float64(bbox1.Max.Y-bbox2.Min.Y) > float64(smallerHeight)*DUPLICATE_AREA_THRESHOLD {
		horizontalDuplicate = true
		return
	}

	return
}

//Return true if Y coordinate is less/greater than a certain percentage of slide image height. True = lteq. False = greater
func (sReference Slide) isYCoordinateInHeightPercentage(yCoord int, maxPercentage float64) (valid bool, err error) {
	var im image.Config
	if im, err = sReference.getImageConfig(); err != nil {
		return
	}

	if float64(yCoord) > float64(im.Height) * maxPercentage {
		valid = false
	} else {
		valid = true
	}
	return
}

//Check *(sql.DB) handle initialized and connected
func checkDatabaseHandleValid(targetHandle *(sql.DB)) (err error) {
	if db == nil {
		err = fmt.Errorf("DB handle is nil")
		return

	}
	
	if err = db.Ping(); err != nil {
		err = fmt.Errorf("DB ping failed.")
		return
	}
	return
}