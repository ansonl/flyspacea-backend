package main

import (
	"log"
)

var validTerminals int
var terminalsWith72HRAlbum int
var noErrorTerminals int

var photosFound int
var photosFoundDateHeader int
var photosProcessed int

var noMatchDateHeaderInputs []string

func incrementValidTerminals() {
	validTerminals++
}

func incrementTerminalsWith72HRAlbum() {
	terminalsWith72HRAlbum++
}

func incrementNoErrorTerminals() {
	noErrorTerminals++
}

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
	log.Printf(`Valid Terminals online%v\n
		Terminals w/ found 72HR Album %v\n
		Terminals w/ no errors (incl any old photos) %v\n
		Photos Found %v\n
		Photos Found Date Header %v\n
		Photos Processed %v\n`, 
		validTerminals,
		terminalsWith72HRAlbum,
		noErrorTerminals,
		photosFound, 
		photosFoundDateHeader, 
		photosProcessed)
	//log.Printf("No match date header inputs %v\n", noMatchDateHeaderInputs)
}
