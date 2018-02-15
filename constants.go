package main

//Graph API URL domain and path. Nodes and edges.
const (
	GRAPH_API_URL     string = "https://graph.facebook.com"
	GRAPH_API_VERSION string = "v2.12"
	GRAPH_EDGE_PHOTOS string = "photos"
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
	SAVE_IMAGE_TRAINING_PROCESSED
)

//Image storage directories
const (
	IMAGE_TRAINING_DIRECTORY string = "training_images"
	IMAGE_TRAINING_PROCESSED_DIRECTORY string = "training_images_processed"
)