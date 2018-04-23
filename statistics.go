package main

import (
	"log"
)

var validTerminals int
var terminalsWith72HRAlbum int
var noErrorTerminals int
var foundFlightsTerminals int

var photosFound int
var photosFoundDateHeader int
var photosProcessed int

var noMatchDateHeaderInputs []string

func resetStatistics() {
	validTerminals = 0
	terminalsWith72HRAlbum = 0
	noErrorTerminals = 0
	foundFlightsTerminals = 0

	photosFound = 0
	photosFoundDateHeader = 0
	photosProcessed = 0
}

func incrementValidTerminals() {
	validTerminals++
}

func incrementTerminalsWith72HRAlbum() {
	terminalsWith72HRAlbum++
}

func incrementNoErrorTerminals() {
	noErrorTerminals++
}

func incrementFoundFlightsTerminals() {
	foundFlightsTerminals++
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
	log.Printf(`Valid Terminals online%v
		Terminals w/ found 72HR Album %v
		Terminals w/ no errors (incl any old photos) %v
		Terminals w/ found flights %v
		Photos Found %v
		Photos Found Date Header %v
		Photos Processed %v`,
		validTerminals,
		terminalsWith72HRAlbum,
		noErrorTerminals,
		foundFlightsTerminals,
		photosFound,
		photosFoundDateHeader,
		photosProcessed)
	//log.Printf("No match date header inputs %v\n", noMatchDateHeaderInputs)
}
