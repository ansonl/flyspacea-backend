package main

//Terminal info files
const (
	TERMINAL_FILE          string = "terminals.json"
	LOCATION_KEYWORDS_FILE string = "location_keywords.json"
)

//Graph API URL domain and path. Nodes and edges.
const (
	GRAPH_API_URL     string = "https://graph.facebook.com"
	GRAPH_API_VERSION string = "v2.12"
	GRAPH_EDGE_PHOTOS string = "photos"
	GRAPH_EDGE_ALBUMS string = "albums"
)

//Graph API parameter keys
const (
	GRAPH_ACCESS_TOKEN_KEY string = "access_token"
	GRAPH_FIELDS_KEY       string = "fields"
	GRAPH_TYPE_KEY         string = "type"
)

//Graph API parameter key-values and key-values' related returned map keys
const (
	GRAPH_ACCESS_TOKEN            string = "522755171230853%7ChxS9OzJ4I0CqmmrESRpNHfx77vs"
	GRAPH_FIELD_IMAGES_KEY        string = "images"
	GRAPH_FIELD_IMAGES_SOURCE_KEY string = "source"
	GRAPH_FIELD_UPDATED_TIME_KEY  string = "updated_time"
	GRAPH_FIELD_NAME_KEY          string = "name"
	GRAPH_TYPE_UPLOADED_KEY       string = "uploaded"
)

//Graph API returned map keys
const (
	GRAPH_DATA_KEY string = "data"
	GRAPH_ID_KEY   string = "id"
)

//Image storage types
type SaveImageType int

const (
	SAVE_IMAGE_TRAINING SaveImageType = iota
	SAVE_IMAGE_TRAINING_PROCESSED_BLACK
	SAVE_IMAGE_TRAINING_PROCESSED_WHITE
)

//Image storage directories
const (
	IMAGE_TMP_DIRECTORY                      string = "tmp"
	IMAGE_TRAINING_DIRECTORY                 string = "training_images"
	IMAGE_TRAINING_PROCESSED_DIRECTORY_BLACK string = "training_images_processed_black"
	IMAGE_TRAINING_PROCESSED_DIRECTORY_WHITE string = "training_images_processed_white"
)

//Image storage suffixes
const (
	IMAGE_SUFFIX_CROPPED string = "c"
)

//OCR whitelist names
type OCRWhiteListType int

const (
	OCR_WHITELIST_NORMAL OCRWhiteListType = iota
	OCR_WHITELIST_SA
)

//OCR keywords
const (
	KEYWORD_DESTINATION string = "destination"
	KEYWORD_SEATS       string = "seats"
)

//Slide processing constants
const (
	IMAGE_PROCESSING_TMP_COLOR string = "purple"

	OCR_WORD_CONFIDENCE_THRESHOLD int     = 10
	DUPLICATE_AREA_THRESHOLD      float64 = 0.5

	SEATS_CROP_HORIZONTAL_BUFFER int = 5

	ROLLCALLS_DESTINATION_LINK_VERTICAL_THRESHOLD int = 50
	ROLLCALLS_SEATS_LINK_VERTICAL_THRESHOLD       int = -5
)

//Storage Database constants
const (
	LOCATIONS_TABLE string = "locations"
	FLIGHTS_72HR_TABLE string = "HR72_flights"
)