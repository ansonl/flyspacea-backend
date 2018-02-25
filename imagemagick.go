package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	//"log"
)

//Run image color filter on a slide from 'sourceSaveType' directory to sReference.saveType location
func runImageMagickColorProcess(sourceSaveType SaveImageType, sReference Slide) (err error) {
	cmd := "convert"

	//Create tmp orginal slide struct to pass photoPath() to get original source image path
	originalSlide := sReference
	originalSlide.SaveType = sourceSaveType
	originalSavePath := photoPath(originalSlide)

	processedSavePath := photoPath(sReference)

	//variables for replace color
	var replaceColor bool
	var fill string
	var opaque string

	//variables for invert color
	var invertColor bool

	switch sReference.SaveType {
	case SAVE_IMAGE_TRAINING_PROCESSED_BLACK:
		replaceColor = true
		fill = "white"
		opaque = "black"
		break
	case SAVE_IMAGE_TRAINING_PROCESSED_WHITE:
		replaceColor = true
		fill = "black"
		opaque = "white"
		invertColor = true
		break
	default:
		err = errors.New("Unknown save type for image convert")
		return
	}

	if replaceColor {
		args := []string{"-alpha", "off", "-fuzz", "35%", "-fill", fill, "+opaque", opaque, originalSavePath, processedSavePath}
		//log.Printf("%v", args)
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

//Run image crop on a slide from sReference.saveType location to some directory with suffix on output filename
func runImageMagickDateCropProcess(sReference Slide, cropVerticalOffset int, cropHeight int) (err error) {

	//Create new save patch with suffix to indicate crop
	originalSavePath := photoPath(sReference)
	sReference.Suffix = IMAGE_SUFFIX_CROPPED
	processedSavePath := photoPath(sReference)

	//find original dimensions
	var reader *os.File
	if reader, err = os.Open(originalSavePath); err != nil {
		return
	}

	//load image
	defer reader.Close()
	var im image.Config
	if im, _, err = image.DecodeConfig(reader); err != nil {
		return
	}

	//run imagemagick crop
	cmd := "convert"
	args := []string{"-crop", fmt.Sprintf("%vx%v+%v+%v", im.Width, cropHeight, 0, cropVerticalOffset), originalSavePath, processedSavePath}
	//log.Printf("%v", args)
	if err = exec.Command(cmd, args...).Run(); err != nil {
		return
	}

	return
}
