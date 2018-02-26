package main

//Facebook Graph API structs for retrieving page photos
//https://developers.facebook.com/docs/graph-api/reference/page/photos/
//?type=uploaded
type PhotosEdge struct {
	Data  []PhotosEdgePhoto  `json:"data"`
	Error GraphErrorResponse `json:"error"`
}

type PhotosEdgePhoto struct {
	CreatedTime string `json:"created_time"`
	Name        string `json:"name"`
	Id          string `json:"id"`
}

//https://developers.facebook.com/docs/graph-api/reference/photo/
//?fields=images
type PhotoNode struct {
	Images []PhotoNodeImage   `json:"images"`
	Error  GraphErrorResponse `json:"error"`
}

type PhotoNodeImage struct {
	Source string `json:"source"`
}

//Facebook Graph API page album edge
//https://developers.facebook.com/docs/graph-api/reference/page/albums
type AlbumsEdge struct {
	Data  []AlbumsEdgeAlbum  `json:"data"`
	Error GraphErrorResponse `json:"error"`
}

type AlbumsEdgeAlbum struct {
	CreatedTime string `json:"created_time"`
	Name        string `json:"name"`
	Id          string `json:"id"`
}

//Facebook Graph API standard error response
//https://developers.facebook.com/docs/graph-api/using-graph-api#errors
type GraphErrorResponse struct {
	Message          string `json:"message"`
	Type             string `json:"type"`
	Code             int    `json:"code"`
	Error_Subcode    int    `json:"error_subcode"`
	Error_User_Title string `json:"error_user_title"`
	Error_User_Msg   string `json:"error_user_msg"`
	FBTrace_Id       string `json:"fbtrace_id"`
}

/*
type Departure struct {
	RollCall    time.Time `json:"rollCall"`
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
	SeatCount   int       `json:"seatCount"`
	SeatType    string    `json:"seatType"`
	Canceled    bool      `json:"canceled"`
	PhotoSource string    `json:"photoSource"`
}
*/

type Terminal struct {
	Title    string `json:"title"`
	Id       string `json:"id"`
}

//Processed version of downloaded photo
type Slide struct {
	SaveType  SaveImageType
	Suffix    string
	Extension string
	Terminal  Terminal
	FBNodeId  string

	PlainText string
	HOCRText  string
}
