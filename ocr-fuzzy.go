package main

import (
	"errors"
	"fmt"
	"github.com/jbowtie/gokogiri/html"
	"github.com/jbowtie/gokogiri/xml"
	"github.com/otiai10/gosseract"
	"github.com/sajari/fuzzy"
	"image"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var fuzzyModelForKeyword map[string]*fuzzy.Model

func createFuzzyModels(modelMap *map[string]*fuzzy.Model) {
	var keywordList []string
	keywordList = []string{KEYWORD_DESTINATION}
	for i := time.January; i <= time.December; i++ {
		keywordList = append(keywordList, i.String(), i.String()[0:3])
	}
	createFuzzyModelsForKeywords(keywordList, modelMap)
}

//Create separate fuzzy model object for each keyword. Store fuzzy models in map
func createFuzzyModelsForKeywords(keywords []string, modelMap *map[string]*fuzzy.Model) {
	*modelMap = make(map[string]*fuzzy.Model)
	for _, v := range keywords {
		(*modelMap)[v] = fuzzy.NewModel()
		(*modelMap)[v].SetThreshold(1)
		(*modelMap)[v].SetDepth(len(v) / 2)
		(*modelMap)[v].TrainWord(v)
	}
}

//Perform OCR on file for slide and set s.PlainText and s.HOCRText
func doOCRForSlide(s *Slide) (err error) {
	filepath := photoPath(*s)

	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage(filepath)
	client.SetWhitelist("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,:=().*")

	if (*s).PlainText, err = client.Text(); err != nil {
		return
	} else if len((*s).PlainText) == 0 {
		displayMessageForSlide((*s), fmt.Sprintf("No plain text extracted from slide"))
	}

	if (*s).HOCRText, err = client.HOCRText(); err != nil {
		return
	} else if len((*s).HOCRText) == 0 {
		displayMessageForSlide(*s, fmt.Sprintf("No hOCR text extracted from slide"))
	}

	return
}

//Find closest spelling of keyword for multiple slides (processed versions)
func findKeywordClosestSpellingInPhotoInSaveImageTypes(keyword string, slides []Slide) (closestSpelling string, closestSpellingSlide Slide, err error) {
	//Store closest spelling of keyword for each image processing type
	var closestKeywordSpellings []string

	for _, s := range slides {
		//Find closest keyword spelling
		foundKeywordSpelling := findKeywordClosestSpellingInPlainText(keyword, s.PlainText)
		if len(foundKeywordSpelling) == 0 {
			displayMessageForSlide(s, fmt.Sprintf("No close spelling extracted from photo type %v", s.SaveType))
		}

		closestKeywordSpellings = append(closestKeywordSpellings, foundKeywordSpelling)
	}

	var closestSpellingDistance int
	for i, v := range closestKeywordSpellings {
		distance := fuzzy.Levenshtein(&v, &keyword)
		if len(closestSpelling) == 0 && len(v) > 0 {
			closestSpelling = v
			closestSpellingDistance = distance
			closestSpellingSlide = slides[i]
		} else if distance < closestSpellingDistance {
			closestSpelling = v
			closestSpellingDistance = distance
			closestSpellingSlide = slides[i]
		}
	}

	if len(closestSpelling) != 0 {
		displayMessageForTerminal(closestSpellingSlide.Terminal, fmt.Sprintf("Close spelling found in save type %v distance %v", closestSpellingSlide.SaveType, closestSpellingDistance))
	}

	return
}

//Find closest spelling of keyword in plain text
func findKeywordClosestSpellingInPlainText(keyword string, plainText string) (closestSpelling string) {

	fuzzyModel := fuzzyModelForKeyword[keyword]

	if fuzzyModel == nil {
		log.Fatal("No fuzzy model for %v", keyword)
	}

	plainText = strings.ToLower(plainText)

	ocrWords := strings.FieldsFunc(plainText, func(c rune) bool {
		return c == ' ' || c == '\n' || c == '\r'
	})

	var closestSpellingDistance int
	for _, ocrWord := range ocrWords {
		//trimmed := strings.Trim(ocrWord, "\r\n ")
		trimmed := ocrWord
		//log.Printf("%v results %q\n", trimmed, fuzzyModel.Suggestions(trimmed, true))
		for _, suggestion := range fuzzyModel.Suggestions(trimmed, true) {
			if suggestion == keyword {
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

func getTextBounds(hocr string, textSpelling string) (bounds image.Rectangle, err error) {
	var doc *html.HtmlDocument
	doc, err = html.Parse([]byte(hocr), xml.DefaultEncodingBytes, nil, xml.DefaultParseOption, xml.DefaultEncodingBytes)
	if err != nil {
		return
	}

	if doc.Root() == nil || doc.Root().CountChildren() == 0 {
		err = errors.New("No root node to document or no children to root")
		return
	}

	html := doc.Root().FirstChild()

	var xPathQuery string
	xPathQuery = fmt.Sprintf("//*[contains(translate(text(), '%v', '%v'), '%v')]/parent::*", strings.ToUpper(textSpelling), strings.ToLower(textSpelling), strings.ToLower(textSpelling))

	results, err := html.Search(xPathQuery)
	if err != nil {
		return
	}

	//If no results found, print xpathquery and write document to file for debugging
	if len(results) == 0 {
		err = fmt.Errorf("No %v found. Xpathquery %v", textSpelling, xPathQuery)
		ioutil.WriteFile("xpath-debug.html", []byte(fmt.Sprintf("%v", doc)), 0644)
		log.Fatal(err)
		return
	}

	title := results[0].Attr("title")
	//displayMessageForTerminal(targetTerminal, fmt.Sprintf("%v\n", len(results)))
	//displayMessageForTerminal(targetTerminal, fmt.Sprintf("%v\n", results[0].String()))

	bboxRegEx, err := regexp.Compile("bbox ([0-9]*) ([0-9]*) ([0-9]*) ([0-9]*);")
	if err != nil {
		return
	}

	bboxMatch := bboxRegEx.FindStringSubmatch(title)

	if len(bboxMatch) == 0 {
		err = errors.New("No bbox found in regex")
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
