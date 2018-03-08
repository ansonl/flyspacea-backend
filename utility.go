package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"
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
