package main

import (
	"fmt"
)

//Need a very tall display to run, issue is that cursor cannot move "up" off the screen if the term has too few rows

var errorRowCount int

func setupDisplayForTerminalArray(terminalArray []Terminal) {
	for i, v := range terminalArray {
		//Bold/unbold - \u001b[1m \u001b[0m
		fmt.Printf("\u001b[1m%v %v\u001b[0m\n\n", v.Title, i)
	}
}

/*
func displayMessageForTerminal(t Terminal, message string) {
	var cursorLinesOffset int
	cursorLinesOffset = 1 + t.OffsetUp * 2
	//Carriage return to beginning of line - \r
	//Move cursor %v lines up - u001b[%vF
	//Clear entire line - \u001b[2K
	//Move cursor %v lines down - u001b[%vE
	fmt.Printf("\r\u001b[%vF\u001b[0K%v %v %v\u001b[%vE\r", cursorLinesOffset, t.OffsetUp, cursorLinesOffset, message, cursorLinesOffset)
}
*/

//Drop in function for normal printing
func displayMessageForTerminal(t Terminal, message string) {
	fmt.Printf("%v - %v\n", t.Title, message)
}

//Drop in function for normal printing
func displayMessageForSlide(s Slide, message string) {
	fmt.Printf("%v %v %v - %v\n", s.Terminal.Title, s.SaveType, s.FBNodeId, message)
}

/*
func displayErrorForTerminalAtBottom(t Terminal, message string) {
	//Carriage return to beginning of line - \r
	//Move cursor %v lines down - u001b[%vE
	//Insert line - \u001b[L
	//Bold/unbold - \u001b[1m \u001b[0m
	//Move cursor %v lines up - u001b[%vF
	fmt.Printf("\r\u001b[%vB\u001b[1m%v\u001b[0m %v\u001b[%vA\r", errorRowCount, t.Title, message, errorRowCount)

	errorRowCount++
}
*/
//Drop in function for normal printing
func displayErrorForTerminal(t Terminal, message string) {
	fmt.Printf("\u001b[1m\u001b[31m%v\u001b[0m - %v\n", t.Title, message)
}

func endDisplayForTerminalArray(terminalArray []Terminal) {
	fmt.Printf("\r\u001b[%vE", errorRowCount)
}
