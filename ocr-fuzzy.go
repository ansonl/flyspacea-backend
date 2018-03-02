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

var locationKeywordMap map[string]string
var fuzzyModelByDepth map[int]*fuzzy.Model
var fuzzyBannedSpellings map[string]int

//Create fuzzy models for slide labels and terminal keywords
func createFuzzyModels() (err error) {
	//Create fuzzy models for slide labels
	var keywordList []string
	keywordList = []string{KEYWORD_DESTINATION}
	for i := time.January; i <= time.December; i++ {
		keywordList = append(keywordList, i.String(), i.String()[0:3])
	}
	createFuzzyModelsForKeywords(keywordList, &fuzzyModelForKeyword)

	//Create fuzzy models for terminal keywords
	var terminalKeywordsArray []TerminalKeywords
	if terminalKeywordsArray, err = readTerminalKeywordsFileToArray(TERMINAL_KEYWORDS_FILE); err != nil {
		return
	}

	locationKeywordMap = make(map[string]string)
	fuzzyModelByDepth = make(map[int]*fuzzy.Model)

	//Add keyword to locationKeywordMap and fuzzyModelByDepth
	addKeyword := func(keyword string, title string) {
		keyword = strings.ToLower(keyword)

		if (len(keyword) < 5) {
			err = fmt.Errorf("Keyword length less than 5. %v", keyword)
		}

		//Add to keyword -> terminal title map
		locationKeywordMap[keyword] = title

		//Determine depth
		depth := len(keyword) / 2

		//Limit depth for speed and false positives
		limit := 2
		if depth > limit {
			depth = limit
		}

		//Create fuzzy model at depth if needed
		if fuzzyModelByDepth[depth] == nil {
			fuzzyModelByDepth[depth] = fuzzy.NewModel()
			fuzzyModelByDepth[depth].SetThreshold(1)
			fuzzyModelByDepth[depth].SetDepth(depth)
		}

		//Add to fuzzy model at depth
		log.Printf("training %v", keyword)
		fuzzyModelByDepth[depth].TrainWord(keyword)
		
	}

	//Split runes
	splitRunes := func(r rune) bool {
		return r == ' ' || r == '-'
	}

	for _, v := range terminalKeywordsArray {

		//Determine title (ex: Hill AFB) without location (ex: Utah)
		trimmed := strings.Split(v.Title, ",")[0]

		//Add trimmed title
		addKeyword(trimmed, v.Title)

		//Add componenets of trimmed title with len() > 5 and not contains parens
		components := strings.FieldsFunc(trimmed, splitRunes)
		for _, k := range components {

			if len(k) > 5  && !strings.Contains(k, "(") && !strings.Contains(k, ")") {
				addKeyword(k, v.Title)
			}
		}

		//Add special keywords
		for _, k := range v.Keywords {
			addKeyword(k, v.Title)
		}
	}

	//Create ban spelling list to not match
	fuzzyBannedSpellings = make(map[string]int)
	fuzzyBannedSpellings["listed"] = 0

	return
}

