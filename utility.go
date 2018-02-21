package main

import (
	"os"
	"fmt"
	"log"
	"strings"
)

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

//Return local photo path for SaveImageType
func photoPath(slide Slide) string {
	var photoDirectory string
	var photoFilename string
	switch (slide.saveType) {
		case SAVE_IMAGE_TRAINING:
			photoDirectory = IMAGE_TRAINING_DIRECTORY
			break;
		case SAVE_IMAGE_TRAINING_PROCESSED:
			photoDirectory = IMAGE_TRAINING_PROCESSED_DIRECTORY
			break;
		default:
			log.Println("Unknown save type")
	}

	//prefixterminal_title_n.png
	strippedTerminalTitle := strings.Replace(strings.ToLower(slide.terminal.Title), " ", "_", -1)
	photoFilename = fmt.Sprintf("%v_%v.png", strippedTerminalTitle, slide.fbNodeId)

	return fmt.Sprintf("%v/%v", photoDirectory, photoFilename)
}