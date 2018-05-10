package main

const (
	DEBUG_EXPORT_TERMINAL_TZ bool = false
	DEBUG_TERMINAL_SINGLE_FILE bool = false
	DEBUG_MANUAL_IMAGE_FILE_TARGET bool = false
	DEBUG_MANUAL_IMAGE_FILE_TARGET_TRAINING_DIRECTORY string = "debug_training"
	DEBUG_MANUAL_FILENAME string = "nsf.jpg" //relative path of image
)

//Terminal info files
const (
	TERMINAL_SINGLE_FILE   string = "terminals-single.json"
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
	GRAPH_FIELD_ID_KEY          string = "id"
	GRAPH_FIELD_NAME_KEY          string = "name"
	GRAPH_TYPE_UPLOADED_KEY       string = "uploaded"

	GRAPH_FIELD_PHONE_KEY        string = "phone"
	GRAPH_FIELD_EMAILS_KEY       string = "emails"
	GRAPH_FIELD_GENERAL_INFO_KEY string = "general_info"
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

//OCR config constants
const (
	FUZZY_MODEL_KEYWORD_MIN_LENGTH int = 5
)

//OCR whitelist names
type OCRWhiteListType int

const (
	OCR_WHITELIST_NORMAL OCRWhiteListType = iota
	OCR_WHITELIST_SA
)

//Tesseract OCR whitelist config filenames
const (
	TESS_CONFIGFILE_DIRECTORY       string = "tesseract_configfiles"
	TESS_CONFIGFILE_NORMAL_FILENAME string = "normal_whitelist"
	TESS_CONFIGFILE_SA_FILENAME     string = "seats_whitelist"
	TESS_CONFIGFILE_HOCR_FILENAME   string = "_hocr"
	TESS_CONFIGFILE_EXTENSION       string = "config"

	TESS_OUTPUTBASE            string = "output"
	TESS_OUTPUT_TXT_EXTENSION  string = "txt"
	TESS_OUTPUT_HOCR_EXTENSION string = "hocr"
)

//OCR keywords
const (
	KEYWORD_DESTINATION string = "destination"
	KEYWORD_SEATS       string = "seats"
)

//Slide processing constants
const (
	//Working color for image processing.
	IMAGE_PROCESSING_TMP_COLOR string = "purple"

	//Minimum OCR word confidence threshold 0-100 to process a result.
	OCR_WORD_CONFIDENCE_THRESHOLD int = 10

	//Max DestinationLabel.YCoord/ImageTotalHeight to detect incorrect Destination label matches.
	DESTINATION_TEXT_VERTICAL_THRESHOLD float64 = 0.5

	//Percentage of area between two bounding boxes to overlap to be considered duplicates.
	DUPLICATE_AREA_THRESHOLD float64 = 0.5

	//Horizontal distance to add to left and right of seats bounds when cropping seats text.
	SEATS_CROP_HORIZONTAL_BUFFER int = 5

	//Furthest vertical distance of destination bounding boxes to be linked.
	ROLLCALLS_DESTINATION_LINK_VERTICAL_THRESHOLD int = 50

	//Minimum vertical overlap required between seats and rollcall text to be linked. Negative value is positive overlap.
	ROLLCALLS_SEATS_LINK_VERTICAL_THRESHOLD int = -5
)

//Storage Database constants
const (
	LOCATIONS_TABLE                         string = "locations"
	FLIGHTS_72HR_TABLE                      string = "hr72_flights"
	FLIGHTS_72HR_TABLE_INDEX_RC             string = "hr72_flights_index_rc"
	FLIGHTS_72HR_TABLE_INDEX_ORIGIN_RC      string = "hr72_flights_index_origin_rc"
	FLIGHTS_72HR_TABLE_INDEX_DEST_RC        string = "hr72_flights_index_dest_rc"
	FLIGHTS_72HR_TABLE_INDEX_ORIGIN_DEST_RC string = "hr72_flights_index_origin_dest_rc"
	PHOTOS_REPORTS_TABLE string = "photo_reports"
)

//Server REST API constants
const (
	REST_ORIGIN_KEY      string = "origin"
	REST_DESTINATION_KEY string = "destination"

	//REST_START_TIME_KEY is UTC time
	//Represented in ISO 1806 / RFC 3339 format 2006-01-02T15:04:05Z
	REST_START_TIME_KEY string = "startTime"

	//REST_DURATION_DAYS_KEY is numbers of days forward from REST_START_TIME_KEY to select.
	REST_DURATION_DAYS_KEY string = "durationDays"

	//PhotoReport keys
	REST_LOCATION_KEY string = "location"
	REST_PHOTOSOURCE_KEY string = "photoSource"
	REST_COMMENT_KEY string = "comment"
)
