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

	workingSavePath := processedSavePath + "m"

	//variables for replace color
	var fill string
	var opaque1, opaque2 string
	var fuzz string

	//variables for invert color

	var args []string
	switch sReference.SaveType {
	case SAVE_IMAGE_TRAINING_PROCESSED_BLACK:
		opaque1 = "black"
		opaque2 = "white"
		fill = "black"
		fuzz = "25%"

		/*
					convert bwi.png -fuzz 25% -fill purple -opaque black +transparent purple bwi_m.png
			convert bwi.png -fuzz 25% -fill purple -opaque white +transparent purple bwi_m.png -composite bwi_p25.png
			convert bwi_p25.png -fill black -opaque purple -transparent-color white -flatten bwi_p25.jpeg
		*/

		args = []string{originalSavePath, "-fuzz", fuzz, "-fill", IMAGE_PROCESSING_TMP_COLOR, "+opaque", opaque1, "+transparent", IMAGE_PROCESSING_TMP_COLOR, workingSavePath}
		if err = exec.Command(cmd, args...).Run(); err != nil {
			return
		}

		args = []string{originalSavePath, "-fuzz", fuzz, "-fill", IMAGE_PROCESSING_TMP_COLOR, "+opaque", opaque2, "+transparent", IMAGE_PROCESSING_TMP_COLOR, workingSavePath, "-composite", processedSavePath}
		if err = exec.Command(cmd, args...).Run(); err != nil {
			return
		}

		args = []string{processedSavePath, "-fill", fill, "-opaque", IMAGE_PROCESSING_TMP_COLOR, "-transparent-color", "white", "-flatten", processedSavePath}
		if err = exec.Command(cmd, args...).Run(); err != nil {
			return
		}

		break
	case SAVE_IMAGE_TRAINING_PROCESSED_WHITE:
		fill = "black"
		opaque1 = "white"
		fuzz = "35%"

		args = []string{"-alpha", "off", "-fuzz", fuzz, "-fill", fill, "+opaque", opaque1, originalSavePath, processedSavePath}
		//log.Printf("%v", args)
		if err = exec.Command(cmd, args...).Run(); err != nil {
			return
		}

		args = []string{processedSavePath, "-negate", processedSavePath}
		if err = exec.Command(cmd, args...).Run(); err != nil {
			return
		}

		break
	default:
		err = errors.New("Unknown save type for image convert")
		return
	}

	return
}

//Run image crop on a slide from sReference.saveType location to some directory with suffix on output filename.
//Crop vertically to form horizonal row image
func runImageMagickDateCropVerticalProcess(sReference Slide, cropVerticalOffset int, cropHeight int) (err error) {

	//find original dimensions
	originalSavePath := photoPath(sReference)
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

	err = runImageMagickCropProcess(sReference, []int{im.Width, cropHeight, 0, cropVerticalOffset})

	return
}

//Run image crop on a slide from sReference.saveType location to some directory with suffix on output filename.
//Crop horizontal starting top at cropVerticalOffset to form vertical column image
func runImageMagickDateCropHorizontalProcess(sReference Slide, cropHorizontalMinMax image.Point, cropVerticalOffset int) (err error) {

	var im image.Config
	if im, err = sReference.getImageConfig(); err != nil {
		return
	}

	err = runImageMagickCropProcess(sReference, []int{cropHorizontalMinMax.Y - cropHorizontalMinMax.X, im.Height - cropVerticalOffset, cropHorizontalMinMax.X, cropVerticalOffset})

	return
}

//Run image crop with bbox with geometry params
func runImageMagickCropProcess(sReference Slide, cropGeometry []int) (err error) {
	if len(cropGeometry) != 4 {
		err = fmt.Errorf("crop geometry param is not length 4 is length ", len(cropGeometry))
	}

	//Create new save patch with suffix to indicate crop
	originalSavePath := photoPath(sReference)
	sReference.Suffix = IMAGE_SUFFIX_CROPPED
	processedSavePath := photoPath(sReference)

	//run imagemagick crop
	//http://www.imagemagick.org/script/command-line-processing.php#geometry
	//{size}{+-}x{+-}y
	cmd := "convert"
	args := []string{"-crop", fmt.Sprintf("%vx%v+%v+%v", cropGeometry[0], cropGeometry[1], cropGeometry[2], cropGeometry[3]), originalSavePath, processedSavePath}
	fmt.Printf("%v", args)
	if err = exec.Command(cmd, args...).Run(); err != nil {
		return
	}

	return
}
