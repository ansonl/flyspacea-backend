package main

import (
	"fmt"
	"log"
	"os"
	"strings"
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

	//prefixterminal_title_n_suffix.png
	strippedTerminalTitle := strings.Replace(strings.ToLower(slide.Terminal.Title), " ", "_", -1)

	photoFilename = fmt.Sprintf("%v_%v%v.png", strippedTerminalTitle, slide.FBNodeId, slide.Suffix)

	return fmt.Sprintf("%v/%v", photoDirectory, photoFilename)
}