//Create separate fuzzy model object for each keyword. Store fuzzy models in map
func createFuzzyModelsForKeywords(keywords []string, modelMap *map[string]*fuzzy.Model) {
	*modelMap = make(map[string]*fuzzy.Model)
	for _, v := range keywords {
		v = strings.ToLower(v)
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
	client.SetWhitelist("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,:=().*-/")

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
		//fmt.Println("SAVETYPE ", s.SaveType, photoPath(s))

		//Find closest keyword spelling
		foundKeywordSpelling := findKeywordClosestSpellingInPlainText(keyword, s.PlainText)
		if len(foundKeywordSpelling) == 0 {
			//displayMessageForSlide(s, fmt.Sprintf("No close spelling extracted from photo type %v", s.SaveType))
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

	/*
		//Print closest spelling and distance
		if len(closestSpelling) != 0 {
			displayMessageForTerminal(closestSpellingSlide.Terminal, fmt.Sprintf("Close spelling %v found in save type %v distance %v", closestSpelling, closestSpellingSlide.SaveType, closestSpellingDistance))
		}
	*/

	return
}

//Find closest spelling of keyword in plain text
func findKeywordClosestSpellingInPlainText(keyword string, plainText string) (closestSpelling string) {

	//lowercase keyword and plaintext
	keyword = strings.ToLower(keyword)
	plainText = strings.ToLower(plainText)

	fuzzyModel := fuzzyModelForKeyword[keyword]

	if fuzzyModel == nil {
		log.Fatal("No fuzzy model for %v", keyword)
	}

	/*
		fmt.Println("keyword ", keyword)
		fmt.Println("plaintext", plainText)
	*/

	//Split by the special characters in our whitelist including \r and \n
	ocrWords := strings.FieldsFunc(plainText, func(c rune) bool {
		return c == ' ' || c == '\n' || c == '\r' || c == ',' || c == ':' || c == '=' || c == '(' || c == ')' || c == '.' || c == '*' || c == '-' || c == '/'
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

//Find all best terminal keyword matches for every word in plaintext. Return map[spelling]TerminalKeywordsResult{Keyword, Distance}
func findTerminalKeywordsInPlainText(plainText string) (found map[string]TerminalKeywordsResult) {
	found = make(map[string]TerminalKeywordsResult)

	//lowercase keyword and plaintext
	plainText = strings.ToLower(plainText)
	//log.Println(plainText)

	//Split by the special characters in our whitelist including \r and \n
	ocrWords := strings.FieldsFunc(plainText, func(c rune) bool {
		return c == ' ' || c == '\n' || c == '\r' || c == ',' || c == ':' || c == '=' || c == '(' || c == ')' || c == '.' || c == '*' || c == '-' || c == '/'
	})

	//Search single word
	for _, ocrWord := range ocrWords {



		//If word exists in spelling ban list, ignore it
		if _, ok := fuzzyBannedSpellings[ocrWord]; ok == true {
			continue
		}

		//Find best match from all fuzzy models for current word
		var closestSuggestion string
		var closestSpelling string
		var closestSpellingDistance int
		for depth, fuzzyModel := range fuzzyModelByDepth {
			if fuzzyModel == nil {
				log.Fatal("No fuzzy model for depth %v", depth)
			}

			for _, suggestion := range fuzzyModel.Suggestions(ocrWord, true) {

				distance := fuzzy.Levenshtein(&ocrWord, &suggestion)
				if len(closestSpelling) == 0 && len(ocrWord) > 0 {
					closestSuggestion = suggestion
					closestSpelling = ocrWord
					closestSpellingDistance = distance
				} else if distance < closestSpellingDistance {
					closestSuggestion = suggestion
					closestSpelling = ocrWord
					closestSpellingDistance = distance
				}
			}
		}

		/*
		//Unneeded because we are running each spelling against all fuzzy models so optimal suggestion will be found the first time.
		//Find best match of all found suggestions (keywords)
		if len(closestSuggestion) > 0 && found[closestSpelling] != nil && closestSpellingDistance < found[closestSpelling].Distance  {
			found[closestSpelling] = new TerminalKeyWordsResult{closestSuggestion, closestSpellingDistance}
		}
		*/

		//Add to found spelling map
		_, ok := found[closestSpelling]
		if !ok && len(closestSuggestion) > 0 {
			tmp := TerminalKeywordsResult{Keyword: closestSuggestion, Distance: closestSpellingDistance}
			found[closestSpelling] = tmp
		}
	}

	//Search two words
	var prevWord string
	for _, ocrWord := range ocrWords {
		if len(prevWord) == 0 {
			prevWord = ocrWord
			continue
		}

		prevLength := len(prevWord)
		curLength := len(ocrWord)
		twoWord := fmt.Sprintf("%v %v", prevWord, ocrWord)
		
		//log.Println(twoWord)

		//If word exists in spelling ban list, ignore it
		if _, ok := fuzzyBannedSpellings[twoWord]; ok == true {
			continue
		}

		//Find best match from all fuzzy models for current word
		var closestSuggestion string
		var closestSpelling string
		var closestSpellingDistance int
		for depth, fuzzyModel := range fuzzyModelByDepth {
			if fuzzyModel == nil {
				log.Fatal("No fuzzy model for depth %v", depth)
			}

			for _, suggestion := range fuzzyModel.Suggestions(twoWord, true) {
				distance := fuzzy.Levenshtein(&twoWord, &suggestion)
				if len(closestSpelling) == 0 && len(twoWord) > 0 {
					closestSuggestion = suggestion

					if prevLength >= curLength {
						closestSpelling = prevWord
					} else {
						closestSpelling = ocrWord
					}
					
					closestSpellingDistance = distance
				} else if distance < closestSpellingDistance {
					closestSuggestion = suggestion
					
					if prevLength >= curLength {
						closestSpelling = prevWord
					} else {
						closestSpelling = ocrWord
					}

					closestSpellingDistance = distance
				}
			}
		}

		/*
		//Unneeded because we are running each spelling against all fuzzy models so optimal suggestion will be found the first time.
		//Find best match of all found suggestions (keywords)
		if len(closestSuggestion) > 0 && found[closestSpelling] != nil && closestSpellingDistance < found[closestSpelling].Distance  {
			found[closestSpelling] = new TerminalKeyWordsResult{closestSuggestion, closestSpellingDistance}
		}
		*/

		//Add to found spelling map
		_, ok := found[closestSpelling]
		if !ok && len(closestSuggestion) > 0 {
			tmp := TerminalKeywordsResult{Keyword: closestSuggestion, Distance: closestSpellingDistance}
			found[closestSpelling] = tmp
		}

		prevWord = ocrWord
	}

	log.Println(found)
	return

}

func getTextBounds(hocr string, textSpelling string) (bboxes []image.Rectangle, err error) {
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

	var results []xml.Node
	results, err = html.Search(xPathQuery)
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


	for _, r := range results {
		title := r.Attr("title")
		//fmt.Printf("%v\n", len(results))
		//fmt.Printf("%v\n", results[0].String())

		var bboxRegEx *regexp.Regexp
		if bboxRegEx, err = regexp.Compile("bbox ([0-9]*) ([0-9]*) ([0-9]*) ([0-9]*);"); err != nil {
			return
		}

		bboxMatch := bboxRegEx.FindStringSubmatch(title)

		if len(bboxMatch) == 0 {
			err = errors.New("No bbox found in regex")
			return
		}

		var minX, minY, maxX, maxY int

		if minX, err = strconv.Atoi(bboxMatch[1]); err != nil {
			return
		}
		if minY, err = strconv.Atoi(bboxMatch[2]); err != nil {
			return
		}
		if maxX, err = strconv.Atoi(bboxMatch[3]); err != nil {
			return
		}
		if maxY, err = strconv.Atoi(bboxMatch[4]); err != nil {
			return
		}

		bboxes = append(bboxes, image.Rectangle{image.Point{minX, minY}, image.Point{maxX, maxY}})
	}
	
	return
}
