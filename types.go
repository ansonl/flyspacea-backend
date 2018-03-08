package main

import (
	"image"
	"log"
	"time"
)

/*
 * Facebook Graph API structs for retrieving page photos
 */

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

//Special keywords for a terminal in location_keywords.json
type TerminalKeywords struct {
	Title    string   `json:"title"`
	Keywords []string `json:"keywords"`
}

//Returned when searching all terminal keywords in plaintext
type TerminalKeywordsResult struct {
	Keyword  string
	Distance int
}

//Terminal representation
type Terminal struct {
	Title string `json:"title"`
	Id    string `json:"id"`
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

/*
 * Representation of data in an image
 */

//Destination representation
type Destination struct {
	TerminalTitle    string
	Spelling         string
	SpellingDistance int
	BBox             image.Rectangle

	//RollCall for the Destination
	LinkedRollCall RollCall
}

//RollCall representation
type RollCall struct {
	Time time.Time
	BBox image.Rectangle
}

//SeatsAvailable representation
type SeatsAvailable struct {
	Number int
	Letter string
	BBox   image.Rectangle
}

//"Grouping" of multiples Destinations for single RollCall/SeatsAvailable
type Grouping struct {
	Destinations []Destination
	BBox         image.Rectangle
}

//Update Grouping struct BBox to include all Destinations in grouping
func (g Grouping) updateBBox() {
	if len(g.Destinations) == 0 {
		log.Fatal("Grouping empty. Cannot updateBBox")
	}

	g.BBox = g.Destinations[0].BBox

	for _, d := range g.Destinations {
		if d.BBox.Min.X < g.BBox.Min.X {
			g.BBox.Min.X = d.BBox.Min.X
		}
		if d.BBox.Min.Y < g.BBox.Min.Y {
			g.BBox.Min.Y = d.BBox.Min.Y
		}
		if d.BBox.Max.X > g.BBox.Max.X {
			g.BBox.Max.X = d.BBox.Max.X
		}
		if d.BBox.Max.Y > g.BBox.Max.Y {
			g.BBox.Max.Y = d.BBox.Max.Y
		}
	}
}