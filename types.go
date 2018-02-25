package main

type PhotosEdge struct {
	Data []PhotosEdgePhoto `json:"data"`
}

type PhotosEdgePhoto struct {
	CreatedTime string `json:"created_time"`
	Name        string `json:"name"`
	Id          string `json:"id"`
}

type PhotoNode struct {
	Images []PhotoNodeImage `json:"images"`
}

type PhotoNodeImage struct {
	Source string `json:"source"`
}

type TextBlock struct {
	Text string `json:"text"`
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
	Hr72Id          string `json:"hr-72-id"`
	OffsetUp int    `json:"offsetUp"`
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
