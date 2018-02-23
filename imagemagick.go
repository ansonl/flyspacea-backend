package main

import (
	"os/exec"
	"errors"
	"log"
)

//Run image color filter on a slide from 'sourceSaveType' directory to sReference.saveType location
func runImageMagickConvert(sourceSaveType SaveImageType, sReference Slide) (err error) {
	cmd := "convert"

	//Create tmp orginal slide struct to pass photoPath() to get original source image path
	originalSlide := sReference
	originalSlide.saveType = sourceSaveType
	originalSavePath := photoPath(originalSlide)

	processedSavePath := photoPath(sReference)

	//variables for replace color
	var replaceColor bool
	var fill string
	var opaque string
	
	//variables for invert color
	var invertColor bool

	switch (sReference.saveType) {
		case SAVE_IMAGE_TRAINING_PROCESSED_BLACK:
			replaceColor = true
			fill = "white"
			opaque = "black"
			break;
		case SAVE_IMAGE_TRAINING_PROCESSED_WHITE:
			replaceColor = true
			fill = "black"
			opaque = "white"
			invertColor = true
			break;
		default:
			err = errors.New("Unknown save type for image convert")
			return
	}

	if replaceColor {
		args := []string{"-alpha", "off", "-fuzz", "35%", "-fill", fill, "+opaque", opaque , originalSavePath, processedSavePath}
		log.Printf("%v", args)
		if err = exec.Command(cmd, args...).Run(); err != nil {
			return
		}
	}

	if invertColor {
		args := []string{processedSavePath, "-negate", processedSavePath}
		if err = exec.Command(cmd, args...).Run(); err != nil {
			return
		}
	}

	return
}