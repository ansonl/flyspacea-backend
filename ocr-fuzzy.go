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
	//Create ban spelling list to not match
	fuzzyBannedSpellings = make(map[string]int)
	fuzzyBannedSpellings["listed"] = 0

	//Slide data header label keyword to train in fuzzy with customizable training depth
	type LabelKeyword struct {
		Spelling     string
		DepthToTrain int
	}

	//Create separate fuzzy model object for each keyword. Store fuzzy models in map
	createFuzzyModelsForKeywords := func (keywords []LabelKeyword, modelMap *map[string]*fuzzy.Model) {
		*modelMap = make(map[string]*fuzzy.Model)
		for _, v := range keywords {
			v.Spelling = strings.ToLower(v.Spelling)
			(*modelMap)[v.Spelling] = fuzzy.NewModel()
			(*modelMap)[v.Spelling].SetThreshold(1)
			(*modelMap)[v.Spelling].SetDepth(v.DepthToTrain)
			(*modelMap)[v.Spelling].TrainWord(v.Spelling)
		}
	}

	//Create fuzzy models for slide labels
	var keywordList []LabelKeyword
	keywordList = []LabelKeyword{
		LabelKeyword{Spelling: KEYWORD_DESTINATION,
			DepthToTrain: 5},
		LabelKeyword{Spelling: KEYWORD_SEATS,
			DepthToTrain: 1}}

	for i := time.January; i <= time.December; i++ {
		fullSpelling := LabelKeyword{Spelling: i.String(),
			DepthToTrain: len(i.String()) / 2}
		shortSpelling := LabelKeyword{Spelling: i.String()[0:3],
			DepthToTrain: len(i.String()[0:3]) / 2}

		keywordList = append(keywordList, fullSpelling, shortSpelling)
	}
	createFuzzyModelsForKeywords(keywordList, &fuzzyModelForKeyword)

	//Create fuzzy models for terminal keywords
	var locationKeywordsArray []Terminal
	if locationKeywordsArray, err = readKeywordsToArrayFromFiles(TERMINAL_FILE, LOCATION_KEYWORDS_FILE); err != nil {
		return
	}

	locationKeywordMap = make(map[string]string)
	fuzzyModelByDepth = make(map[int]*fuzzy.Model)

	//Add keyword to locationKeywordMap and fuzzyModelByDepth
	addKeyword := func(keyword string, title string) {
		//If keyword exists in spelling ban list, do not add it to dict
		//We must also check banned spellings when OCRing because a banned spelling may be too close to an actual keyword
		if _, ok := fuzzyBannedSpellings[title]; ok == true {
			return
		}

		keyword = strings.ToLower(keyword)

		if len(keyword) < 5 {
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

	for _, v := range locationKeywordsArray {

		//Determine title (ex: Hill AFB) without location (ex: Utah)
		trimmed := strings.Split(v.Title, ",")[0]

		//Add trimmed title
		addKeyword(trimmed, v.Title)

		//Add componenets of trimmed title with len() > 5 and not contains parens
		components := strings.FieldsFunc(trimmed, splitRunes)
		for _, k := range components {

			if len(k) > 5 && !strings.Contains(k, "(") && !strings.Contains(k, ")") {
				addKeyword(k, v.Title)
			}
		}

		//Add special keywords
		for _, k := range v.Keywords {
			addKeyword(k, v.Title)
		}
	}

	return
}

//Perform OCR on file for slide and set s.PlainText and s.HOCRText
func doOCRForSlide(s *Slide, wl OCRWhiteListType) (err error) {
	filepath := photoPath(*s)

	client := gosseract.NewClient()
	defer client.Close()
	client.SetPageSegMode(gosseract.PSM_AUTO) //C++ API may have different PSM than command line. https://groups.google.com/d/msg/tesseract-ocr/bD1zJNiDubY/kb7NZPIV38AJ
	client.SetImage(filepath)

	switch(wl) {
	case OCR_WHITELIST_NORMAL:
		client.SetWhitelist("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,:=().*-/")
		break;
	case OCR_WHITELIST_SA:
		client.SetWhitelist("1234567890TFSP")
		break;
	default:
		log.Fatal("Unknown white list type ", wl)
	}
	

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

		//If found spelling exists in spelling ban list, skip it
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

	//log.Println(found)
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

	//XPath query to find ocrx_word element with text node containing target text.
	var xPathQuery string
	xPathQuery = fmt.Sprintf("//*[@class='ocrx_word' and contains(translate(text(), '%v', '%v'), '%v')]", strings.ToUpper(textSpelling), strings.ToLower(textSpelling), strings.ToLower(textSpelling))

	//XPath query to find parent element of element with text node containing target text. Used when target text enclosed in non hocr element.
	var xPathQueryParent string
	xPathQueryParent = fmt.Sprintf("//*[contains(translate(text(), '%v', '%v'), '%v')]/ancestor::span[@class='ocrx_word']", strings.ToUpper(textSpelling), strings.ToLower(textSpelling), strings.ToLower(textSpelling))

	//Try to search for .ocrx_word element containing target text.
	//If no results found, element may be enclosed in <strong> tags, so look for parent.
	var results []xml.Node
	if results, err = html.Search(xPathQuery); err != nil {
		return
	}
	if len(results) == 0 {
		if results, err = html.Search(xPathQueryParent); err != nil {
			return
		}
	}

	//If no results found, print xpathquery and write document to file for debugging
	//Testing http://www.xpathtester.com/xpath/f5f4ce066286d9fdd780998a67d73415
	if len(results) == 0 {
		errorText := fmt.Sprintf("No %v found. Xpathquery %v", textSpelling, xPathQuery)
		//err = fmt.Errorf("No %v found. Xpathquery %v", textSpelling, xPathQuery)
		ioutil.WriteFile("xpath-debug.html", []byte(fmt.Sprintf("%v", doc)), 0644)
		//log.Fatal(err)

		//Display OCR issue
		fmt.Printf("\u001b[1m\u001b[31m%v\u001b[0m\n", errorText)
	}

	for _, r := range results {
		title := r.Attr("title")
		//fmt.Printf("%v\n", len(results))
		fmt.Printf("%v\n", r.String())

		//Regex search for bbox and optional confidence (x_wconf) attr.
		var bboxRegEx *regexp.Regexp
		if bboxRegEx, err = regexp.Compile("bbox ([0-9]*) ([0-9]*) ([0-9]*) ([0-9]*);(?: x_wconf )?(\\d\\d)?"); err != nil {
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

		//Check if x_wconf > OCR_WORD_CONFIDENCE_THRESHOLD if word confidence provided by OCR.
		//Skip element if confidence too low (<threshold)
		if len(bboxMatch[5]) > 0 {
			var wconf int
			if wconf, err = strconv.Atoi(bboxMatch[5]); err != nil {
				return
			}
			if wconf < OCR_WORD_CONFIDENCE_THRESHOLD {
				continue
			}
		}

		bboxes = append(bboxes, image.Rectangle{image.Point{minX, minY}, image.Point{maxX, maxY}})
	}

	return
}
