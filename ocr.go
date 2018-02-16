package main

import (
	"github.com/otiai10/gosseract"
	"github.com/jbowtie/gokogiri/html"
	"github.com/jbowtie/gokogiri/xml"
	"github.com/sajari/fuzzy"
	"image"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var fuzzyModelLabels *fuzzy.Model;

func findDestinationInPlainText (plainText string) (actualSpelling string) {

	var keyword = "destination"
	if fuzzyModelLabels == nil {
		fuzzyModelLabels = fuzzy.NewModel()
		fuzzyModelLabels.SetThreshold(1)
		fuzzyModelLabels.SetDepth(5)
		fuzzyModelLabels.TrainWord(keyword)
	}

	plainText = strings.ToLower(plainText)

	ocrWords := strings.FieldsFunc(plainText, func(c rune) bool {
		return c == ' ' || c == '\n' || c =='\r'
		})

	//log.Printf("ocr %q\n", ocrWords)
	var closestMatch string
	var closestMatchDistance int
	for _, ocrWord := range ocrWords {
		//trimmed := strings.Trim(ocrWord, "\r\n ")
		trimmed := ocrWord
		//log.Printf("%v results %q\n", trimmed, fuzzyModelLabels.Suggestions(trimmed, true))
		for _, suggestion := range fuzzyModelLabels.Suggestions(trimmed, true) {
			if (suggestion == keyword) {
				distance := fuzzy.Levenshtein(&closestMatch, &keyword)
				if len(closestMatch) == 0 {
					closestMatch = trimmed
					closestMatchDistance = distance
				} else if distance < closestMatchDistance {
					closestMatch = trimmed
					closestMatchDistance = distance
				}
			}
		}
	}

	actualSpelling = closestMatch
	return
}

func getPlainText(saveType SaveImageType, targetTerminal Terminal, photoNumber int) (text string, err error) {
	filepath := photoPath(saveType, "", targetTerminal, photoNumber)

	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage(filepath)
	client.SetWhitelist("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,:=")
	
	text, err = client.Text()
	return
}

func getHOCRText(saveType SaveImageType, targetTerminal Terminal, photoNumber int) (hocr string, err error) {
	filepath := photoPath(saveType, "", targetTerminal, photoNumber)

	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage(filepath)
	client.SetWhitelist("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,:=")
	
	hocr, err = client.HOCRText()
	return
}

func getDestinationBounds(hocr string) (bounds image.Rectangle, err error) {
	doc, err := html.Parse([]byte(hocr), xml.DefaultEncodingBytes, nil, xml.DefaultParseOption, xml.DefaultEncodingBytes)
	if err != nil {
		return
	}

	html := doc.Root().FirstChild()
	results, err := html.Search("//strong[contains(translate(text(), 'DEST', 'dest'), 'dest')]/parent::*")
	if err != nil {
		return
	}

	if len(results) == 0 {
		log.Println("No dest found")
		return
	}

	title := results[0].Attr("title")
	log.Printf("%v", len(results))
	log.Printf("%v", results[0].String())

	bboxRegEx, err := regexp.Compile("bbox ([0-9]*) ([0-9]*) ([0-9]*) ([0-9]*);")
	if err != nil {
		return
	} 
	
	bboxMatch := bboxRegEx.FindStringSubmatch(title)

	if len(bboxMatch) == 0 {
		log.Println("No bbox found in regex")
		return
	}

	minX, err := strconv.Atoi(bboxMatch[1])
	if err != nil {
		return
	} 
	minY, err := strconv.Atoi(bboxMatch[2])
	if err != nil {
		return
	} 
	maxX, err := strconv.Atoi(bboxMatch[3])
	if err != nil {
		return
	} 
	maxY, err := strconv.Atoi(bboxMatch[4])
	if err != nil {
		return
	} 

	bounds = image.Rectangle{image.Point{minX, minY}, image.Point{maxX, maxY}}

	return
}