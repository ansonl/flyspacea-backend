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

var fuzzyModelForKeyword map[string] *fuzzy.Model

//Create separate fuzzy model object for each keyword. Store fuzzy models in map
func createFuzzyModelsForKeywords(keywords []string, modelMap *map[string] *fuzzy.Model) {
	*modelMap = make(map[string] *fuzzy.Model)
	for _, v := range keywords {
		(*modelMap)[v] = fuzzy.NewModel()
		(*modelMap)[v].SetThreshold(1)
		(*modelMap)[v].SetDepth(5)
		(*modelMap)[v].TrainWord(v)
	}
}

//Find closest spelling of keyword in plain text
func findKeywordClosestSpellingInPlainText (keyword string, plainText string) (closestSpelling string) {

	fuzzyModel := fuzzyModelForKeyword[keyword]

	if fuzzyModel == nil {
		log.Fatal("No fuzzy model for %v", keyword)
	}

	plainText = strings.ToLower(plainText)

	ocrWords := strings.FieldsFunc(plainText, func(c rune) bool {
		return c == ' ' || c == '\n' || c =='\r'
		})

	//log.Printf("ocr %q\n", ocrWords)
	var closestSpellingDistance int
	for _, ocrWord := range ocrWords {
		//trimmed := strings.Trim(ocrWord, "\r\n ")
		trimmed := ocrWord
		//log.Printf("%v results %q\n", trimmed, fuzzyModel.Suggestions(trimmed, true))
		for _, suggestion := range fuzzyModel.Suggestions(trimmed, true) {
			if (suggestion == keyword) {
				distance := fuzzy.Levenshtein(&trimmed, &keyword)
				if len(closestSpelling) == 0 && len(trimmed) > 0 {
					closestSpelling = trimmed
					closestSpellingDistance = distance
				} else if distance < closestSpellingDistance {
					closestSpelling = trimmed
					closestSpellingDistance = distance
				}
			}
		}
	}
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

//Find closest spelling of keyword for multiple processed types of a photo
func findKeywordClosestSpellingInPhotoInSaveImageTypes(keyword string, targetTerminal Terminal, photoNumber int, saveTypes []SaveImageType) (closestSpelling string) {
	//Store closest spelling of keyword for each image processing type
	var closestKeywordSpellings []string

	for _, v := range saveTypes {
		//Get OCR plaintext from original image
		plainText, ocrError := getPlainText(v, targetTerminal, photoNumber)
		if ocrError != nil {
			log.Println(ocrError)
			continue
		}
		if len(plainText) == 0 {
			log.Printf("No plain text extracted from photo %v type %v\n", photoNumber, v)
		}

		//Find clostest keyword spelling
		foundKeywordSpelling := findKeywordClosestSpellingInPlainText(keyword, plainText)
		if len(foundKeywordSpelling) == 0 {
			log.Printf("No close spelling extracted from photo %v type %v\n", photoNumber, v)
		}

		closestKeywordSpellings = append(closestKeywordSpellings, foundKeywordSpelling)
	}

	var closestSpellingDistance int
	var closestSpellingSaveType SaveImageType
	for i, v := range closestKeywordSpellings {
		distance := fuzzy.Levenshtein(&v, &keyword)
		if len(closestSpelling) == 0 && len(v) > 0 {
			closestSpelling = v
			closestSpellingDistance = distance
			closestSpellingSaveType = saveTypes[i]
		} else if distance < closestSpellingDistance {
			closestSpelling = v
			closestSpellingDistance = distance
			closestSpellingSaveType = saveTypes[i]
		}
	}

	if len(closestSpelling) != 0 {
		log.Printf("Closest spelling found in save type %v\n", closestSpellingSaveType)
	}

	return
}