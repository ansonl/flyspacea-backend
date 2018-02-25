package main

import (
	"log"
)

var photosFound int
var photosProcessed int
var photosFoundDateHeader int

var noMatchDateHeaderInputs []string

func incrementPhotosFound() {
	photosFound++
}

func incrementPhotosProcessed() {
	photosProcessed++
}

func incrementPhotosFoundDateHeader() {
	photosFoundDateHeader++
}

func displayStatistics() {
	log.Printf("Photos Found %v\nPhotos Processed %v\nPhotos Found Date Header %v\n", photosFound, photosProcessed, photosFoundDateHeader)
	//log.Printf("No match date header inputs %v\n", noMatchDateHeaderInputs)
}
