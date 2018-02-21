package main

import (
	"os/exec"
)

func runImageMagickConvert(saveTypes []SaveImageType, sReference Slide) (err error) {
	cmd := "convert"

	sReference.saveType = saveTypes[0]
	savePath := photoPath(sReference)

	sReference.saveType = saveTypes[1]
	processedSavePath := photoPath(sReference)

	args := []string{"-alpha", "off", "-fuzz", "35%", "-fill", "white", "+opaque", "black" , savePath, processedSavePath}
	if err = exec.Command(cmd, args...).Run(); err != nil {
		return
	}

	return
}