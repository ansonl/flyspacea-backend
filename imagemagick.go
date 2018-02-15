package main

import (
	"os"
	"os/exec"
	"fmt"
)

func runImageMagickConvert(targetTerminal Terminal, photoNumber int) {
	cmd := "convert"
	savePath := photoPath(SAVE_IMAGE_TRAINING, "", targetTerminal, photoNumber)
	processedSavePath := photoPath(SAVE_IMAGE_TRAINING_PROCESSED, "", targetTerminal, photoNumber)
	args := []string{"-alpha", "off", "-fuzz", "35%", "-fill", "white", "+opaque", "black" , savePath, processedSavePath}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Println(fmt.Sprintf("%v", processedSavePath))
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}